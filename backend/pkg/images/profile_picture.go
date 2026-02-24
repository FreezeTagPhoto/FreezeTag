package images

import (
	"crypto/sha256"
	"fmt"
	"freezetag/backend/pkg/images/imagedata"
	"image"
	"image/color"

	"gopkg.in/gographics/imagick.v3/imagick"
)

var dimension = 256

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
	err = mw.ResizeImage(uint(dimension), uint(dimension), imagick.FILTER_LANCZOS)
	if err != nil {
		return nil, err
	}

	mw.SetImageFormat("WEBP")
	mw.StripImage()
	mw.SetImageCompressionQuality(uint(0.8 * 100))
	mw.SetOption("webp:lossless", "false")

	return mw.GetImageBlob()
}

func DefaultProfilePicture(filename string) ([]byte, error) {
    filehash := sha256.Sum256([]byte(filename))
    
	// use the first 3 bytes of the hash to create a unique color for the profile picture
    c1 := color.RGBA{filehash[0], filehash[1], filehash[2], 255}
    c2 := color.RGBA{255, 255, 255, 255}
    
    img := image.NewRGBA(image.Rect(0, 0, dimension, dimension))
    
    // To use all 256 bits symmetrically is impossible.
	// we can use 128 bits to create a 8x16 image, and mirror it to create a 32x32 image
	// This is why the numColumns is 8 and the block size is 16
    numColumns := 8
	blockSize := dimension / (numColumns * 2) // 16 blocks across, each block is 16 pixels wide

    for i := range 128 {
        byteIdx := i / 8
        bitIdx := i % 8
        
        activeColor := c2
        if (filehash[byteIdx] & (1 << bitIdx)) != 0 {
            activeColor = c1
        }

        // Standard 1D to 2D mapping math
        pixelX := (i % numColumns) * blockSize
        pixelY := (i / numColumns) * blockSize
		
		img.SetRGBA(pixelX, pixelY, activeColor)
		img.SetRGBA(dimension-1-pixelX, pixelY, activeColor)

        // Draw the block and its mirror on the opposite side of the image
        for y := pixelY; y < pixelY+blockSize; y++ {
            for x := pixelX; x < pixelX+blockSize; x++ {
                img.SetRGBA(x, y, activeColor)
                img.SetRGBA(dimension-1-x, y, activeColor)
            }
        }
    }
	mw := imagick.NewMagickWand()
	defer mw.Destroy()
	err := mw.ConstituteImage(uint(dimension), uint(dimension), "RGBA", imagick.PIXEL_CHAR, img.Pix)
	if err != nil {
		return nil, err
	}
	mw.SetImageFormat("WEBP")
	mw.StripImage()
	mw.SetImageCompressionQuality(uint(0.8 * 100))
	mw.SetOption("webp:lossless", "false")

	return mw.GetImageBlob()
}