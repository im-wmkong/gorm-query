package main

import (
	"context"
	"log"

	"github.com/im-wmkong/gorm-query/db"
	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/im-wmkong/gorm-query/example/repository"
	"github.com/im-wmkong/gorm-query/example/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// 1. 初始化 GORM 数据库连接 (这里以 SQLite 为例)
	gormDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// 自动迁移表结构 (仅作演示)
	_ = gormDB.AutoMigrate(&model.User{})

	// 2. 初始化核心组件：db.Client
	// 它同时实现了 repo 需要的 db.DBProvider 和 service 需要的 db.Transactor
	dbClient := db.NewClient(gormDB)

	// 3. 依赖注入 (模拟 Wire 等 DI 工具的过程)
	userRepo := repository.NewUserRepository(dbClient)
	userService := service.NewUserService(userRepo, dbClient)

	// 4. 执行业务逻辑
	ctx := context.Background()

	// 演示 1：通过 Service 执行带事务的业务
	err = userService.CreateUser(ctx, &model.User{
		UserName: "Alice",
		Age:      18,
		Email:    "alice@example.com",
	})
	if err != nil {
		log.Printf("CreateUser failed: %v", err)
	}

	// 演示 2：通过 Service 进行查询
	users, err := userService.GetActiveUsers(ctx, 18, "Alice")
	if err != nil {
		log.Printf("GetActiveUsers failed: %v", err)
	} else {
		log.Printf("Found active users: %d", len(users))
	}
}
