package service

import (
	"context"
	"errors"

	"github.com/im-wmkong/gorm-query/db"
	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/im-wmkong/gorm-query/example/repository"
	"github.com/im-wmkong/gorm-query/query"
)

var ErrUserAlreadyExists = errors.New("user already exists")

// UserService defines the service interface for User.
type UserService interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetActiveUsers(ctx context.Context, minAge int, keyword string) ([]*model.User, error)
}

// userService implements UserService.
type userService struct {
	repo       repository.UserRepository // Keep a concrete repo reference if you need custom methods.
	transactor db.Transactor
}

// NewUserService creates a new user service.
func NewUserService(repo repository.UserRepository, tm db.Transactor) UserService {
	return &userService{
		repo:       repo,
		transactor: tm,
	}
}

// CreateUser creates a new user with basic validation.
func (s *userService) CreateUser(ctx context.Context, user *model.User) error {
	return s.transactor.Transaction(ctx, func(ctx context.Context) error {
		// Check if the email already exists.
		q := query.New().Where(model.UserProps.Email.Eq(user.Email))
		count, err := s.repo.Count(ctx, q)
		if err != nil {
			return err
		}
		if count > 0 {
			return ErrUserAlreadyExists
		}
		if err = s.repo.Create(ctx, user); err != nil {
			return err
		}
		// TODO: additional business logic.
		return nil
	})
}

// GetActiveUsers demonstrates building a complex query with type-safe columns.
func (s *userService) GetActiveUsers(ctx context.Context, minAge int, keyword string) ([]*model.User, error) {
	// Build query:
	// 1) Status = 1
	// 2) Age >= minAge
	// 3) UserName NOT IN ["admin", "root"] (NotIn)
	// 4) Email LIKE %keyword% (if keyword is provided)
	// 5) ORDER BY CreatedAt DESC

	q := query.New().Where(
		model.UserProps.Status.Eq(1),
		model.UserProps.Age.Gte(minAge),
		model.UserProps.UserName.NotIn([]string{"admin", "root"}), // NotIn demo
	)

	if keyword != "" {
		q = q.Where(model.UserProps.Email.Like("%" + keyword + "%")) // Like demo
	}

	q = q.Order(model.UserProps.CreatedAt.Desc())

	return s.repo.Find(ctx, q)
}
