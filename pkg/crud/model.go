package crud

import "context"

// Service retrieves item
type Service interface {
	Unmarsall(data []byte) (interface{}, error)
	Check(ctx context.Context, old, new interface{}) []error
	List(ctx context.Context, page, pageSize uint, sortKey string, sortDesc bool, filters map[string][]string) ([]interface{}, uint, error)
	Get(ctx context.Context, ID uint64) (interface{}, error)
	Create(ctx context.Context, o interface{}) (interface{}, error)
	Update(ctx context.Context, o interface{}) (interface{}, error)
	Delete(ctx context.Context, o interface{}) error
}
