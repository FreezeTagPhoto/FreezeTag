package formats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// these are just good-practice tests for covering errors
// in helpers where the EXIF data can just never be formatted wrong

func TestConvertRationalInvalid(t *testing.T) {
	_, err := convertRational("bad")
	assert.ErrorContains(t, err, "failed to parse rational64u")
}

func TestConvertTripleRationalInvalid(t *testing.T) {
	_, err := convertGPSRational("bad,worse")
	assert.ErrorContains(t, err, "3-part rational only had 2 parts")
}

func TestConvertTripleRationalAllFails(t *testing.T) {
	_, err := convertGPSRational("a,1/2,2/3")
	assert.Error(t, err)
	_, err = convertGPSRational("1/2,a,2/3")
	assert.Error(t, err)
	_, err = convertGPSRational("1/2,2/3,a")
	assert.Error(t, err)
}
