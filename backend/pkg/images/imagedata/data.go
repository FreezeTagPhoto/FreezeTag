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

// metadata stores a more limited set of image information
// in a way that can be easily serialized to JSON for frontend serving
type Metadata struct {
	FileName     *string  `json:"fileName"`
	DateTaken    *int64   `json:"dateTaken"`
	DateUploaded *int64   `json:"dateUploaded"`
	CameraMake   *string  `json:"cameraMake"`
	CameraModel  *string  `json:"cameraModel"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
}
