package queries

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testQuery string

func (t testQuery) StatementWithArgs() (string, []any) {
	return string(t), []any{string(t)}
}

func TestAndQueryOne(t *testing.T) {
	s, as := And(testQuery("foo")).StatementWithArgs()
	assert.Equal(t, "(foo)", s)
	assert.Equal(t, []any{"foo"}, as)
}

func TestAndQueryMulti(t *testing.T) {
	s, as := And(testQuery("foo"), testQuery("bar"), testQuery("baz")).StatementWithArgs()
	assert.Equal(t, "(foo) AND (bar) AND (baz)", s)
	assert.Equal(t, []any{"foo", "bar", "baz"}, as)
}

func TestOrQueryOne(t *testing.T) {
	s, as := Or(testQuery("foo")).StatementWithArgs()
	assert.Equal(t, "(foo)", s)
	assert.Equal(t, []any{"foo"}, as)
}

func TestOrQueryMulti(t *testing.T) {
	s, as := Or(testQuery("foo"), testQuery("bar"), testQuery("baz")).StatementWithArgs()
	assert.Equal(t, "(foo) OR (bar) OR (baz)", s)
	assert.Equal(t, []any{"foo", "bar", "baz"}, as)
}
