package repositories

import (
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/images"
	"os"
)

type ImageHandleSuccess struct {
	Id       database.ImageId `json:"id"`
	Filename string           `json:"filename"`
}

type ImageHandleFail struct {
	Reason   string `json:"reason"`
	Filename string `json:"filename"`
}

type Result struct {
	Success *ImageHandleSuccess `json:"success"`
	Err     *ImageHandleFail    `json:"error"`
}

type ImageRepository interface {
	StoreImageBytes(data []byte, filename string) Result
	RetrieveThumbnail(id database.ImageId, quality uint) ([]byte, error)
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

func errorResult(filename string, err error) Result {
	return Result{
		Success: nil,
		Err: &ImageHandleFail{
			Reason:   err.Error(),
			Filename: filename,
		},
	}
}

// TODO: Doesn't handle name collisions
func (repo *DefaultImageRepository) StoreImageBytes(data []byte, filename string) Result {
	max_height := 512
	quality := float32(0)

	imagedata, err := repo.parser.ParseImage(filename, data)
	if err != nil {
		return errorResult(filename, err)
	}
	thumbSmall, err := images.CreateThumbnail(imagedata, max_height, quality)
	if err != nil {
		return errorResult(filename, err)
	}
	thumbLarge, err := images.CreateThumbnail(imagedata, 0, 1)
	if err != nil {
		return errorResult(filename, err)
	}

	id, err := repo.db.AddImage(filename, imagedata)
	if err != nil {
		return errorResult(filename, err)
	}

	ok, err := repo.db.AddImageThumbnail(id, 1, thumbSmall)
	if err != nil || !ok {
		return errorResult(filename, err)
	}
	ok, err = repo.db.AddImageThumbnail(id, 2, thumbLarge)
	if err != nil || !ok {
		return errorResult(filename, err)
	}

	// 0644 is rw-r--r-- permissions for this new file
	// 0755 is rwxr-xr-x permissions for this new directory (if it doesn't exist)
	if err := os.MkdirAll(repo.folderPath, 0755); err != nil {
		return errorResult(filename, err)
	}
	if err := os.WriteFile(repo.folderPath+"/"+filename, data, 0644); err != nil {
		return errorResult(filename, err)
	}
	return Result{
		Success: &ImageHandleSuccess{
			Id:       id,
			Filename: filename,
		},
		Err: nil,
	}
}

func (repo *DefaultImageRepository) RetrieveThumbnail(id database.ImageId, quality uint) ([]byte, error) {
	return repo.db.GetImageThumbnail(id, quality)
}
