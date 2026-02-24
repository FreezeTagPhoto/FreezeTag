package images

import (
	"fmt"
	"freezetag/backend/pkg/images/imagedata"

	"gopkg.in/gographics/imagick.v3/imagick"
)

// Create a WEBP profile picture from image RGBA bytes.
// You can pass a quality between 0 and 1 (default is 0.8 lossy if quality=0, quality=1 is lossless),
// and a maxHeight to shrink the image while keeping aspect ratio (default 0 to retain original resolution)
func CreateProfilePicture(data imagedata.Data) ([]byte, error) {
	if len(data.PixelsRGBA) != data.Width*data.Height*4 {
		return nil, fmt.Errorf("incorrect image data structure")
	}

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	err := mw.ConstituteImage(uint(data.Width), uint(data.Height), "RGBA", imagick.PIXEL_CHAR, data.PixelsRGBA)
	if err != nil {
		return nil, err
	}


	w, h := int(data.Width), int(data.Height)
	var size, x, y int

	if w > h {
		size = h
		x = int((w - h) / 2)
		y = 0
	} else {
		size = w
		x = 0
		y = int((h - w) / 2)
	}
	err = mw.CropImage(uint(size), uint(size), int(x), int(y))
	if err != nil {
		return nil, err
	}
	err = mw.ResizeImage(256, 256, imagick.FILTER_LANCZOS)
	if err != nil {
		return nil, err
	}

	mw.SetImageFormat("WEBP")
	mw.StripImage()
	mw.SetImageCompressionQuality(uint(0.8 * 100))
	mw.SetOption("webp:lossless", "false")

	return mw.GetImageBlob()
}
