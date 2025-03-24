package service

type StoreService interface {
	Get(key string) (string, error)
	Put(key string, val string) error
	Delete(key string) error
}
