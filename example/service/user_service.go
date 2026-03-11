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

// UserService 定义了继承自通用 Service 接口的服务接口
type UserService interface {
	CreateUser(ctx context.Context, user *model.User) error
	GetActiveUsers(ctx context.Context, minAge int, keyword string) ([]*model.User, error)
}

// userService 实现继承自 BaseService
type userService struct {
	repo repository.UserRepository // 如果需要自定义方法，保留特定的 repo 引用
	tm   db.TransactionManager
}

// NewUserService 创建一个新的 user service
func NewUserService(repo repository.UserRepository, tm db.TransactionManager) UserService {
	return &userService{
		repo: repo,
		tm:   tm,
	}
}

// CreateUser 创建一个新用户并进行验证
func (s *userService) CreateUser(ctx context.Context, user *model.User) error {
	return s.tm.Transaction(ctx, func(ctx context.Context) error {
		// 检查用户邮箱是否已存在
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
		// TODO: 其他业务逻辑
		return nil
	})
}

// GetActiveUsers 演示了使用类型安全列进行复杂的查询构造
func (s *userService) GetActiveUsers(ctx context.Context, minAge int, keyword string) ([]*model.User, error) {
	// 构造查询:
	// 1. Status = 1
	// 2. Age >= minAge
	// 3. UserName NOT IN ["admin", "root"] (演示 NotIn)
	// 4. Email LIKE %keyword% (如果有 keyword)
	// 5. 按 CreatedAt DESC 排序

	q := query.New().Where(
		model.UserProps.Status.Eq(1),
		model.UserProps.Age.Gte(minAge),
		model.UserProps.UserName.NotIn([]string{"admin", "root"}), // 演示 NotIn
	)

	if keyword != "" {
		q.Where(model.UserProps.Email.Like("%" + keyword + "%")) // 演示 Like
	}

	q.Order(model.UserProps.CreatedAt.Desc())

	return s.repo.Find(ctx, q)
}
