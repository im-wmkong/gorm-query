package gormx

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type user struct {
	gorm.Model
	UserName string `gorm:"column:user_name;size:255;not null"`
	Email    string `gorm:"column:email;size:255;unique"`
	Age      int    `gorm:"column:age"`
	Status   int    `gorm:"column:status;default:1"`
}

func TestGormxBuildNested(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&user{}))
	require.NoError(t, db.Create([]user{
		{UserName: "Alice", Age: 19, Status: 1, Email: "alice@example.com"},
		{UserName: "Bob", Age: 30, Status: 1, Email: "bob@example.com"},
		{UserName: "Charlie", Age: 35, Status: 1, Email: "charlie@example.com"},
	}).Error)

	base := db.Model(&user{})
	nested := BuildNested(base, []func(*gorm.DB) *gorm.DB{
		func(tx *gorm.DB) *gorm.DB { return tx.Where("age >= ?", 30) },
		func(tx *gorm.DB) *gorm.DB { return tx.Where("user_name = ?", "Bob") },
	})

	_, polluted := base.Statement.Clauses["WHERE"]
	assert.False(t, polluted)

	var users []user
	require.NoError(t, db.Model(&user{}).Where(nested).Order("age asc").Find(&users).Error)
	require.Len(t, users, 1)
	assert.Equal(t, "Bob", users[0].UserName)
	assert.Equal(t, 30, users[0].Age)
}

func TestGormxBuildNested_EmptyConds(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	require.NoError(t, err)

	base := db.Model(&user{})
	nested := BuildNested(base, []func(*gorm.DB) *gorm.DB{})
	require.NotNil(t, nested)

	// Empty conds must yield a fresh NewDB session without any clauses inherited from base.
	_, polluted := nested.Statement.Clauses["WHERE"]
	assert.False(t, polluted)
}

func TestGormxBuildNested_NilConds(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	require.NoError(t, err)

	nested := BuildNested[func(*gorm.DB) *gorm.DB](db.Model(&user{}), nil)
	require.NotNil(t, nested)
}
