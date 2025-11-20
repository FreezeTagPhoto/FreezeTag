package repositories

import (
	"errors"
	"fmt"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/database/queries"
	"freezetag/backend/pkg/images"
	"os"
	"strconv"
)

type ImageUploadSuccess struct {
	Id       database.ImageId `json:"id"`
	Filename string           `json:"filename"`
}

type ImageUploadFail struct {
	Reason   string `json:"reason"`
	Filename string `json:"filename"`
}

type UploadResult struct {
	Success *ImageUploadSuccess `json:"success"`
	Err     *ImageUploadFail    `json:"error"`
}

type ImageTagFail struct {
	Reason string           `json:"reason"`
	Id     database.ImageId `json:"id"`
}

type ImageTagSuccess struct {
	Id    database.ImageId `json:"id"`
	Count int              `json:"count"`
}

type ImageTagResult struct {
	Success *ImageTagSuccess `json:"success"`
	Err     *ImageTagFail    `json:"error"`
}

type ImageRepository interface {
	SearchImage(query queries.DatabaseQuery) ([]database.ImageId, error)
	StoreImageBytes(data []byte, filename string) UploadResult
	RetrieveThumbnail(id database.ImageId, quality uint) ([]byte, error)
	RetrieveAllTags() ([]string, error)
	RetrieveImageTags(id database.ImageId) ([]string, error)
	AddImageTags(id database.ImageId, tags []string) ImageTagResult
	RemoveImageTags(id database.ImageId, tags []string) ImageTagResult
}

type DefaultImageRepository struct {
	db         database.ImageDatabase
	parser     images.Parser
	folderPath string
}

func InitImageRepository(db database.ImageDatabase, paser images.Parser, folderPath string) *DefaultImageRepository {
	return &DefaultImageRepository{
		db:         db,
		parser:     paser,
		folderPath: folderPath,
	}
}

func errorUploadResult(filename string, err error) UploadResult {
	return UploadResult{
		Success: nil,
		Err: &ImageUploadFail{
			Reason:   err.Error(),
			Filename: filename,
		},
	}
}

func safeFilePath(filepath, filename string) (string, error) {
	tmpName := filename
	for i := int64(1); ; i++ {
		_, err := os.Stat(filepath + "/" + tmpName)
		switch {
		case err == nil:
			tmpName = "copy " + strconv.FormatInt(i, 10) + " " + filename
			continue
		case errors.Is(err, os.ErrNotExist):
			return tmpName, nil
		default:
			return "", fmt.Errorf("failed to check file existance via os.Stat %q: %w", filename, err)
		}
	}
}

// errors and results are given using the simple filename,
// the full filepath (e.g /tmp/filename) is given to the database
func (repo *DefaultImageRepository) StoreImageBytes(data []byte, filename string) UploadResult {
	max_height := 512
	quality := float32(0)

	max_height_large := 0
	quality_large := float32(1)

	filename, err := safeFilePath(repo.folderPath, filename)
	if err != nil {
		return errorUploadResult(filename, err)
	}
	filepath := repo.folderPath + "/" + filename

	imagedata, err := repo.parser.ParseImage(filename, data)
	if err != nil {
		return errorUploadResult(filename, err)
	}
	thumbSmall, err := images.CreateThumbnail(imagedata, max_height, quality)
	if err != nil {
		return errorUploadResult(filename, err)
	}
	thumbLarge, err := images.CreateThumbnail(imagedata, max_height_large, quality_large)
	if err != nil {
		return errorUploadResult(filename, err)
	}

	id, err := repo.db.AddImage(filepath, imagedata)
	if err != nil {
		return errorUploadResult(filename, err)
	}

	ok, err := repo.db.AddImageThumbnail(id, 1, thumbSmall)
	if err != nil {
		return errorUploadResult(filename, err)
	}
	if !ok {
		return errorUploadResult(filename, fmt.Errorf("database returned false when adding thumbnail"))
	}

	ok, err = repo.db.AddImageThumbnail(id, 2, thumbLarge)
	if err != nil {
		return errorUploadResult(filename, err)
	}
	if !ok {
		return errorUploadResult(filename, fmt.Errorf("database returned false when adding thumbnail"))
	}

	// 0644 is rw-r--r-- permissions for this new file
	// 0755 is rwxr-xr-x permissions for this new directory (if it doesn't exist)
	if err := os.MkdirAll(repo.folderPath, 0755); err != nil {
		return errorUploadResult(filename, err)
	}
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return errorUploadResult(filename, err)
	}
	return UploadResult{
		Success: &ImageUploadSuccess{
			Id:       id,
			Filename: filename,
		},
		Err: nil,
	}
}

func (repo *DefaultImageRepository) RetrieveThumbnail(id database.ImageId, quality uint) ([]byte, error) {
	return repo.db.GetImageThumbnail(id, quality)
}

func (repo *DefaultImageRepository) SearchImage(query queries.DatabaseQuery) ([]database.ImageId, error) {
	return repo.db.GetImages(query)
}

func (repo *DefaultImageRepository) RetrieveAllTags() ([]string, error) {
	return repo.db.GetAllTags()
}

func (repo *DefaultImageRepository) RetrieveImageTags(id database.ImageId) ([]string, error) {
	return repo.db.GetImageTags(id)
}

func (repo *DefaultImageRepository) AddImageTags(id database.ImageId, tags []string) ImageTagResult {
	count, err := repo.db.AddImageTags(id, tags)
	if err != nil {
		return ImageTagResult{
			Success: nil,
			Err: &ImageTagFail{
				Id:     id,
				Reason: err.Error(),
			},
		}
	}
	return ImageTagResult{
		Success: &ImageTagSuccess{
			Id:    id,
			Count: count,
		},
		Err: nil,
	}
}

func (repo *DefaultImageRepository) RemoveImageTags(id database.ImageId, tags []string) ImageTagResult {
	count, err := repo.db.RemoveImageTags(id, tags)
	if err != nil {
		return ImageTagResult{
			Success: nil,
			Err: &ImageTagFail{
				Id:     id,
				Reason: err.Error(),
			},
		}
	}
	return ImageTagResult{
		Success: &ImageTagSuccess{
			Id:    id,
			Count: count,
		},
		Err: nil,
	}
}
