package test

import (
	"context"
	"testing"

	"github.com/im-wmkong/gorm-query/db"
	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/im-wmkong/gorm-query/example/model/schema"
	"github.com/im-wmkong/gorm-query/query"
	"github.com/im-wmkong/gorm-query/repo"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/clause"
)

// SQL() remains authoritative even when a custom fragment also implements
// GORM's expression interface with bound values.
type customFragment struct {
	clause.Expr
	sql string
}

func (f *customFragment) SQL() string {
	return f.sql
}

func TestCompatibility_CustomSQLFragments(t *testing.T) {
	for _, mode := range []string{"select", "distinct", "group", "order"} {
		t.Run(mode, func(t *testing.T) {
			d := openDB(t)
			seed(t, d)
			fragment := &customFragment{
				Expr: clause.Expr{SQL: "?", Vars: []any{999}},
				sql:  "age",
			}
			qb := schema.User.Query().Select(schema.User.Age)
			want := []int{10, 20, 30, 40}
			switch mode {
			case "select":
				qb = qb.Select(fragment).Order(schema.User.ID)
			case "distinct":
				qb = qb.Distinct(fragment).Order(schema.User.ID)
			case "group":
				qb = qb.Group(fragment).Order(schema.User.ID)
			case "order":
				fragment.sql = "age DESC"
				qb = qb.Order(fragment)
				want = []int{40, 30, 20, 10}
			}
			fragment.sql = "status"

			var users []model.User
			require.NoError(t, qb.Apply(d.Model(&model.User{})).Find(&users).Error)
			ages := make([]int, len(users))
			for i, user := range users {
				ages[i] = user.Age
			}
			require.Equal(t, want, ages)
		})
	}
}

func TestCompatibility_ProjectionControlsUpdates(t *testing.T) {
	for _, omit := range []bool{false, true} {
		name := "select"
		if omit {
			name = "omit"
		}
		t.Run(name, func(t *testing.T) {
			d := openDB(t)
			seed(t, d)
			r := repo.New[model.User](db.NewClient(d))
			ctx := context.Background()
			base := schema.User.Query().Where(schema.User.UserName.Eq("Alice"))
			qb := base.Select(schema.User.Age)
			if omit {
				qb = base.Omit(schema.User.Age)
			}
			n, err := r.Update(ctx, qb, schema.User.Age.Set(0), schema.User.Email.Set("changed@example.com"))
			require.NoError(t, err)
			require.EqualValues(t, 1, n)
			row, err := r.First(ctx, base)
			require.NoError(t, err)
			if omit {
				require.Equal(t, 10, row.Age)
				require.Equal(t, "changed@example.com", row.Email)
			} else {
				require.Zero(t, row.Age)
				require.Equal(t, "alice@example.com", row.Email)
			}
		})
	}
}

func TestContract_AliasAndProjectionCombinations(t *testing.T) {
	d := openDB(t)
	seed(t, d)
	age := schema.User.Age.WithTable("unused").WithTable("u")
	var users []model.User
	require.NoError(t, schema.User.Query().Where(age.Gte(20)).Order(age.Desc()).Apply(d.Table("users AS u")).Find(&users).Error)
	require.Len(t, users, 3)
	require.Equal(t, "Dave", users[0].UserName)
	require.Len(t, names(t, d, schema.User.Query().Where(schema.User.Age.Gte(20))), 3)
	r := repo.New[model.User](db.NewClient(d))
	ctx := context.Background()
	rows, err := r.Find(ctx, schema.User.Query().Select(schema.User.UserName).Joins(schema.User.Profile).Order(schema.User.ID))
	require.NoError(t, err)
	require.Len(t, rows, 4)
	require.Equal(t, "SF", rows[0].Profile.Bio)
	n, err := r.Count(ctx, schema.User.Query().Distinct(schema.User.Status))
	require.NoError(t, err)
	require.EqualValues(t, 1, n)
	var values []int
	require.NoError(t, r.Pluck(ctx, schema.User.Query().Select(schema.User.Age).Order(schema.User.ID), schema.User.Status, &values))
	require.Equal(t, []int{10, 20, 30, 40}, values)
	var sums []struct{ Total int }
	require.NoError(t, schema.User.Query().Select(schema.User.Age.Sum().As("total")).Apply(d.Model(&model.User{})).Scan(&sums).Error)
	require.Equal(t, 100, sums[0].Total)
	require.NoError(t, schema.User.Query().Order(query.RawFragment("age DESC")).Apply(d.Model(&model.User{})).Find(&users).Error)
}
