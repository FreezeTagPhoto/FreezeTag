package formats

// this file and the errors returned from it will be missing coverage,
// because they mostly deal with MagickWand image conversions
// that have nothing to do with our stuff and should realistically not fail

import (
	"gopkg.in/gographics/imagick.v3/imagick"
)

// This function works to extract RGBA pixel information from
// any image type that ImageMagick supports in blob format
func imageToRGBA(mw *imagick.MagickWand) ([]byte, error) {
	rgba := mw.Clone()
	defer rgba.Destroy()
	err := rgba.SetImageFormat("RGBA")
	if err != nil {
		return []byte{}, err
	}
	err = rgba.SetImageDepth(8)
	if err != nil {
		return []byte{}, err
	}
	rgba.ResetIterator()
	return rgba.GetImageBlob()
}
