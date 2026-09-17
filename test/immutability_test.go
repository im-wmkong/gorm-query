package test

import (
	"testing"

	"github.com/im-wmkong/gorm-query/example/model"
	"github.com/im-wmkong/gorm-query/example/model/schema"
	"github.com/im-wmkong/gorm-query/query"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestContract_BuilderSnapshotsMembershipValues(t *testing.T) {
	for _, exclude := range []bool{false, true} {
		name := "in"
		if exclude {
			name = "not_in"
		}
		t.Run(name, func(t *testing.T) {
			d := openDB(t)
			seed(t, d)
			values := make([]string, 1, 4)
			values[0] = "Alice"
			cond := schema.User.UserName.In(values)
			want := []string{"Alice"}
			if exclude {
				cond = schema.User.UserName.NotIn(values)
				want = []string{"Bob", "Carol", "Dave"}
			}
			qb := schema.User.Query().Where(cond)
			require.Equal(t, want, names(t, d, qb))
			values[0] = "Bob"
			require.Equal(t, want, names(t, d, qb))
			values = append(values[:0], "Carol", "Dave")
			require.Len(t, values, 2)
			require.Equal(t, want, names(t, d, qb))
		})
	}
}

func TestContract_BuilderSnapshotsConditionSlices(t *testing.T) {
	for _, mode := range []string{"where", "or", "not"} {
		t.Run(mode, func(t *testing.T) {
			d := openDB(t)
			seed(t, d)
			conds := []query.Condition{schema.User.UserName.Eq("Alice")}
			qb := schema.User.Query()
			want := []string{"Alice"}
			switch mode {
			case "where":
				qb = qb.Where(conds...)
			case "or":
				qb = qb.Or(conds...)
			case "not":
				qb = qb.Not(conds...)
				want = []string{"Bob", "Carol", "Dave"}
			}
			require.Equal(t, want, names(t, d, qb))
			conds[0] = schema.User.UserName.Eq("Bob")
			require.Equal(t, want, names(t, d, qb))
		})
	}
}

func TestContract_BuilderSnapshotsScopesAndHavingArgs(t *testing.T) {
	t.Run("scopes", func(t *testing.T) {
		d := openDB(t)
		seed(t, d)
		scopes := []func(*gorm.DB) *gorm.DB{func(tx *gorm.DB) *gorm.DB { return tx.Where("age = ?", 10) }}
		qb := schema.User.Query().Scope(scopes...)
		require.Equal(t, []string{"Alice"}, names(t, d, qb))
		scopes[0] = func(tx *gorm.DB) *gorm.DB { return tx.Where("age = ?", 20) }
		require.Equal(t, []string{"Alice"}, names(t, d, qb))
	})
	t.Run("having", func(t *testing.T) {
		d := openDB(t)
		seed(t, d)
		args := []any{0}
		qb := schema.User.Query().Select(schema.User.UserName).Group(schema.User.UserName).Having("COUNT(*) > ?", args...)
		want := []string{"Alice", "Bob", "Carol", "Dave"}
		require.Equal(t, want, names(t, d, qb))
		args[0] = 100
		require.Equal(t, want, names(t, d, qb))
	})
}

func TestContract_BuilderSnapshotsAssociationConditions(t *testing.T) {
	for _, mode := range []string{"preload", "left_join", "inner_join"} {
		t.Run(mode, func(t *testing.T) {
			d := openDB(t)
			seed(t, d)
			// Bio is unambiguous in these models, so a bare column works for both Preload and JOIN.
			bio := query.NewStringColumn[string]("", "bio")
			conds := []query.Condition{bio.Eq("SF")}
			qb := schema.User.Query().Where(schema.User.UserName.Eq("Alice"))
			switch mode {
			case "preload":
				qb = qb.Preload(schema.User.Profile, conds...)
			case "left_join":
				qb = qb.Joins(schema.User.Profile, conds...)
			case "inner_join":
				qb = qb.InnerJoins(schema.User.Profile, conds...)
			}
			check := func() {
				var users []model.User
				require.NoError(t, qb.Apply(d.Model(&model.User{})).Find(&users).Error)
				require.Len(t, users, 1)
				require.NotNil(t, users[0].Profile)
				require.Equal(t, "SF", users[0].Profile.Bio)
			}
			check()
			conds[0] = bio.Eq("NY")
			check()
		})
	}
}
