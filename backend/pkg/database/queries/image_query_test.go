package queries

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	OnlyTagsQuery = "((id IN (SELECT imageId FROM Tags WHERE %s GROUP BY imageId HAVING COUNT(DISTINCT imageId) = ?)))"
)

func TestImageQueryEmpty(t *testing.T) {
	q := CreateImageQuery()
	s, as := q.StatementWithArgs()
	assert.Equal(t, "(TRUE)", s)
	assert.Equal(t, []any{}, as)
}

func TestImageQueryOneLiteralTag(t *testing.T) {
	q := CreateImageQuery().
		WithTag("test")
	s, as := q.StatementWithArgs()
	assert.Equal(t, fmt.Sprintf(OnlyTagsQuery, "tag IN (?)"), s)
	assert.Equal(t, []any{"test", 1}, as)
}

func TestImageQueryMultiLiteralTag(t *testing.T) {
	q := CreateImageQuery().
		WithTags("A", "B")
	s, as := q.StatementWithArgs()
	assert.Equal(t, fmt.Sprintf(OnlyTagsQuery, "tag IN (?, ?)"), s)
	assert.Equal(t, []any{"A", "B", 2}, as)
}

func TestImageQuerySingleFuzzyTag(t *testing.T) {
	q := CreateImageQuery().
		WithTagLike("par")
	s, as := q.StatementWithArgs()
	assert.Equal(t, fmt.Sprintf(OnlyTagsQuery, "(tag LIKE ? ESCAPE '!')"), s)
	assert.Equal(t, []any{"%par%", 1}, as)
}

func TestImageQueryMultiFuzzyTag(t *testing.T) {
	q := CreateImageQuery().
		WithTagsLike("par", "tial")
	s, as := q.StatementWithArgs()
	assert.Equal(t, fmt.Sprintf(OnlyTagsQuery, "(tag LIKE ? ESCAPE '!' OR tag LIKE ? ESCAPE '!')"), s)
	assert.Equal(t, []any{"%par%", "%tial%", 2}, as)
}

func TestImageQueryLiteralAndFuzzyTags(t *testing.T) {
	q := CreateImageQuery().
		WithTags("test", "foo").
		WithTagsLike("a", "b").
		WithTag("bar")
	s, as := q.StatementWithArgs()
	assert.Equal(t, fmt.Sprintf(OnlyTagsQuery, "tag IN (?, ?, ?) OR (tag LIKE ? ESCAPE '!' OR tag LIKE ? ESCAPE '!')"), s)
	assert.Equal(t, []any{"test", "foo", "bar", "%a%", "%b%", 5}, as)
}

func TestImageQueryLocationSimilar(t *testing.T) {
	q := CreateImageQuery().
		WithLocation(6.9, 42.0, 6.7)
	s, as := q.StatementWithArgs()
	assert.Equal(t, "((SQRT(POW(latitude - ?, 2), POW(longitude - ?, 2)) < ?))", s)
	assert.Equal(t, []any{6.9, 42.0, 6.7}, as)
}

func TestImageQueryMakeExact(t *testing.T) {
	q := CreateImageQuery().
		WithMake("test")
	s, as := q.StatementWithArgs()
	assert.Equal(t, "((cameraMake = ?))", s)
	assert.Equal(t, []any{"test"}, as)
}

func TestImageQueryMakeFuzzy(t *testing.T) {
	q := CreateImageQuery().
		WithMakeLike("te")
	s, as := q.StatementWithArgs()
	assert.Equal(t, "((cameraMake LIKE ? ESCAPE '!'))", s)
	assert.Equal(t, []any{"%te%"}, as)
}

func TestImageQueryModelExact(t *testing.T) {
	q := CreateImageQuery().
		WithModel("test")
	s, as := q.StatementWithArgs()
	assert.Equal(t, "((cameraModel = ?))", s)
	assert.Equal(t, []any{"test"}, as)
}

func TestImageQueryModelFuzzy(t *testing.T) {
	q := CreateImageQuery().
		WithModelLike("te")
	s, as := q.StatementWithArgs()
	assert.Equal(t, "((cameraModel LIKE ? ESCAPE '!'))", s)
	assert.Equal(t, []any{"%te%"}, as)
}

func TestImageQueryMakeAndModelExact(t *testing.T) {
	q := CreateImageQuery().
		WithMake("test").
		WithModel("foo")
	s, as := q.StatementWithArgs()
	assert.Equal(t, "((cameraMake = ?) AND (cameraModel = ?))", s)
	assert.Equal(t, []any{"test", "foo"}, as)
}

func TestImageQueryMakeExactModelFuzzy(t *testing.T) {
	q := CreateImageQuery().
		WithMake("test").
		WithModelLike("f")
	s, as := q.StatementWithArgs()
	assert.Equal(t, "((cameraMake = ?) AND (cameraModel LIKE ? ESCAPE '!'))", s)
	assert.Equal(t, []any{"test", "%f%"}, as)
}

func TestImageQueryTagAndMake(t *testing.T) {
	q := CreateImageQuery().
		WithTag("test").
		WithMake("foo")
	s, as := q.StatementWithArgs()
	assert.Equal(t, "((id IN (SELECT imageId FROM Tags WHERE tag IN (?) GROUP BY imageId HAVING COUNT(DISTINCT imageId) = ?)) AND (cameraMake = ?))", s)
	assert.Equal(t, []any{"test", 1, "foo"}, as)
}

func TestImageQueryAllPieces(t *testing.T) {
	q := CreateImageQuery().
		WithTag("test").
		WithTagLike("foo").
		WithMake("bar").
		WithModel("baz").
		WithLocation(6.9, 42.0, 67.0)
	s, as := q.StatementWithArgs()
	assert.Equal(t, "((id IN (SELECT imageId FROM Tags WHERE tag IN (?) OR (tag LIKE ? ESCAPE '!') GROUP BY imageId HAVING COUNT(DISTINCT imageId) = ?)) AND (SQRT(POW(latitude - ?, 2), POW(longitude - ?, 2)) < ?) AND (cameraMake = ?) AND (cameraModel = ?))", s)
	assert.Equal(t, []any{"test", "%foo%", 2, 6.9, 42.0, 67.0, "bar", "baz"}, as)
}
