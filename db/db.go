// Package db provides context-aware GORM database access and transaction management.
//
// It defines the DBProvider and Transactor interfaces so the Service layer can
// start and propagate transactions without directly depending on *gorm.DB,
// keeping the Repo layer decoupled from the Service layer.
package db

import (
	"context"

	"gorm.io/gorm"
)

type txKeyType struct{}

var txKey = txKeyType{}

// DBProvider provides a *gorm.DB bound to the given context.
type DBProvider interface {
	DB(ctx context.Context) *gorm.DB
}

// Transactor provides context-aware transaction management.
type Transactor interface {
	Transaction(ctx context.Context, fn func(ctx context.Context) error) error
}

var _ DBProvider = (*Client)(nil)

var _ Transactor = (*Client)(nil)

type Client struct {
	db *gorm.DB
}

// NewClient creates a new Client instance. db must not be nil, otherwise it panics.
//
// Example:
//
//	gormDB, _ := gorm.Open(...)
//	client := db.NewClient(gormDB)
//	_ = client
func NewClient(db *gorm.DB) *Client {
	if db == nil {
		panic("db: NewClient called with nil *gorm.DB")
	}
	return &Client{db: db}
}

// DB returns the GORM DB instance for the given context.
//
// Example:
//
//	ctx := context.Background()
//	session := client.DB(ctx)
//	_ = session
func (c *Client) DB(ctx context.Context) *gorm.DB {
	v := ctx.Value(txKey)
	if v != nil {
		if tx, ok := v.(*gorm.DB); ok {
			return tx.WithContext(ctx)
		}
	}
	return c.db.WithContext(ctx)
}

// Transaction starts a new transaction and passes a transaction-aware context to fn.
//
// Example:
//
//	err := client.Transaction(ctx, func(txCtx context.Context) error {
//	    // Any repo/db call using txCtx will use the same transaction.
//	    return nil
//	})
//	_ = err
func (c *Client) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return c.DB(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey, tx)
		return fn(txCtx)
	})
}
