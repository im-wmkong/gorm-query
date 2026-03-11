package db

import (
	"context"

	"gorm.io/gorm"
)

type txKeyType struct{}

var txKey = txKeyType{}

type Connector interface {
	DB(ctx context.Context) *gorm.DB
}

type TransactionManager interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

var _ Connector = (*Client)(nil)

var _ TransactionManager = (*Client)(nil)

type Client struct {
	db *gorm.DB
}

func NewClient(db *gorm.DB) *Client {
	return &Client{db: db}
}

func (c *Client) DB(ctx context.Context) *gorm.DB {
	v := ctx.Value(txKey)
	if v != nil {
		if tx, ok := v.(*gorm.DB); ok {
			return tx.WithContext(ctx)
		}
	}
	return c.db.WithContext(ctx)
}

func (c *Client) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return c.DB(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		return fn(txCtx)
	})
}
