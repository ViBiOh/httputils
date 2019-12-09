package crud

import "context"

// Item describe item
type Item interface {
	SetID(uint64)
}

// Service retrieves item
type Service interface {
	Unmarsall(data []byte) (Item, error)
	Check(old, new Item) []error
	List(ctx context.Context, page, pageSize uint, sortKey string, sortDesc bool, filters map[string][]string) ([]Item, uint, error)
	Get(ctx context.Context, ID uint64) (Item, error)
	Create(ctx context.Context, o Item) (Item, uint64, error)
	Update(ctx context.Context, o Item) (Item, error)
	Delete(ctx context.Context, o Item) error
}
