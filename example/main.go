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
	// 1) Initialize a GORM database connection (SQLite in this example).
	gormDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Auto-migrate schema (demo only).
	_ = gormDB.AutoMigrate(&model.User{}, &model.Profile{})

	// 2) Initialize the core component: db.Client.
	// It implements both db.DBProvider (for repos) and db.Transactor (for services).
	dbClient := db.NewClient(gormDB)

	// 3) Dependency injection (simulating DI tools like Wire).
	userRepo := repository.NewUserRepository(dbClient)
	userService := service.NewUserService(userRepo, dbClient)

	// 4) Execute business logic.
	ctx := context.Background()

	// Demo 1: execute transactional business logic via the Service.
	user := &model.User{
		UserName: "Alice",
		Age:      18,
		Email:    "alice@example.com",
	}
	err = userService.CreateUser(ctx, user)
	if err != nil {
		log.Printf("CreateUser failed: %v", err)
	}

	// Demo 2: query via the Service.
	users, err := userService.GetActiveUsers(ctx, 18, "Alice")
	if err != nil {
		log.Printf("GetActiveUsers failed: %v", err)
	} else {
		log.Printf("Found active users: %d", len(users))
	}
}
