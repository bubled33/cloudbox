package domain

type Storage interface {
	UploadFile(key string, data []byte) error
	DonwnloadFile(key string) ([]byte, error)
	DeleteFile(key string) error
}
