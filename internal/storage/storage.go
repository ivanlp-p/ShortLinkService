package storage

type Storage interface {
	Set(id string, url string)
	Get(id string) (string, bool)
}
