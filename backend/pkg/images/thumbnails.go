package images

import (
	"fmt"
	"freezetag/backend/pkg/images/imagedata"

	"gopkg.in/gographics/imagick.v3/imagick"
)

// Create a WEBP thumbnail from image RGBA bytes.
// You can pass a quality between 0 and 1 (default is 0.8 lossy if quality=0, quality=1 is lossless),
// and a maxHeight to shrink the image while keeping aspect ratio (default 0 to retain original resolution)
func CreateThumbnail(data imagedata.Data, maxHeight int, quality float32) ([]byte, error) {
	if len(data.PixelsRGBA) != data.Width*data.Height*4 {
		return []byte{}, fmt.Errorf("incorrect image data structure: %d bytes for a %dx%d image", len(data.PixelsRGBA), data.Width, data.Height)
	}
	mw := imagick.NewMagickWand()
	defer mw.Destroy()
	err := mw.ConstituteImage(uint(data.Width), uint(data.Height), "RGBA", imagick.PIXEL_CHAR, data.PixelsRGBA)
	if err != nil {
		return []byte{}, err
	}
	err = mw.SetImageFormat("WEBP")
	if err != nil {
		return []byte{}, err
	}
	if maxHeight > 0 && data.Height > maxHeight {
		change := float32(maxHeight) / float32(data.Height)
		err = mw.ResizeImage(uint(float32(data.Width)*change), uint(data.Height), imagick.FILTER_BOX)
		if err != nil {
			return []byte{}, err
		}
	}
	if quality < 1 {
		if quality == 0 {
			quality = 0.8
		}
		err = mw.SetOption("webp:lossless", "false")
		if err != nil {
			return []byte{}, err
		}
		err = mw.SetCompressionQuality(uint(quality * 100))
		if err != nil {
			return []byte{}, err
		}
	} else {
		err = mw.SetOption("webp:lossless", "true")
		if err != nil {
			return []byte{}, err
		}
		err = mw.SetCompressionQuality(80)
		if err != nil {
			return []byte{}, err
		}
	}
	mw.ResetIterator()
	return mw.GetImageBlob()
}
