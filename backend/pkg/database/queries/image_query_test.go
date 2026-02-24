package queries

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestImageQueryEmpty(t *testing.T) {
	q := CreateImageQuery()
	s, as := q.StatementWithArgs()
	assert.Equal(t, "(TRUE)", s)
	assert.Equal(t, []any{}, as)
}

func TestTags(t *testing.T) {
	q, args := CreateImageQuery().WithTag("foo").WithTagLike("bar").StatementWithArgs()
	assert.Equal(t, "((id IN (SELECT imageId FROM ImageTags INNER JOIN Tags ON Tags.id = ImageTags.tagId WHERE tag = ? OR tag LIKE ? ESCAPE '!' GROUP BY imageId HAVING COUNT(DISTINCT CASE WHEN tag = ? THEN 0 WHEN tag LIKE ? ESCAPE '!' THEN 1 ELSE NULL END) = ?)))", q)
	assert.Equal(t, []any{"foo", "%bar%", "foo", "%bar%", 2}, args)
}

func TestImageQueryLocationSimilar(t *testing.T) {
	q := CreateImageQuery().
		WithLocation(6.9, 42.0, 6.7)
	s, as := q.StatementWithArgs()
	assert.Equal(t, "((latitude IS NOT NULL AND longitude IS NOT NULL AND gcirc(latitude, longitude, ?, ?) <= ?))", s)
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

func TestImageQueryTakenBefore(t *testing.T) {
	now := time.Now()
	q := CreateImageQuery().TakenBefore(now)
	s, as := q.StatementWithArgs()
	assert.Equal(t, "((dateTaken <= ?))", s)
	assert.Equal(t, []any{now.Unix()}, as)
}

func TestImageQueryTakenAfter(t *testing.T) {
	now := time.Now()
	q := CreateImageQuery().TakenAfter(now)
	s, as := q.StatementWithArgs()
	assert.Equal(t, "((dateTaken >= ?))", s)
	assert.Equal(t, []any{now.Unix()}, as)
}

func TestImageQueryUploadedBefore(t *testing.T) {
	now := time.Now()
	q := CreateImageQuery().UploadedBefore(now)
	s, as := q.StatementWithArgs()
	assert.Equal(t, "((dateUploaded <= ?))", s)
	assert.Equal(t, []any{now.Unix()}, as)
}

func TestImageQueryUploadedAfter(t *testing.T) {
	now := time.Now()
	q := CreateImageQuery().UploadedAfter(now)
	s, as := q.StatementWithArgs()
	assert.Equal(t, "((dateUploaded >= ?))", s)
	assert.Equal(t, []any{now.Unix()}, as)
}

func TestImageQueryUploadedRange(t *testing.T) {
	now := time.Now()
	then := now.Add(-24 * time.Hour)
	q := CreateImageQuery().UploadedAfter(then).UploadedBefore(now)
	s, as := q.StatementWithArgs()
	assert.Equal(t, "((dateUploaded >= ?) AND (dateUploaded <= ?))", s)
	assert.Equal(t, []any{then.Unix(), now.Unix()}, as)
}

func TestImageQueryTakenRange(t *testing.T) {
	now := time.Now()
	then := now.Add(-24 * time.Hour)
	q := CreateImageQuery().TakenAfter(then).TakenBefore(now)
	s, as := q.StatementWithArgs()
	assert.Equal(t, "((dateTaken >= ?) AND (dateTaken <= ?))", s)
	assert.Equal(t, []any{then.Unix(), now.Unix()}, as)
}

func TestImageQueryMakeUploadedTakenAfter(t *testing.T) {
	now := time.Now()
	q := CreateImageQuery().
		TakenAfter(now).
		UploadedAfter(now).
		WithMake("test")
	s, as := q.StatementWithArgs()
	assert.Equal(t, "((cameraMake = ?) AND (dateTaken >= ?) AND (dateUploaded >= ?))", s)
	assert.Equal(t, []any{"test", now.Unix(), now.Unix()}, as)
}
