package test

import (
	"context"
	"testing"

	"github.com/im-wmkong/gorm-query/db"
	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/im-wmkong/gorm-query/example/repository"
	"github.com/im-wmkong/gorm-query/example/service"
	"github.com/im-wmkong/gorm-query/query"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestCreateUsers 创建用户测试 (独立测试，不使用 setupTest 的默认数据，而是手动验证创建过程)
func TestCreateUsers(t *testing.T) {
	// 设置新的 DB
	gormDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	require.NoError(t, err)
	err = gormDB.AutoMigrate(&model.User{})
	require.NoError(t, err)

	connector := db.NewClient(gormDB)
	repo := repository.NewUserRepository(connector)
	ctx := context.Background()

	users := []struct {
		Name  string
		Email string
		Age   int
	}{
		{"Alice", "alice@example.com", 25},
		{"Bob", "bob@example.com", 30},
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

	// 验证创建
	q := query.New()
	count, err := repo.Count(ctx, q)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

// TestGetActiveUsers 测试获取活跃用户 (Service 层测试)
func TestGetActiveUsers(t *testing.T) {
	ctx, svc, _ := setupTest(t)

	// 用例 1: 无关键字
	activeUsers, err := svc.GetActiveUsers(ctx, 25, "")
	require.NoError(t, err)

	// 期望 Charlie(35), Bob(30), Alice(25)。"admin" 应被 NotIn 逻辑排除。
	// David(20) 被年龄排除。
	require.Len(t, activeUsers, 3)
	assert.Equal(t, "Charlie", activeUsers[0].UserName)
	assert.Equal(t, "Bob", activeUsers[1].UserName)
	assert.Equal(t, "Alice", activeUsers[2].UserName)

	// 用例 2: 带关键字 "ali" (匹配 Alice)
	activeUsersAli, err := svc.GetActiveUsers(ctx, 25, "ali")
	require.NoError(t, err)
	require.Len(t, activeUsersAli, 1)
	assert.Equal(t, "Alice", activeUsersAli[0].UserName)
}

// TestFindUserByName 测试通过用户名查找用户
func TestFindUserByName(t *testing.T) {
	ctx, _, repo := setupTest(t)

	q := query.New().Where(model.UserProps.UserName.Eq("Bob"))
	bob, err := repo.First(ctx, q)

	require.NoError(t, err)
	require.NotNil(t, bob)
	assert.Equal(t, "Bob", bob.UserName)
	assert.Equal(t, "bob@example.com", bob.Email)
}

// TestGetByID 测试 BaseService First (通过 ID 获取)
func TestGetByID(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// ID 1 应该是 Alice
	qID := query.New().Where(model.UserProps.ID.Eq(1))
	user1, err := repo.First(ctx, qID)
	require.NoError(t, err)
	require.NotNil(t, user1)
	assert.Equal(t, "Alice", user1.UserName)

	// 不存在的 ID
	qID999 := query.New().Where(model.UserProps.ID.Eq(999))
	user999, err := repo.First(ctx, qID999)

	require.Error(t, err)
	require.Nil(t, user999)
}

// TestDelete 测试删除
func TestDelete(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// 1. 根据 ID 删除 (通过 Where ID = ? 模拟)
	// 先找到 Alice 获取 ID
	alice, err := repo.First(ctx, query.New().Where(model.UserProps.UserName.Eq("Alice")))
	require.NoError(t, err)

	// 删除 Alice
	err = repo.Delete(ctx, query.New().Where(model.UserProps.ID.Eq(alice.ID)))
	require.NoError(t, err)

	// 验证 Alice 已删除
	_, err = repo.First(ctx, query.New().Where(model.UserProps.ID.Eq(alice.ID)))
	require.Error(t, err) // 应该记录未找到

	// 2. 批量删除 (删除所有剩余年龄 > 30 的用户)
	// 剩余: Bob(30), Charlie(35), David(20), admin(40)
	// Age > 30: Charlie(35), admin(40)
	err = repo.Delete(ctx, query.New().Where(model.UserProps.Age.Gt(30)))
	require.NoError(t, err)

	// 验证
	remaining, err := repo.Find(ctx, query.New())
	require.NoError(t, err)
	// 应该剩余: Bob(30), David(20) -> 2 个用户
	require.Len(t, remaining, 2)
	for _, u := range remaining {
		if u.UserName == "Charlie" || u.UserName == "admin" {
			t.Errorf("User %s should have been deleted", u.UserName)
		}
	}
}

// TestCreateUser 测试创建用户 (Service 层)
func TestCreateUser(t *testing.T) {
	ctx, svc, repo := setupTest(t)

	// 1. 创建新用户 - 应该成功
	newUser := &model.User{
		UserName: "Eve",
		Email:    "eve@example.com",
		Age:      22,
		Status:   1,
	}
	err := svc.CreateUser(ctx, newUser)
	require.NoError(t, err)

	// 验证创建
	q := query.New().Where(model.UserProps.UserName.Eq("Eve"))
	user, err := repo.First(ctx, q)
	require.NoError(t, err)
	assert.Equal(t, "Eve", user.UserName)

	// 2. 创建重复用户 - 应该失败
	duplicateUser := &model.User{
		UserName: "Eve",
		Email:    "eve@example.com",
		Age:      22,
		Status:   1,
	}
	err = svc.CreateUser(ctx, duplicateUser)
	require.Error(t, err)
	assert.Equal(t, service.ErrUserAlreadyExists, err)
}

// TestUpdate 测试 Update
func TestUpdate(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// 更新 Alice 的 Age 为 26
	q := query.New().Where(model.UserProps.UserName.Eq("Alice"))
	err := repo.Update(ctx, q, "age", 26)
	require.NoError(t, err)

	// 验证
	alice, err := repo.First(ctx, q)
	require.NoError(t, err)
	assert.Equal(t, 26, alice.Age)
}

// TestUpdates 测试 Updates
func TestUpdates(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// 更新 Bob 的 Age 和 Status
	q := query.New().Where(model.UserProps.UserName.Eq("Bob"))
	updates := map[string]interface{}{
		"age":    31,
		"status": 2,
	}
	err := repo.Updates(ctx, q, updates)
	require.NoError(t, err)

	// 验证
	bob, err := repo.First(ctx, q)
	require.NoError(t, err)
	assert.Equal(t, 31, bob.Age)
	assert.Equal(t, 2, bob.Status)
}

// TestSave 测试 Save
func TestSave(t *testing.T) {
	ctx, _, repo := setupTest(t)

	// 获取 Charlie
	q := query.New().Where(model.UserProps.UserName.Eq("Charlie"))
	charlie, err := repo.First(ctx, q)
	require.NoError(t, err)

	// 修改
	charlie.Age = 36
	charlie.Status = 0

	// 保存
	err = repo.Save(ctx, charlie)
	require.NoError(t, err)

	// 验证
	charlieNew, err := repo.First(ctx, q)
	require.NoError(t, err)
	assert.Equal(t, 36, charlieNew.Age)
	assert.Equal(t, 0, charlieNew.Status)
}
