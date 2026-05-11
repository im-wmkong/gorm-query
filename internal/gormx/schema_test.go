package gormx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm/schema"
)

type tableNameModel struct {
	ID uint
}

// tableNameTabbed declares an explicit Table override via the GORM Tabler interface.
type tableNameTabbed struct {
	ID uint
}

func (tableNameTabbed) TableName() string { return "explicit_override" }

func TestTableName_ExplicitWins(t *testing.T) {
	sch := parseSchema(t, &tableNameTabbed{})
	got := TableName(sch, schema.NamingStrategy{SingularTable: true})
	assert.Equal(t, "explicit_override", got)
}

func TestTableName_FallsBackToNamingStrategy(t *testing.T) {
	sch := parseSchema(t, &tableNameModel{})
	// Force-empty Table so we hit the fallback path even if the parser pre-filled it.
	sch.Table = ""

	t.Run("singular", func(t *testing.T) {
		got := TableName(sch, schema.NamingStrategy{SingularTable: true})
		assert.Equal(t, "table_name_model", got)
	})

	t.Run("plural", func(t *testing.T) {
		got := TableName(sch, schema.NamingStrategy{})
		assert.Equal(t, "table_name_models", got)
	})
}
