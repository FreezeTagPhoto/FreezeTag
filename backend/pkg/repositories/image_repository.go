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
	quality    = float32(0.5)

	max_height_large = 2160
	quality_large    = float32(0.8)
)

type ImageUploadSuccess struct {
	ID       database.ImageID `json:"id"`
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
	ID     database.ImageID `json:"id,omitempty"`
}

type ImageTagSuccess struct {
	ID    database.ImageID `json:"id,omitempty"`
	Count int              `json:"count"`
}

type ImageTagResult struct {
	Success *ImageTagSuccess `json:"success"`
	Err     *ImageTagFail    `json:"error"`
}

type TagResult struct {
	Success bool   `json:"success"`
	Err     string `json:"reason,omitempty"`
}

type ImageRepository interface {
	// a userID of 0 is used for unauthenticated requests
	SearchImage(query queries.DatabaseQuery, userId database.UserID) ([]database.ImageID, error)
	SearchImageOrdered(query queries.DatabaseQuery, field queries.SortField, order queries.SortOrder, userId database.UserID) ([]database.ImageID, error)
	SearchImageOrderedPaged(query queries.DatabaseQuery, field queries.SortField, order queries.SortOrder, pageSize uint, page uint, userId database.UserID) ([]database.ImageID, error)
	GetQueryTagCounts(query queries.DatabaseQuery, userId database.UserID) (map[string]int64, error)

	StoreImageBytes(data []byte, filename string) (database.ImageID, error)
	RetrieveThumbnail(id database.ImageID, quality uint) ([]byte, error)
	RetrieveImageFile(id database.ImageID) ([]byte, error)
	RetrieveAllTags() (map[string]int64, error)
	RetrieveImageTags(id database.ImageID) ([]string, error)
	AddTags(tags []string) TagResult
	AddImageTags(id database.ImageID, tags []string) ImageTagResult
	RemoveImageTags(id database.ImageID, tags []string) ImageTagResult
	DeleteTags(tags []string) (int, error)
	DeleteImage(id database.ImageID) (string, error)
	GetImageFilepath(id database.ImageID) (string, error)
	GetImageMetadata(id database.ImageID) (imagedata.Metadata, error)
	GetImageResolution(id database.ImageID) (int, int, error)
	GetTagCounts(ids []database.ImageID) (map[string]int64, error)
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
func (repo *DefaultImageRepository) StoreImageBytes(data []byte, filename string) (database.ImageID, error) {
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

func (repo *DefaultImageRepository) RetrieveThumbnail(id database.ImageID, quality uint) ([]byte, error) {
	return repo.db.GetImageThumbnail(id, quality)
}

func (repo *DefaultImageRepository) RetrieveImageFile(id database.ImageID) ([]byte, error) {
	filepath, err := repo.GetImageFilepath(id)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (repo *DefaultImageRepository) SearchImage(query queries.DatabaseQuery, userId database.UserID) ([]database.ImageID, error) {
	return repo.db.GetImages(query, userId)
}

func (repo *DefaultImageRepository) SearchImageOrdered(query queries.DatabaseQuery, field queries.SortField, order queries.SortOrder, userId database.UserID) ([]database.ImageID, error) {
	return repo.db.GetImagesOrder(query, field, order, userId)
}

func (repo *DefaultImageRepository) SearchImageOrderedPaged(query queries.DatabaseQuery, field queries.SortField, order queries.SortOrder, pageSize uint, pageNo uint, userId database.UserID) ([]database.ImageID, error) {
	return repo.db.GetImagesOrderPaged(query, field, order, pageSize, pageNo, userId)
}

func (repo *DefaultImageRepository) RetrieveAllTags() (map[string]int64, error) {
	return repo.db.GetAllTags()
}

func (repo *DefaultImageRepository) RetrieveImageTags(id database.ImageID) ([]string, error) {
	return repo.db.GetImageTags(id)
}

func (repo *DefaultImageRepository) AddTags(tags []string) TagResult {
	_, err := repo.db.AddTags(tags)
	if err != nil {
		return TagResult{
			Success: false,
			Err:     err.Error(),
		}
	}
	return TagResult{
		Success: true,
		Err:     "",
	}
}

func (repo *DefaultImageRepository) AddImageTags(id database.ImageID, tags []string) ImageTagResult {
	count, err := repo.db.AddImageTags(id, tags)
	if err != nil {
		return ImageTagResult{
			Success: nil,
			Err: &ImageTagFail{
				ID:     id,
				Reason: err.Error(),
			},
		}
	}
	return ImageTagResult{
		Success: &ImageTagSuccess{
			ID:    id,
			Count: count,
		},
		Err: nil,
	}
}

func (repo *DefaultImageRepository) RemoveImageTags(id database.ImageID, tags []string) ImageTagResult {
	count, err := repo.db.RemoveImageTags(id, tags)
	if err != nil {
		return ImageTagResult{
			Success: nil,
			Err: &ImageTagFail{
				ID:     id,
				Reason: err.Error(),
			},
		}
	}
	return ImageTagResult{
		Success: &ImageTagSuccess{
			ID:    id,
			Count: count,
		},
		Err: nil,
	}
}

func (repo *DefaultImageRepository) DeleteTags(tags []string) (int, error) {
	return repo.db.RemoveTags(tags)
}

func (repo *DefaultImageRepository) GetImageFilepath(id database.ImageID) (string, error) {
	fileName, err := repo.db.GetImageFile(id)
	if err != nil {
		return "", err
	}
	if fileName == nil || *fileName == "" {
		return "", fmt.Errorf("nil or empty file")
	}

	return path.Join(repo.folderPath, *fileName), nil
}

func (repo *DefaultImageRepository) DeleteImage(id database.ImageID) (string, error) {
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

func (repo *DefaultImageRepository) GetImageMetadata(id database.ImageID) (imagedata.Metadata, error) {
	metadata, err := repo.db.GetImageMetadata(id)
	if err != nil {
		return imagedata.Metadata{}, err
	}
	return metadata, nil
}

func (repo *DefaultImageRepository) GetImageResolution(id database.ImageID) (w int, h int, err error) {
	w, h, err = repo.db.GetImageResolution(id)
	return
}

func (repo *DefaultImageRepository) GetTagCounts(ids []database.ImageID) (map[string]int64, error) {
	return repo.db.GetTagCounts(ids)
}

func (repo *DefaultImageRepository) GetQueryTagCounts(query queries.DatabaseQuery, userId database.UserID) (map[string]int64, error) {
	images, err := repo.SearchImage(query, userId)
	if err != nil {
		return nil, err
	}
	return repo.db.GetTagCounts(images)
}
