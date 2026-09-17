package test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/im-wmkong/gorm-query/db"
	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/im-wmkong/gorm-query/example/model/schema"
	"github.com/im-wmkong/gorm-query/repo"
	"github.com/stretchr/testify/require"
)

func TestContract_TransactionClientOwnership(t *testing.T) {
	for _, rollback := range []bool{false, true} {
		name := "commit_A"
		if rollback {
			name = "rollback_A"
		}
		t.Run(name, func(t *testing.T) {
			a, b := openDB(t), openDB(t)
			ca, cb := db.NewClient(a), db.NewClient(b)
			rb := repo.New[model.User](cb)
			stop := errors.New("rollback A")
			err := ca.Transaction(context.Background(), func(ctx context.Context) error {
				if err := rb.Create(ctx, &model.User{UserName: "B only", Email: "b@example.com"}); err != nil {
					return err
				}
				if rollback {
					return stop
				}
				return nil
			})
			if rollback {
				require.ErrorIs(t, err, stop)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, []string{}, names(t, a, schema.User.Query()))
			require.Equal(t, []string{"B only"}, names(t, b, schema.User.Query()))
		})
	}
}

func TestContract_ClientsSharingDBKeepTransactionsSeparate(t *testing.T) {
	d := openDB(t)
	ca, cb := db.NewClient(d), db.NewClient(d)
	ra, rb := repo.New[model.User](ca), repo.New[model.User](cb)
	stop := errors.New("rollback A")
	err := ca.Transaction(context.Background(), func(aCtx context.Context) error {
		// Write through B before A writes, so SQLite's writer lock does not
		// prevent the independent transaction from committing.
		if err := cb.Transaction(aCtx, func(bCtx context.Context) error {
			return rb.Create(bCtx, &model.User{UserName: "B committed", Email: "b@example.com"})
		}); err != nil {
			return err
		}
		if err := ra.Create(aCtx, &model.User{UserName: "A rolled back", Email: "a@example.com"}); err != nil {
			return err
		}
		return stop
	})
	require.ErrorIs(t, err, stop)
	require.Equal(t, []string{"B committed"}, names(t, d, schema.User.Query()))
}

