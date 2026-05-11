package gormx

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/schema"
)

type relParent struct {
	ID    uint
	B     relChildB `gorm:"foreignKey:ParentID"`
	A     relChildA `gorm:"foreignKey:ParentID"`
	Plain string
}

type relChildA struct {
	ID       uint
	ParentID uint
}

type relChildB struct {
	ID       uint
	ParentID uint
}

func parseSchema(t *testing.T, model any) *schema.Schema {
	t.Helper()
	sch, err := schema.Parse(model, &sync.Map{}, schema.NamingStrategy{})
	require.NoError(t, err)
	return sch
}

func TestOrderedRelations(t *testing.T) {
	t.Run("nil schema returns nil", func(t *testing.T) {
		assert.Nil(t, OrderedRelations(nil))
	})

	t.Run("model without relations returns nil", func(t *testing.T) {
		sch := parseSchema(t, &relChildA{})
		assert.Nil(t, OrderedRelations(sch))
	})

	t.Run("preserves declaration order", func(t *testing.T) {
		sch := parseSchema(t, &relParent{})
		got := OrderedRelations(sch)
		require.Len(t, got, 2)
		assert.Equal(t, "B", got[0].Name)
		assert.Equal(t, "A", got[1].Name)
	})

	t.Run("child type points at related model", func(t *testing.T) {
		sch := parseSchema(t, &relParent{})
		got := OrderedRelations(sch)
		require.Len(t, got, 2)
		// Both children come from this same package; checking the type name is enough.
		assert.Equal(t, "relChildB", got[0].ChildType.Name())
		assert.Equal(t, "relChildA", got[1].ChildType.Name())
	})
}
