package repositories


type ImageHandleSuccess struct {
	Id uint
	Filename string
}

type ImageHandleFail struct {
	Reason error
	Filename string
}	

type Result struct {
	Success *ImageHandleSuccess
	Err     *ImageHandleFail
}

type ImageRepository interface {
	StoreImageBytes(data []byte, filename string) Result
	RetrieveImage(id uint) (any, error)
}
