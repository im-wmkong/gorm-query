package service

import (
	"context"
	"errors"
	"testing"

	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/im-wmkong/gorm-query/query"
	"gorm.io/gorm"

	"github.com/stretchr/testify/require"
)

type stubTransactor struct{}

func (stubTransactor) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type stubUserRepo struct {
	count int64
	// return values
	countErr  error
	createErr error

	countCalled  int
	createCalled int
}

func (r *stubUserRepo) DB(context.Context) *gorm.DB { panic("not used") }
func (r *stubUserRepo) Save(context.Context, *model.User) error {
	panic("not used")
}
func (r *stubUserRepo) Create(_ context.Context, _ *model.User) error {
	r.createCalled++
	return r.createErr
}
func (r *stubUserRepo) CreateInBatches(context.Context, []*model.User, int) (int64, error) {
	panic("not used")
}
func (r *stubUserRepo) Update(context.Context, *query.Builder[model.User], ...query.Assignment) (int64, error) {
	panic("not used")
}
func (r *stubUserRepo) Delete(context.Context, *query.Builder[model.User]) (int64, error) {
	panic("not used")
}
func (r *stubUserRepo) Find(context.Context, *query.Builder[model.User]) ([]*model.User, error) {
	panic("not used")
}
func (r *stubUserRepo) First(context.Context, *query.Builder[model.User]) (*model.User, error) {
	panic("not used")
}
func (r *stubUserRepo) Take(context.Context, *query.Builder[model.User]) (*model.User, error) {
	panic("not used")
}
func (r *stubUserRepo) Last(context.Context, *query.Builder[model.User]) (*model.User, error) {
	panic("not used")
}
func (r *stubUserRepo) Count(context.Context, *query.Builder[model.User]) (int64, error) {
	r.countCalled++
	return r.count, r.countErr
}
func (r *stubUserRepo) Exists(context.Context, *query.Builder[model.User]) (bool, error) {
	panic("not used")
}
func (r *stubUserRepo) Pluck(context.Context, *query.Builder[model.User], query.SQLFragment, any) error {
	panic("not used")
}

func TestUserService_CreateUser_CoveragePaths(t *testing.T) {
	ctx := context.Background()

	t.Run("count error", func(t *testing.T) {
		repo := &stubUserRepo{countErr: errors.New("count failed")}
		svc := NewUserService(repo, stubTransactor{})

		err := svc.CreateUser(ctx, &model.User{Email: "a@example.com"})
		require.Error(t, err)
		require.Equal(t, 1, repo.countCalled)
		require.Equal(t, 0, repo.createCalled)
	})

	t.Run("already exists", func(t *testing.T) {
		repo := &stubUserRepo{count: 1}
		svc := NewUserService(repo, stubTransactor{})

		err := svc.CreateUser(ctx, &model.User{Email: "b@example.com"})
		require.ErrorIs(t, err, ErrUserAlreadyExists)
		require.Equal(t, 1, repo.countCalled)
		require.Equal(t, 0, repo.createCalled)
	})

	t.Run("create error", func(t *testing.T) {
		repo := &stubUserRepo{createErr: errors.New("create failed")}
		svc := NewUserService(repo, stubTransactor{})

		err := svc.CreateUser(ctx, &model.User{Email: "c@example.com"})
		require.Error(t, err)
		require.Equal(t, 1, repo.countCalled)
		require.Equal(t, 1, repo.createCalled)
	})

	t.Run("success", func(t *testing.T) {
		repo := &stubUserRepo{}
		svc := NewUserService(repo, stubTransactor{})

		err := svc.CreateUser(ctx, &model.User{Email: "d@example.com"})
		require.NoError(t, err)
		require.Equal(t, 1, repo.countCalled)
		require.Equal(t, 1, repo.createCalled)
	})
}
