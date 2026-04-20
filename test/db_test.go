package test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/im-wmkong/gorm-query/db"
	"github.com/im-wmkong/gorm-query/example/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:dbtest_%s?mode=memory&cache=shared", t.Name())
	gormDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, gormDB.AutoMigrate(&model.User{}))
	return gormDB
}

// TestDB_WithoutTransaction 验证无事务时 DB(ctx) 返回普通连接
func TestDB_WithoutTransaction(t *testing.T) {
	gormDB := openTestDB(t)
	client := db.NewClient(gormDB)
	ctx := context.Background()

	session := client.DB(ctx)
	require.NotNil(t, session)

	// 正常 CRUD 可用
	err := session.Create(&model.User{UserName: "test", Email: "t@t.com", Age: 1}).Error
	require.NoError(t, err)

	var count int64
	require.NoError(t, session.Model(&model.User{}).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

// TestDB_InsideTransaction 验证事务上下文中 DB(ctx) 返回事务连接
func TestDB_InsideTransaction(t *testing.T) {
	gormDB := openTestDB(t)
	client := db.NewClient(gormDB)
	ctx := context.Background()

	err := client.Transaction(ctx, func(txCtx context.Context) error {
		txDB := client.DB(txCtx)
		require.NotNil(t, txDB)

		return txDB.Create(&model.User{UserName: "inTx", Email: "tx@t.com", Age: 1}).Error
	})
	require.NoError(t, err)

	// 事务提交后数据可见
	var count int64
	require.NoError(t, client.DB(ctx).Model(&model.User{}).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

// TestDB_TransactionRollback 验证 fn 返回 error 时事务回滚
func TestDB_TransactionRollback(t *testing.T) {
	gormDB := openTestDB(t)
	client := db.NewClient(gormDB)
	ctx := context.Background()

	expectedErr := errors.New("rollback me")
	err := client.Transaction(ctx, func(txCtx context.Context) error {
		txDB := client.DB(txCtx)
		_ = txDB.Create(&model.User{UserName: "ghost", Email: "g@t.com", Age: 1}).Error
		return expectedErr
	})
	require.ErrorIs(t, err, expectedErr)

	// 回滚后数据不可见
	var count int64
	require.NoError(t, client.DB(ctx).Model(&model.User{}).Count(&count).Error)
	assert.Equal(t, int64(0), count)
}

// TestDB_NestedTransaction 验证嵌套事务（GORM SavePoint）
func TestDB_NestedTransaction(t *testing.T) {
	gormDB := openTestDB(t)
	client := db.NewClient(gormDB)
	ctx := context.Background()

	err := client.Transaction(ctx, func(outerCtx context.Context) error {
		_ = client.DB(outerCtx).Create(&model.User{UserName: "outer", Email: "o@t.com", Age: 1}).Error

		// 内层事务失败并回滚
		innerErr := client.Transaction(outerCtx, func(innerCtx context.Context) error {
			_ = client.DB(innerCtx).Create(&model.User{UserName: "inner", Email: "i@t.com", Age: 1}).Error
			return errors.New("inner fail")
		})
		require.Error(t, innerErr)

		return nil // 外层提交
	})
	require.NoError(t, err)

	// 外层数据可见，内层数据已回滚
	var count int64
	require.NoError(t, client.DB(ctx).Model(&model.User{}).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

// TestDB_ContextValueNotGormDB 验证 context 中 txKey 值类型不匹配时的兜底
func TestDB_ContextValueNotGormDB(t *testing.T) {
	gormDB := openTestDB(t)
	client := db.NewClient(gormDB)

	// 无法直接设置 txKey（未导出），但我们可以通过验证正常 context 不会误触发来确认安全性
	// 在正常使用中，只有 Transaction 方法会设置 txKey，所以此测试确认普通 context 不会出问题
	ctx := context.WithValue(context.Background(), struct{ name string }{"unrelated"}, "value")
	session := client.DB(ctx)
	require.NotNil(t, session)

	// 应该返回普通连接，不会 panic
	var count int64
	require.NoError(t, session.Model(&model.User{}).Count(&count).Error)
}
