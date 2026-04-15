package repository

import (
	"github.com/im-wmkong/gorm-query/db"
	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/im-wmkong/gorm-query/repo"
)

// UserRepository 接口继承了通用的 Repository 接口
// 你可以在这里添加针对 User 的特定方法
type UserRepository interface {
	repo.Repository[model.User]
}

// userRepository 实现继承自 BaseRepository
type userRepository struct {
	*repo.BaseRepository[model.User]
}

// NewUserRepository 创建一个新的 user repository
func NewUserRepository(dbProvider db.DBProvider) UserRepository {
	return &userRepository{
		BaseRepository: repo.New[model.User](dbProvider),
	}
}
