package test

import (
	"path/filepath"
	"testing"

	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/im-wmkong/gorm-query/internal/gormx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestInternalGormxBuildNested(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "gormx.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}))
	require.NoError(t, db.Create([]model.User{
		{UserName: "Alice", Age: 19, Status: 1, Email: "alice@example.com"},
		{UserName: "Bob", Age: 30, Status: 1, Email: "bob@example.com"},
		{UserName: "Charlie", Age: 35, Status: 1, Email: "charlie@example.com"},
	}).Error)

	base := db.Model(&model.User{})
	nested := gormx.BuildNested(base, []func(*gorm.DB) *gorm.DB{
		func(tx *gorm.DB) *gorm.DB { return tx.Where("age >= ?", 30) },
		func(tx *gorm.DB) *gorm.DB { return tx.Where("user_name = ?", "Bob") },
	})

	_, polluted := base.Statement.Clauses["WHERE"]
	assert.False(t, polluted)

	var users []model.User
	require.NoError(t, db.Model(&model.User{}).Where(nested).Order("age asc").Find(&users).Error)
	require.Len(t, users, 1)
	assert.Equal(t, "Bob", users[0].UserName)
	assert.Equal(t, 30, users[0].Age)
}
