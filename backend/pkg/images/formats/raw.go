package formats

import (
	"freezetag/backend/pkg/images/imagedata"
	"log"

	"gopkg.in/gographics/imagick.v3/imagick"
)

func ParseRaw(name string, data []byte) (imagedata.Data, error) {
	mw := imagick.NewMagickWand()
	defer mw.Destroy()
	if err := loadRawImage(name, data, mw); err != nil {
		return imagedata.Data{}, failedConversionError{name, err}
	}
	data, err := imageToRGBA(mw)
	if err != nil {
		return imagedata.Data{}, failedConversionError{name, err}
	}
	meta, err := parseEXIFData(mw)
	if err != nil {
		log.Printf("[WARNING] failed to extract EXIF from %v: %v", name, err)
	}
	width, height := int(mw.GetImageWidth()), int(mw.GetImageHeight())
	return imagedata.Data{
		PixelsRGBA:  data,
		Width:       width,
		Height:      height,
		DateCreated: meta.DateCreated,
		Geo:         meta.Geo,
		Cam:         meta.Cam,
	}, nil
}
