// Package db 提供了基于 context.Context 的 GORM 数据库连接与事务管理能力。
//
// 它定义了 Connector 和 TransactionManager 接口，使得业务的 Service 层可以
// 在不感知底层 *gorm.DB 实例的情况下开启和传递事务，从而实现 Repo 层与 Service 层的完美解耦。
package db

import (
	"context"

	"gorm.io/gorm"
)

type txKeyType struct{}

var txKey = txKeyType{}

// Connector 定义了数据库连接的接口
type Connector interface {
	DB(ctx context.Context) *gorm.DB
}

// TransactionManager 定义了数据库事务管理的接口
type TransactionManager interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

var _ Connector = (*Client)(nil)

var _ TransactionManager = (*Client)(nil)

type Client struct {
	db *gorm.DB
}

// NewClient 创建一个新的 Client 实例
func NewClient(db *gorm.DB) *Client {
	return &Client{db: db}
}

// DB 返回当前上下文的 GORM DB 实例
func (c *Client) DB(ctx context.Context) *gorm.DB {
	v := ctx.Value(txKey)
	if v != nil {
		if tx, ok := v.(*gorm.DB); ok {
			return tx.WithContext(ctx)
		}
	}
	return c.db.WithContext(ctx)
}

// Transaction 在当前上下文开启一个新的数据库事务
func (c *Client) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return c.DB(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		return fn(txCtx)
	})
}
