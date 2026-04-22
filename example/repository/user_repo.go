package repository

import (
	"github.com/im-wmkong/gorm-query/db"
	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/im-wmkong/gorm-query/repo"
)

// UserRepository extends the generic Repository interface.
// You can add user-specific methods here.
type UserRepository interface {
	repo.Repository[model.User]
}

// userRepository embeds BaseRepository.
type userRepository struct {
	*repo.BaseRepository[model.User]
}

// NewUserRepository creates a new user repository.
func NewUserRepository(dbProvider db.DBProvider) UserRepository {
	return &userRepository{
		BaseRepository: repo.New[model.User](dbProvider),
	}
}
