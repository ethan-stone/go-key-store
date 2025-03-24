package service

type GetResult struct {
	Ok  bool
	Val string
}

type StoreService interface {
	Get(key string) (*GetResult, error)
	Put(key string, val string) error
	Delete(key string) error
}
