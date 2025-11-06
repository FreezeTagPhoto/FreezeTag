package formats

import (
	"fmt"
	"freezetag/backend/pkg/images"
	"log"

	"gopkg.in/gographics/imagick.v3/imagick"
)

func ParsePNG(name string, data []byte) (images.Data, error) {
	mw := imagick.NewMagickWand()
	defer mw.Destroy()
	if err := mw.ReadImageBlob(data); err != nil {
		return images.Data{}, fmt.Errorf("failed to convert %v: %w", name, err)
	}
	// get image data
	data, err := imageToRGBA(mw)
	if err != nil {
		return images.Data{}, fmt.Errorf("failed to convert %v: %w", name, err)
	}
	// get image metadata
	geo, err := extractGeoMetadata(mw)
	if err != nil {
		log.Printf("[WARNING] failed to extract GPS metadata from %v: %v\n", name, err)
	}
	cam := extractManufacturerMetadata(mw)
	created := extractDateCreated(mw)
	width, height := int(mw.GetImageWidth()), int(mw.GetImageHeight())
	return images.Data{
		PixelsRGBA:  data,
		Width:       width,
		Height:      height,
		DateCreated: created,
		Geo:         geo,
		Cam:         cam,
	}, nil
}
