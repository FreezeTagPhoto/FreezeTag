package formats

import "fmt"

type failedConversionError struct {
	file  string
	cause error
}

func (f failedConversionError) Error() string {
	return fmt.Sprintf("failed to convert %v to RGBA: %v", f.file, f.cause)
}

type exifError struct {
	cause error
}

func (f exifError) Error() string {
	return fmt.Sprintf("failed to parse EXIF metadata: %v", f.cause)
}
