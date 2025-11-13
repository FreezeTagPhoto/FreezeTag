package repositories

import (
	"freezetag/backend/pkg/database"
	"freezetag/backend/pkg/images"
)

type ImageHandleSuccess struct {
	Id       database.ImageId `json:"id"`
	Filename string           `json:"filename"`
}

type ImageHandleFail struct {
	Reason   error  `json:"reason"`
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
	db database.ImageDatabase
	parserCollection images.Parser
}

func InitImageRepository(db database.ImageDatabase, parserCollection images.Parser) *DefaultImageRepository {
	return &DefaultImageRepository{
		db: db,
		parserCollection: parserCollection,
	}
}

func errorResult(filename string, err error) Result {
	return Result{
		Success: nil,
		Err: &ImageHandleFail{
			Reason:   err,
			Filename: filename,
		},
	}
}

func (repo *DefaultImageRepository) StoreImageBytes(data []byte, filename string) Result {
	imagedata, err := repo.parserCollection.ParseImage(filename, data)
	if err != nil {
		return errorResult(filename, err)
	}

	id, err := repo.db.AddImage(filename, imagedata)
	if err != nil {
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