func TestContract_NestedClientsPreserveBothTransactions(t *testing.T) {
	a, b := openDB(t), openDB(t)
	ca, cb := db.NewClient(a), db.NewClient(b)
	ra, rb := repo.New[model.User](ca), repo.New[model.User](cb)
	stop := errors.New("rollback B")
	err := ca.Transaction(context.Background(), func(aCtx context.Context) error {
		if err := ra.Create(aCtx, &model.User{UserName: "A before", Email: "a1@example.com"}); err != nil {
			return err
		}
		err := cb.Transaction(aCtx, func(bCtx context.Context) error {
			if err := ra.Create(bCtx, &model.User{UserName: "A inside B", Email: "a2@example.com"}); err != nil {
				return err
			}
			if err := rb.Create(bCtx, &model.User{UserName: "B rolled back", Email: "b@example.com"}); err != nil {
				return err
			}
			return stop
		})
		if !errors.Is(err, stop) {
			return fmt.Errorf("B transaction returned %v, want %v", err, stop)
		}
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, []string{"A before", "A inside B"}, names(t, a, schema.User.Query()))
	require.Empty(t, names(t, b, schema.User.Query()))
}

func TestContract_RepositoriesShareClientTransaction(t *testing.T) {
	d := openDB(t)
	c := db.NewClient(d)
	users, profiles := repo.New[model.User](c), repo.New[model.Profile](c)
	stop := errors.New("rollback both repositories")
	err := c.Transaction(context.Background(), func(ctx context.Context) error {
		u := &model.User{UserName: "Alice", Email: "a@example.com"}
		if err := users.Create(ctx, u); err != nil {
			return err
		}
		if err := profiles.Create(ctx, &model.Profile{UserID: u.ID, Bio: "SF"}); err != nil {
			return err
		}
		return stop
	})
	require.ErrorIs(t, err, stop)
	n, err := users.Count(context.Background(), nil)
	require.NoError(t, err)
	require.Zero(t, n)
	n, err = profiles.Count(context.Background(), nil)
	require.NoError(t, err)
	require.Zero(t, n)
}

func TestContract_ContextStopsRepositoryOperations(t *testing.T) {
	for _, expired := range []bool{false, true} {
		name := "cancelled"
		if expired {
			name = "deadline"
		}
		t.Run(name, func(t *testing.T) {
			d := openDB(t)
			c := db.NewClient(d)
			r := repo.New[model.User](c)
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if expired {
				ctx, cancel = context.WithDeadline(context.Background(), time.Unix(1, 0))
				defer cancel()
			}
			_, err := r.Find(ctx, schema.User.Query())
			require.ErrorIs(t, err, ctx.Err())
			err = r.Create(ctx, &model.User{Email: "cancelled@example.com"})
			require.ErrorIs(t, err, ctx.Err())
			called := false
			err = c.Transaction(ctx, func(context.Context) error {
				called = true
				return nil
			})
			require.ErrorIs(t, err, ctx.Err())
			require.False(t, called)
			require.Empty(t, names(t, d, schema.User.Query()))
		})
	}
}

func TestContract_IndependentClientCommitOutcomes(t *testing.T) {
	for _, rollbackA := range []bool{false, true} {
		for _, rollbackB := range []bool{false, true} {
			t.Run(fmt.Sprintf("rollback_A=%t_B=%t", rollbackA, rollbackB), func(t *testing.T) {
				a, b := openDB(t), openDB(t)
				ca, cb := db.NewClient(a), db.NewClient(b)
				ra, rb := repo.New[model.User](ca), repo.New[model.User](cb)
				stopA, stopB := errors.New("A"), errors.New("B")
				err := ca.Transaction(context.Background(), func(ctx context.Context) error {
					if err := ra.Create(ctx, &model.User{UserName: "A", Email: "a@example.com"}); err != nil {
						return err
					}
					err := cb.Transaction(ctx, func(ctx context.Context) error {
						if err := rb.Create(ctx, &model.User{UserName: "B", Email: "b@example.com"}); err != nil {
							return err
						}
						if rollbackB {
							return stopB
						}
						return nil
					})
					if rollbackB {
						require.ErrorIs(t, err, stopB)
					} else {
						require.NoError(t, err)
					}
					if rollbackA {
						return stopA
					}
					return nil
				})
				if rollbackA {
					require.ErrorIs(t, err, stopA)
				} else {
					require.NoError(t, err)
				}
				if rollbackA {
					require.Empty(t, names(t, a, schema.User.Query()))
				} else {
					require.Equal(t, []string{"A"}, names(t, a, schema.User.Query()))
				}
				if rollbackB {
					require.Empty(t, names(t, b, schema.User.Query()))
				} else {
					require.Equal(t, []string{"B"}, names(t, b, schema.User.Query()))
				}
			})
		}
	}
}

func TestContract_TransactionAbortsAfterWrite(t *testing.T) {
	for _, mode := range []string{"panic", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			d := openDB(t)
			c := db.NewClient(d)
			r := repo.New[model.User](c)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			run := func() {
				err := c.Transaction(ctx, func(txCtx context.Context) error {
					if err := r.Create(txCtx, &model.User{Email: "abort@example.com"}); err != nil {
						return err
					}
					if mode == "panic" {
						panic("abort")
					}
					cancel()
					_, err := r.Find(txCtx, schema.User.Query())
					require.Error(t, err)
					return err
				})
				require.Error(t, err)
			}
			if mode == "panic" {
				require.PanicsWithValue(t, "abort", run)
			} else {
				run()
			}
			require.Empty(t, names(t, d, schema.User.Query()))
		})
	}
}
