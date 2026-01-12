package queries

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFullImageIdQuery(t *testing.T) {
	q := CreateImageQuery()
	s, as := ImageIdPreparable(q, DateAdded, Descending)
	assert.Equal(t, "SELECT id, latitude, longitude FROM Images WHERE (TRUE) ORDER BY dateUploaded DESC", s)
	assert.Equal(t, []any{}, as)
}

func TestImageIdQueryVariants(t *testing.T) {
	q := CreateImageQuery()
	s, _ := ImageIdPreparable(q, DateAdded, Descending)
	assert.Contains(t, s, "ORDER BY dateUploaded DESC")
	s, _ = ImageIdPreparable(q, DateCreated, Ascending)
	assert.Contains(t, s, "ORDER BY COALESCE(dateTaken,dateUploaded) ASC")
	s, _ = ImageIdPreparable(q, DateCreated, Descending)
	assert.Contains(t, s, "ORDER BY COALESCE(dateTaken,dateUploaded) DESC")
	// test invalid enum variant just in case (default)
	s, _ = ImageIdPreparable(q, SortField(255), SortOrder(254))
	assert.Contains(t, s, "ORDER BY dateUploaded DESC")
}
