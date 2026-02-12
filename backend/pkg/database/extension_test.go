package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// DO NOT USE OUTSIDE OF UNIT TESTS.
// this is just to test that database extensions work as expected
func (db SqliteImageDatabase) executeQuery(query string) (err error) {
	_, err = db.db.Exec(query)
	return
}

func TestRegexMatch(t *testing.T) {
	tmp := createTempDatabase(t).(SqliteImageDatabase)
	err := tmp.executeQuery("SELECT regexp('test', 'test string')")
	assert.NoError(t, err)
}

func TestInvalidRegexMatch(t *testing.T) {
	tmp := createTempDatabase(t).(SqliteImageDatabase)
	err := tmp.executeQuery("SELECT regexp('[', 'test string')")
	assert.Error(t, err)
}

func TestRegexCapture(t *testing.T) {
	tmp := createTempDatabase(t).(SqliteImageDatabase)
	err := tmp.executeQuery("SELECT rextract('(test)', 'test string')")
	assert.NoError(t, err)
}

func TestInvalidRegexCapture(t *testing.T) {
	tmp := createTempDatabase(t).(SqliteImageDatabase)
	err := tmp.executeQuery("SELECT rextract('[', 'test string')")
	assert.Error(t, err)
}

func TestNoRegexCapture(t *testing.T) {
	tmp := createTempDatabase(t).(SqliteImageDatabase)
	err := tmp.executeQuery("SELECT rextract('test', 'test string')")
	assert.Error(t, err)
}
