package service

import (
	"context"
	"errors"
	"testing"

	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/im-wmkong/gorm-query/query"
	"github.com/im-wmkong/gorm-query/repo"

	"github.com/stretchr/testify/require"
)

type stubTransactor struct{}

func (stubTransactor) Transaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type stubUserRepo struct {
	repo.Repository[model.User] // Unused methods fail instead of simulating database behavior.
	count                       int64
	// return values
	countErr  error
	createErr error

	countCalled  int
	createCalled int
}

func (r *stubUserRepo) Create(_ context.Context, _ *model.User) error {
	r.createCalled++
	return r.createErr
}
func (r *stubUserRepo) Count(context.Context, *query.Builder[model.User]) (int64, error) {
	r.countCalled++
	return r.count, r.countErr
}

func TestUserService_CreateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("count error", func(t *testing.T) {
		repo := &stubUserRepo{countErr: errors.New("count failed")}
		svc := NewUserService(repo, stubTransactor{})

		err := svc.CreateUser(ctx, &model.User{Email: "a@example.com"})
		require.ErrorIs(t, err, repo.countErr)
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
		require.ErrorIs(t, err, repo.createErr)
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
