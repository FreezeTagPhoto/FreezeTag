package formats

import (
	"fmt"
	"strings"
	"time"

	"gopkg.in/gographics/imagick.v3/imagick"
)

func extractGeoMetadata(mw *imagick.MagickWand) (*struct {
	Lat float64
	Lon float64
	Alt float64
}, error) {
	if latStr := mw.GetImageProperty("exif:GPSLatitude"); latStr != "" {
		lonStr := mw.GetImageProperty("exif:GPSLongitude")
		altStr := mw.GetImageProperty("exif:GPSAltitude")
		if lonStr == "" || altStr == "" {
			return nil, fmt.Errorf("incomplete GPS data")
		}
		lat, _ := convertGPSRational(latStr) //nolint:errcheck // GPS strings can never be invalid if they exist
		lon, _ := convertGPSRational(lonStr) //nolint:errcheck
		alt, _ := convertRational(altStr)    //nolint:errcheck
		return &struct {
			Lat float64
			Lon float64
			Alt float64
		}{
			lat,
			lon,
			alt,
		}, nil
	} else {
		return nil, nil
	}
}

func extractManufacturerMetadata(mw *imagick.MagickWand) *struct {
	Manufacturer string
	Model        string
} {
	if manufacturer := mw.GetImageProperty("exif:Make"); manufacturer != "" {
		model := mw.GetImageProperty("exif:Model")
		return &struct {
			Manufacturer string
			Model        string
		}{
			manufacturer,
			model,
		}
	} else {
		return nil
	}
}

func extractDateCreated(mw *imagick.MagickWand) *time.Time {
	// attempt to extract EXIF created date
	dateString := mw.GetImageProperty("exif:DateTimeOriginal")
	if dateString != "" {
		layout := "2006:01:02 15:04:05"
		parsed, _ := time.Parse(layout, dateString) //nolint:errcheck // dates always match this layout in EXIF
		return &parsed
	}
	return nil
}

// convert a 64-bit rational number (EXIF loves these)
// to a 64 bit float
func convertRational(rational string) (float64, error) {
	var num, den uint32
	_, err := fmt.Sscanf(rational, "%d/%d", &num, &den)
	if err != nil {
		return 0.0, fmt.Errorf("failed to parse rational64u: %w", err)
	}
	return float64(num) / float64(den), nil
}

// convert the 3-part rational degrees,minutes,seconds format
// used in EXIF to 64-bit decimal
func convertGPSRational(latitude string) (float64, error) {
	parts := strings.SplitN(latitude, ",", 3)
	if len(parts) < 3 {
		return 0.0, fmt.Errorf("3-part rational only had %d parts", len(parts))
	}
	degrees, err := convertRational(parts[0])
	if err != nil {
		return 0.0, err
	}
	minutes, err := convertRational(parts[1])
	if err != nil {
		return 0.0, err
	}
	seconds, err := convertRational(parts[2])
	if err != nil {
		return 0.0, err
	}
	return degrees + minutes/60. + seconds/3600., nil
}
