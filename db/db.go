// Package db 提供了基于 context.Context 的 GORM 数据库连接与事务管理能力。
//
// 它定义了 DBProvider 和 Transactor 接口，使得业务的 Service 层可以
// 在不感知底层 *gorm.DB 实例的情况下开启和传递事务，从而实现 Repo 层与 Service 层的完美解耦。
package db

import (
	"context"

	"gorm.io/gorm"
)

type txKeyType struct{}

var txKey = txKeyType{}

// DBProvider 定义了数据库连接的接口
type DBProvider interface {
	DB(ctx context.Context) *gorm.DB
}

// Transactor 定义了数据库事务管理的接口
type Transactor interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

var _ DBProvider = (*Client)(nil)

var _ Transactor = (*Client)(nil)

type Client struct {
	db *gorm.DB
}

// NewClient 创建一个新的 Client 实例。db 不能为 nil，否则 panic。
func NewClient(db *gorm.DB) *Client {
	if db == nil {
		panic("db: NewClient called with nil *gorm.DB")
	}
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
