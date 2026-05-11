package fragment

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakeFragment string

func (f fakeFragment) SQL() string { return string(f) }

func TestFragmentHelpers(t *testing.T) {
	t.Run("RenderAll renders each fragment's SQL", func(t *testing.T) {
		got := RenderAll([]fakeFragment{"a", "b", "c"})
		assert.Equal(t, []string{"a", "b", "c"}, got)
		assert.Nil(t, RenderAll[fakeFragment](nil))
	})

	t.Run("RenderAllAny boxes rendered SQL into any", func(t *testing.T) {
		got := RenderAllAny([]fakeFragment{"a", "b"})
		assert.Equal(t, []any{"a", "b"}, got)
		assert.Nil(t, RenderAllAny[fakeFragment](nil))
	})

	t.Run("Suffix appends suffix", func(t *testing.T) {
		assert.Equal(t, "age DESC", Suffix("age", " DESC"))
		assert.Equal(t, "age AS a", Suffix("age", " AS a"))
	})

	t.Run("Prefix prepends prefix", func(t *testing.T) {
		assert.Equal(t, "DISTINCT age", Prefix("DISTINCT ", "age"))
	})

	t.Run("Call builds fn(expr)", func(t *testing.T) {
		assert.Equal(t, "SUM(age)", Call("SUM", "age"))
		assert.Equal(t, "COUNT(*)", Call("COUNT", "*"))
	})

	t.Run("JoinPath handles empty sides", func(t *testing.T) {
		assert.Equal(t, "a.b", JoinPath("a", "b", "."))
		assert.Equal(t, "b", JoinPath("", "b", "."))
		assert.Equal(t, "a", JoinPath("a", "", "."))
		assert.Equal(t, "", JoinPath("", "", "."))
	})
}
