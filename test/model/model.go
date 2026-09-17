// Package model contains consumer models for generation contract tests.
package model

import (
	"database/sql"
	"time"
)

type Label string
type Amount int64

type TypedRecord struct {
	ID       uint
	Label    Label
	Amount   Amount
	Active   bool
	At       time.Time
	Optional *int
	Note     sql.NullString
}

type Record struct {
	ID   uint
	Name string
}

type Explicit struct {
	ID   uint
	Name string
}

func (*Explicit) TableName() string { return "custom_records" }

type Address struct {
	City string
}

type Shipment struct {
	ID       uint
	Billing  Address `gorm:"embedded;embeddedPrefix:billing_"`
	Shipping Address `gorm:"embedded;embeddedPrefix:shipping_"`
}

type ReorderedShipment struct {
	Shipping Address `gorm:"embedded;embeddedPrefix:shipping_"`
	Billing  Address `gorm:"embedded;embeddedPrefix:billing_"`
	ID       uint
}

type CityCollision struct {
	ID      uint
	City    string
	Billing Address `gorm:"embedded;embeddedPrefix:billing_"`
}

type CityOverride struct {
	Address Address `gorm:"embedded"`
	ID      uint
	City    string
}

type AmbiguousCity struct {
	ID          uint
	BillingCity string  `gorm:"column:own_city"`
	Billing     Address `gorm:"embedded;embeddedPrefix:billing_"`
	Shipping    Address `gorm:"embedded;embeddedPrefix:shipping_"`
}

type Reserved struct {
	ID    uint
	Value int `gorm:"column:select"`
}

func (*Reserved) TableName() string { return "order" }
