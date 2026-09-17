package db

import (
	"context"
	"errors"
	"fmt"
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
		if err := txDB.Create(&user{UserName: "ghost", Email: "g@t.com", Age: 1}).Error; err != nil {
			return err
		}
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
	innerFailure := errors.New("inner fail")

	err := client.Transaction(ctx, func(outerCtx context.Context) error {
		if err := client.DB(outerCtx).Create(&user{UserName: "outer", Email: "o@t.com", Age: 1}).Error; err != nil {
			return err
		}

		// Inner transaction fails and rolls back.
		innerErr := client.Transaction(outerCtx, func(innerCtx context.Context) error {
			if err := client.DB(innerCtx).Create(&user{UserName: "inner", Email: "i@t.com", Age: 1}).Error; err != nil {
				return err
			}
			return innerFailure
		})
		require.ErrorIs(t, innerErr, innerFailure)

		return nil // Outer transaction commits.
	})
	require.NoError(t, err)

	// Outer data is visible; inner data has been rolled back.
	var count int64
	require.NoError(t, client.DB(ctx).Model(&user{}).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}
