package imagedata

import "time"

// standardized image information
type Data struct {
	// image data, obviously will always appear
	PixelsRGBA []byte
	Width      int
	Height     int
	// image metadata, only some parts of this may appear
	// otherwise they'll be nil
	DateCreated *time.Time
	Geo         *struct {
		Lat float64
		Lon float64
		Alt float64
	}
	Cam *struct {
		Manufacturer string
		Model        string
	}
}
