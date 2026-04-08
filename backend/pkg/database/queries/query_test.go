package queries

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFullImageIdQuery(t *testing.T) {
	q := CreateImageQuery()
	s, as := ImageIDPreparable(q, DateAdded, Descending, 0, 0)
	assert.Equal(t, "SELECT id FROM Images WHERE (TRUE) ORDER BY dateUploaded DESC", s)
	assert.Equal(t, []any{}, as)
}

func TestImageIdQueryVariants(t *testing.T) {
	q := CreateImageQuery()
	s, _ := ImageIDPreparable(q, DateAdded, Descending, 0, 0)
	assert.Contains(t, s, "ORDER BY dateUploaded DESC")
	s, _ = ImageIDPreparable(q, DateCreated, Ascending, 0, 0)
	assert.Contains(t, s, "ORDER BY COALESCE(dateTaken,dateUploaded) ASC")
	s, _ = ImageIDPreparable(q, DateCreated, Descending, 0, 0)
	assert.Contains(t, s, "ORDER BY COALESCE(dateTaken,dateUploaded) DESC")
	// test invalid enum variant just in case (default)
	s, _ = ImageIDPreparable(q, SortField(255), SortOrder(254), 0, 0)
	assert.Contains(t, s, "ORDER BY dateUploaded DESC")
}

func TestImageIdQueryPaged(t *testing.T) {
	q := CreateImageQuery()
	s, _ := ImageIDPreparable(q, DateAdded, Descending, 5, 2)
	assert.Contains(t, s, "LIMIT 5 OFFSET 10")
	s, _ = ImageIDPreparable(q, DateAdded, Descending, 2, 5)
	assert.Contains(t, s, "LIMIT 2 OFFSET 10")
	s, _ = ImageIDPreparable(q, DateAdded, Descending, 100, 0)
	assert.Contains(t, s, "LIMIT 100 OFFSET 0")
}
