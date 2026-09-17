package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/im-wmkong/gorm-query/db"
	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/im-wmkong/gorm-query/example/model/schema"
	"github.com/im-wmkong/gorm-query/query"
	"github.com/im-wmkong/gorm-query/repo"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestContract_GeneratedAssociationConditions(t *testing.T) {
	for _, mode := range []string{"preload", "left_join", "inner_join"} {
		t.Run(mode, func(t *testing.T) {
			d := openDB(t)
			seed(t, d)
			qb := schema.User.Query().Order(schema.User.ID.Asc())
			switch mode {
			case "preload":
				qb = qb.Preload(schema.User.Profile, schema.Profile.Bio.Eq("SF"))
			case "left_join":
				qb = qb.Joins(schema.User.Profile, schema.Profile.Bio.WithTable("Profile").Eq("SF"))
			case "inner_join":
				qb = qb.InnerJoins(schema.User.Profile, schema.Profile.Bio.WithTable("Profile").Eq("SF"))
			}
			users, err := repo.New[model.User](db.NewClient(d)).Find(context.Background(), qb)
			require.NoError(t, err)
			want := 4
			if mode == "inner_join" {
				want = 1
			}
			require.Len(t, users, want)
			require.Equal(t, "Alice", users[0].UserName)
			require.NotNil(t, users[0].Profile)
			require.Equal(t, "SF", users[0].Profile.Bio)
			for _, u := range users[1:] {
				require.Nil(t, u.Profile, "unmatched parents survive without an associated row")
			}
		})
	}
}

func TestContract_GeneratedProjectionAndZeroValueUpdate(t *testing.T) {
	d := openDB(t)
	seed(t, d)
	r := repo.New[model.User](db.NewClient(d))
	ctx := context.Background()
	base := schema.User.Query().Where(schema.User.UserName.Eq("Alice"))
	u, err := r.First(ctx, base.Select(schema.User.UserName))
	require.NoError(t, err)
	require.Equal(t, "Alice", u.UserName)
	require.Empty(t, u.Email)
	u, err = r.First(ctx, base.Omit(schema.User.Email))
	require.NoError(t, err)
	require.Equal(t, "Alice", u.UserName)
	require.Empty(t, u.Email)
	n, err := r.Update(ctx, base, schema.User.Age.Set(0), schema.User.Status.Set(0), schema.User.Email.Set(""))
	require.NoError(t, err)
	require.EqualValues(t, 1, n)
	u, err = r.First(ctx, base)
	require.NoError(t, err)
	require.Zero(t, u.Age)
	require.Zero(t, u.Status)
	require.Empty(t, u.Email)
}

func TestContract_BaseCountAndPagedDerivations(t *testing.T) {
	d := openDB(t)
	seed(t, d)
	r := repo.New[model.User](db.NewClient(d))
	base := schema.User.Query().Where(schema.User.Age.Gte(20)).Order(schema.User.ID.Asc())
	secondPage := base.Page(2, 1)
	require.Equal(t, []string{"Carol"}, names(t, d, secondPage))
	total, err := r.Count(context.Background(), base)
	require.NoError(t, err)
	require.EqualValues(t, 3, total)
	require.Equal(t, []string{"Bob", "Carol", "Dave"}, names(t, d, base))
	_, err = r.Delete(context.Background(), schema.User.Query().Where(schema.User.UserName.Eq("Carol")))
	require.NoError(t, err)
	total, err = r.Count(context.Background(), base)
	require.NoError(t, err)
	require.EqualValues(t, 2, total)
	total, err = r.Count(context.Background(), base.Unscoped())
	require.NoError(t, err)
	require.EqualValues(t, 3, total)
}

