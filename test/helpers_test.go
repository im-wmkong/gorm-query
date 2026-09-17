package test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/im-wmkong/gorm-query/db"
	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/im-wmkong/gorm-query/example/model/schema"
	"github.com/im-wmkong/gorm-query/query"
	"github.com/im-wmkong/gorm-query/repo"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Each call owns a database, including multiple calls from the same test.
func openDB(t *testing.T) *gorm.DB {
	t.Helper()
	d, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := d.DB()
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, sqlDB.Close()) })
	require.NoError(t, d.AutoMigrate(&model.User{}, &model.Profile{}))
	return d
}

func seed(t *testing.T, d *gorm.DB) {
	t.Helper()
	users := []model.User{
		{UserName: "Alice", Email: "alice@example.com", Age: 10, Profile: &model.Profile{Bio: "SF"}},
		{UserName: "Bob", Email: "bob@example.com", Age: 20, Profile: &model.Profile{Bio: "NY"}},
		{UserName: "Carol", Email: "carol@example.com", Age: 30},
		{UserName: "Dave", Email: "dave@example.com", Age: 40},
	}
	require.NoError(t, d.Create(&users).Error)
}

func names(t *testing.T, d *gorm.DB, qb *query.Builder[model.User]) []string {
	t.Helper()
	r := repo.New[model.User](db.NewClient(d))
	users, err := r.Find(context.Background(), qb.Order(schema.User.ID.Asc()))
	require.NoError(t, err)
	out := make([]string, len(users))
	for i, u := range users {
		out[i] = u.UserName
	}
	return out
}
