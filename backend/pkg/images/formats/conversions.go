package formats

// this file and the errors returned from it will be missing coverage,
// because they mostly deal with MagickWand image conversions
// that have nothing to do with our stuff and should realistically not fail

import (
	"path"

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

// This function works to load into memory
// most camera raw formats (HEIF, DNG, CR2-3, NEF, ARW, etc)
//
// BE CAREFUL: many camera raw formats will have crazy resolutions,
// this will eat all of RAM if used too frequently
func loadRawImage(name string, data []byte, mw *imagick.MagickWand) error {
	err := mw.SetFormat(path.Ext(name))
	if err != nil {
		return err
	}
	err = mw.ReadImageBlob(data)
	if err != nil {
		return err
	}
	return nil
}