// Compatibility tests record GORM passthrough behavior; they do not define a
// new count API or change the meaning of an empty exclusion list.
func TestCompatibility_CountAndExistsKeepOffset(t *testing.T) {
	d := openDB(t)
	seed(t, d)
	r := repo.New[model.User](db.NewClient(d))
	ctx := context.Background()
	n, err := r.Count(ctx, schema.User.Query().Page(2, 1))
	require.NoError(t, err)
	require.Zero(t, n, "OFFSET removes the aggregate result row in current GORM semantics")
	ok, err := r.Exists(ctx, schema.User.Query().Offset(4))
	require.NoError(t, err)
	require.False(t, ok)
}

func TestCompatibility_EmptyMembershipAndBooleanGrouping(t *testing.T) {
	d := openDB(t)
	seed(t, d)
	for _, c := range []query.Condition{schema.User.ID.In(nil), schema.User.ID.NotIn(nil)} {
		require.Empty(t, names(t, d, schema.User.Query().Where(c)))
	}
	qb := schema.User.Query().Where(schema.User.UserName.Eq("Alice")).Or(
		schema.User.Age.Gte(30), schema.User.Age.Lt(40),
	)
	require.Equal(t, []string{"Alice", "Carol"}, names(t, d, qb))
	require.Equal(t, []string{"Carol", "Dave"}, names(t, d, schema.User.Query().Not(schema.User.Age.Lt(30))))
}

func TestContract_GroupHavingDistinctAndCount(t *testing.T) {
	d := openDB(t)
	seed(t, d)
	r := repo.New[model.User](db.NewClient(d))
	ctx := context.Background()
	_, err := r.Update(ctx, schema.User.Query().Where(schema.User.Age.Gte(30)), schema.User.Status.Set(2))
	require.NoError(t, err)
	qb := schema.User.Query().Select(schema.User.Status, schema.User.ID.Count().As("n")).
		Group(schema.User.Status).Having("COUNT(*) > ?", 1).Order(schema.User.Status.Asc())
	var groups []struct {
		Status int
		N      int
	}
	require.NoError(t, qb.Apply(d.Model(&model.User{})).Scan(&groups).Error)
	require.Len(t, groups, 2)
	require.Equal(t, 2, groups[0].N)
	require.Equal(t, 2, groups[1].N)
	n, err := r.Count(ctx, schema.User.Query().Distinct(schema.User.Status))
	require.NoError(t, err)
	require.EqualValues(t, 2, n)
}

func TestContract_ConcurrentBuilderReuse(t *testing.T) {
	d := openDB(t)
	seed(t, d)
	r := repo.New[model.User](db.NewClient(d))
	base := schema.User.Query().Where(schema.User.Status.Eq(1))
	const workers = 16
	start, results := make(chan struct{}), make(chan error, workers)
	for i := 0; i < workers; i++ {
		want := []string{"Alice", "Bob", "Carol", "Dave"}[i%4]
		go func(want string) {
			<-start
			u, err := r.First(context.Background(), base.Where(schema.User.UserName.Eq(want)))
			if err == nil && u.UserName != want {
				err = fmt.Errorf("got %s, want %s", u.UserName, want)
			}
			results <- err
		}(want)
	}
	close(start)
	var firstErr error
	for i := 0; i < workers; i++ {
		if err := <-results; err != nil && firstErr == nil {
			firstErr = err
		}
	}
	require.NoError(t, firstErr)
	require.Len(t, names(t, d, base), 4)
}

func TestContract_EmptyUpdateAndMissingWherePreserveData(t *testing.T) {
	d := openDB(t)
	seed(t, d)
	r := repo.New[model.User](db.NewClient(d))
	ctx := context.Background()
	_, err := r.Update(ctx, schema.User.Query())
	require.ErrorIs(t, err, query.ErrNoAssignment)
	_, err = r.Update(ctx, schema.User.Query(), schema.User.Age.Set(0))
	require.ErrorIs(t, err, gorm.ErrMissingWhereClause)
	_, err = r.Delete(ctx, nil)
	require.ErrorIs(t, err, gorm.ErrMissingWhereClause)
	require.Equal(t, []string{"Alice", "Bob", "Carol", "Dave"}, names(t, d, schema.User.Query().Where(schema.User.Age.Gt(0))))
}
