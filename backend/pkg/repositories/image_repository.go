package repositories

import (
	"fmt"
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/database/queries"
	"freezetag/backend/pkg/images"
	"freezetag/backend/pkg/images/imagedata"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

const (
	max_height = 512
	quality    = float32(0)

	max_height_large = 0
	quality_large    = float32(1)
)

type ImageUploadSuccess struct {
	Id       database.ImageId `json:"id"`
	Filename string           `json:"filename"`
}

type ImageUploadFailure struct {
	Reason   string `json:"reason"`
	Filename string `json:"filename"`
}

type UploadResult struct {
	Success *ImageUploadSuccess `json:"success"`
	Err     *ImageUploadFailure `json:"error"`
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
	SearchImageOrdered(query queries.DatabaseQuery, field queries.SortField, order queries.SortOrder) ([]database.ImageId, error)
	SearchImageOrderedPaged(query queries.DatabaseQuery, field queries.SortField, order queries.SortOrder, pageSize uint, page uint) ([]database.ImageId, error)
	StoreImageBytes(data []byte, filename string) (database.ImageId, error)
	RetrieveThumbnail(id database.ImageId, quality uint) ([]byte, error)
	RetrieveAllTags() (map[string]int64, error)
	RetrieveImageTags(id database.ImageId) ([]string, error)
	AddImageTags(id database.ImageId, tags []string) ImageTagResult
	RemoveImageTags(id database.ImageId, tags []string) ImageTagResult
	DeleteTags(tags []string) (int, error)
	DeleteImage(id database.ImageId) (string, error)
	GetImageFilepath(id database.ImageId) (string, error)
	GetImageMetadata(id database.ImageId) (imagedata.Metadata, error)
	GetImageResolution(id database.ImageId) (int, int, error)
	GetTagCounts(ids []database.ImageId) (map[string]int64, error)
	GetQueryTagCounts(query queries.DatabaseQuery) (map[string]int64, error)
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
		folderPath: strings.TrimRight(folderPath, "/") + "/",
	}
}

func (repo *DefaultImageRepository) safeFilePath(path string) (string, error) {
	suffix, err := repo.db.GetNonOverlappingSuffix(path)
	if err != nil {
		return "", err
	}
	if suffix != 0 {
		ext := filepath.Ext(path)
		base := strings.TrimSuffix(path, ext)
		return fmt.Sprintf("%s%d%s", base, suffix, ext), nil
	}
	return path, nil
}

// required so unique names from the database remain unique
var namingMutex sync.Mutex

// errors and results are given using the simple filename,
// the full filepath after the repo base folder is given to the database (this allows things to be moved around)
func (repo *DefaultImageRepository) StoreImageBytes(data []byte, filename string) (database.ImageId, error) {
	imagedata, err := repo.parser.ParseImage(filename, data)
	if err != nil {
		return 0, err
	}

	thumbSmall, err := images.CreateThumbnail(imagedata, max_height, quality)
	if err != nil {
		return 0, err
	}

	thumbLarge, err := images.CreateThumbnail(imagedata, max_height_large, quality_large)
	if err != nil {
		return 0, err
	}

	namingMutex.Lock()
	defer namingMutex.Unlock()
	filepath, err := repo.safeFilePath(filename)
	if err != nil {
		return 0, err
	}

	id, err := repo.db.AddImage(filepath, imagedata)
	if err != nil {
		return 0, err
	}

	ok, err := repo.db.AddImageThumbnail(id, 1, thumbSmall)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, fmt.Errorf("database returned false when adding thumbnail")
	}

	ok, err = repo.db.AddImageThumbnail(id, 2, thumbLarge)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, fmt.Errorf("database returned false when adding thumbnail")
	}

	// 0644 is rw-r--r-- permissions for this new file
	// 0755 is rwxr-xr-x permissions for this new directory (if it doesn't exist)
	if err := os.MkdirAll(repo.folderPath, 0755); err != nil {
		return 0, err
	}
	if err := os.WriteFile(path.Join(repo.folderPath, filepath), data, 0644); err != nil {
		return 0, err
	}
	return id, nil
}

func (repo *DefaultImageRepository) RetrieveThumbnail(id database.ImageId, quality uint) ([]byte, error) {
	return repo.db.GetImageThumbnail(id, quality)
}

func (repo *DefaultImageRepository) SearchImage(query queries.DatabaseQuery) ([]database.ImageId, error) {
	return repo.db.GetImages(query)
}

func (repo *DefaultImageRepository) SearchImageOrdered(query queries.DatabaseQuery, field queries.SortField, order queries.SortOrder) ([]database.ImageId, error) {
	return repo.db.GetImagesOrder(query, field, order)
}

func (repo *DefaultImageRepository) SearchImageOrderedPaged(query queries.DatabaseQuery, field queries.SortField, order queries.SortOrder, pageSize uint, pageNo uint) ([]database.ImageId, error) {
	return repo.db.GetImagesOrderPaged(query, field, order, pageSize, pageNo)
}

func (repo *DefaultImageRepository) RetrieveAllTags() (map[string]int64, error) {
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

func (repo *DefaultImageRepository) DeleteTags(tags []string) (int, error) {
	return repo.db.RemoveTags(tags)
}

func (repo *DefaultImageRepository) GetImageFilepath(id database.ImageId) (string, error) {
	fileName, err := repo.db.GetImageFile(id)
	if err != nil {
		return "", err
	}
	if fileName == nil || *fileName == "" {
		return "", fmt.Errorf("nil or empty file")
	}

	return path.Join(repo.folderPath, *fileName), nil
}

func (repo *DefaultImageRepository) DeleteImage(id database.ImageId) (string, error) {
	fileName, err := repo.GetImageFilepath(id)
	if err != nil {
		return "", err
	}
	_, err = repo.db.RemoveImage(id)
	if err != nil {
		return "", err
	}
	err = os.Remove(fileName)
	if err != nil {
		log.Printf("[ERR]  after deleting image id %d the file could not be deleted: %v", id, err)
		log.Printf("[ERR]  file possibly remaining after deletion: %v", fileName)
		return "", err
	}
	return fileName, nil
}

func (repo *DefaultImageRepository) GetImageMetadata(id database.ImageId) (imagedata.Metadata, error) {
	metadata, err := repo.db.GetImageMetadata(id)
	if err != nil {
		return imagedata.Metadata{}, err
	}
	return metadata, nil
}

func (repo *DefaultImageRepository) GetImageResolution(id database.ImageId) (w int, h int, err error) {
	w, h, err = repo.db.GetImageResolution(id)
	return
}

func (repo *DefaultImageRepository) GetTagCounts(ids []database.ImageId) (map[string]int64, error) {
	return repo.db.GetTagCounts(ids)
}

func (repo *DefaultImageRepository) GetQueryTagCounts(query queries.DatabaseQuery) (map[string]int64, error) {
	images, err := repo.SearchImage(query)
	if err != nil {
		return nil, err
	}
	return repo.db.GetTagCounts(images)
}
