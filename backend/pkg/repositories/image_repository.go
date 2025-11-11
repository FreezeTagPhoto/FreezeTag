package repositories

type ImageHandleSuccess struct {
	Id       uint   `json:"id"`
	Filename string `json:"filename"`
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
	RetrieveImage(id uint) (any, error)
}
