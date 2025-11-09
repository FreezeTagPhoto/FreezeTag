package queries

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFullImageIdQuery(t *testing.T) {
	q := CreateImageQuery()
	s, as := ImageIdPreparable(q)
	assert.Equal(t, "SELECT id FROM Images WHERE (TRUE)", s)
	assert.Equal(t, []any{}, as)
}
