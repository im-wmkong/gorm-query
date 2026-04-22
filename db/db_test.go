package db

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type user struct {
	gorm.Model
	UserName string `gorm:"column:user_name;size:255;not null"`
	Email    string `gorm:"column:email;size:255;unique"`
	Age      int    `gorm:"column:age"`
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:dbtest_%s?mode=memory&cache=shared", t.Name())
	gormDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, gormDB.AutoMigrate(&user{}))
	return gormDB
}

func callDBWithNilContext(t *testing.T, client *Client) *gorm.DB {
	t.Helper()
	results := reflect.ValueOf(client).MethodByName("DB").Call([]reflect.Value{
		reflect.Zero(reflect.TypeOf((*context.Context)(nil)).Elem()),
	})
	return results[0].Interface().(*gorm.DB)
}

func callTransactionWithNilContext(t *testing.T, client *Client, fn func(context.Context) error) error {
	t.Helper()
	results := reflect.ValueOf(client).MethodByName("Transaction").Call([]reflect.Value{
		reflect.Zero(reflect.TypeOf((*context.Context)(nil)).Elem()),
		reflect.ValueOf(fn),
	})
	if results[0].IsNil() {
		return nil
	}
	return results[0].Interface().(error)
}

// TestDB_WithoutTransaction verifies that DB(ctx) returns a normal connection without a transaction.
func TestDB_WithoutTransaction(t *testing.T) {
	gormDB := openTestDB(t)
	client := NewClient(gormDB)
	ctx := context.Background()

	session := client.DB(ctx)
	require.NotNil(t, session)

	// Basic CRUD should work.
	err := session.Create(&user{UserName: "test", Email: "t@t.com", Age: 1}).Error
	require.NoError(t, err)

	var count int64
	require.NoError(t, session.Model(&user{}).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

// TestDB_InsideTransaction verifies that DB(ctx) returns a transactional connection inside Transaction.
func TestDB_InsideTransaction(t *testing.T) {
	gormDB := openTestDB(t)
	client := NewClient(gormDB)
	ctx := context.Background()

	err := client.Transaction(ctx, func(txCtx context.Context) error {
		txDB := client.DB(txCtx)
		require.NotNil(t, txDB)

		return txDB.Create(&user{UserName: "inTx", Email: "tx@t.com", Age: 1}).Error
	})
	require.NoError(t, err)

	// Data should be visible after commit.
	var count int64
	require.NoError(t, client.DB(ctx).Model(&user{}).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

// TestDB_TransactionRollback verifies that returning an error from fn triggers rollback.
func TestDB_TransactionRollback(t *testing.T) {
	gormDB := openTestDB(t)
	client := NewClient(gormDB)
	ctx := context.Background()

	expectedErr := errors.New("rollback me")
	err := client.Transaction(ctx, func(txCtx context.Context) error {
		txDB := client.DB(txCtx)
		_ = txDB.Create(&user{UserName: "ghost", Email: "g@t.com", Age: 1}).Error
		return expectedErr
	})
	require.ErrorIs(t, err, expectedErr)

	// Data should not be visible after rollback.
	var count int64
	require.NoError(t, client.DB(ctx).Model(&user{}).Count(&count).Error)
	assert.Equal(t, int64(0), count)
}

// TestDB_NestedTransaction verifies nested transactions (GORM savepoints).
func TestDB_NestedTransaction(t *testing.T) {
	gormDB := openTestDB(t)
	client := NewClient(gormDB)
	ctx := context.Background()

	err := client.Transaction(ctx, func(outerCtx context.Context) error {
		_ = client.DB(outerCtx).Create(&user{UserName: "outer", Email: "o@t.com", Age: 1}).Error

		// Inner transaction fails and rolls back.
		innerErr := client.Transaction(outerCtx, func(innerCtx context.Context) error {
			_ = client.DB(innerCtx).Create(&user{UserName: "inner", Email: "i@t.com", Age: 1}).Error
			return errors.New("inner fail")
		})
		require.Error(t, innerErr)

		return nil // Outer transaction commits.
	})
	require.NoError(t, err)

	// Outer data is visible; inner data has been rolled back.
	var count int64
	require.NoError(t, client.DB(ctx).Model(&user{}).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

// TestDB_ContextValueNotGormDB verifies the fallback when ctx contains a txKey value of a wrong type.
func TestDB_ContextValueNotGormDB(t *testing.T) {
	gormDB := openTestDB(t)
	client := NewClient(gormDB)

	// We cannot set txKey directly (unexported), but we can validate that a normal context
	// will not accidentally trigger transactional behavior.
	// In real usage, only Transaction sets txKey.
	ctx := context.WithValue(context.Background(), struct{ name string }{"unrelated"}, "value")
	session := client.DB(ctx)
	require.NotNil(t, session)

	// Should return a normal connection and must not panic.
	var count int64
	require.NoError(t, session.Model(&user{}).Count(&count).Error)
}

func TestDB_NilContextFallsBackToBackground(t *testing.T) {
	gormDB := openTestDB(t)
	client := NewClient(gormDB)

	session := callDBWithNilContext(t, client)
	require.NotNil(t, session)
	require.NoError(t, session.Create(&user{UserName: "nilctx", Email: "nil@t.com", Age: 1}).Error)

	var count int64
	require.NoError(t, session.Model(&user{}).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestDB_TransactionWithNilContext(t *testing.T) {
	gormDB := openTestDB(t)
	client := NewClient(gormDB)

	err := callTransactionWithNilContext(t, client, func(txCtx context.Context) error {
		require.NotNil(t, txCtx)
		return client.DB(txCtx).Create(&user{UserName: "txnil", Email: "txnil@t.com", Age: 1}).Error
	})
	require.NoError(t, err)

	var count int64
	require.NoError(t, client.DB(context.Background()).Model(&user{}).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestNewClient_PanicsOnNilDB(t *testing.T) {
	require.PanicsWithValue(t, "db: NewClient called with nil *gorm.DB", func() {
		_ = NewClient(nil)
	})
}
