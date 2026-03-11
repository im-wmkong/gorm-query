package test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/im-wmkong/gorm-query/db"
	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/im-wmkong/gorm-query/example/repository"
	"github.com/im-wmkong/gorm-query/example/service"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTest 初始化测试环境，返回 context, service 和 repository
// 使用内存数据库，确保每次测试都是独立的
func setupTest(t *testing.T) (context.Context, service.UserService, repository.UserRepository) {
	// Setup DB (In-memory)
	// 使用 file::memory:?cache=shared 模式或者随机文件名确保隔离
	// 修正：使用随机数据库名称避免缓存冲突
	dbName := fmt.Sprintf("file:memdb_%s?mode=memory&cache=shared", t.Name())
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,  // Slow SQL threshold
			LogLevel:                  logger.Error, // Log level
			IgnoreRecordNotFoundError: true,         // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,        // Disable color
		},
	)
	gormDB, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{
		Logger: newLogger, // 减少日志噪音
	})
	require.NoError(t, err, "failed to connect database")

	// Migrate
	err = gormDB.AutoMigrate(&model.User{})
	require.NoError(t, err)

	// Initialize components
	client := db.NewClient(gormDB)
	repo := repository.NewUserRepository(client)
	svc := service.NewUserService(repo, client)

	ctx := context.Background()

	// 填充标准数据
	seedUsers(t, ctx, repo)

	return ctx, svc, repo
}

func seedUsers(t *testing.T, ctx context.Context, repo repository.UserRepository) {
	users := []struct {
		Name  string
		Email string
		Age   int
	}{
		{"Alice", "alice@example.com", 25},
		{"Bob", "bob@example.com", 30},
		{"Charlie", "charlie@example.com", 35},
		{"David", "david@example.com", 20},
		{"admin", "admin@example.com", 40}, // 在某些测试中应被排除 (小写 "admin")
	}

	for _, u := range users {
		err := repo.Create(ctx, &model.User{
			UserName: u.Name,
			Email:    u.Email,
			Age:      u.Age,
			Status:   1,
		})
		require.NoError(t, err, "failed to create user %s", u.Name)
	}
}
