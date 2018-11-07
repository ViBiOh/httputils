package crud

import "context"

// Item describe item
type Item interface {}

// ItemService retrieves item
type ItemService interface {
	Unmarsall([]byte) (Item, error)
	List(ctx context.Context, page, pageSize uint, sortKey string, sortAsc bool, filters map[string][]string) ([]Item, uint, error)
	Get(ctx context.Context, ID string) (Item, error)
	Create(ctx context.Context, o Item) (Item, error)
	Update(ctx context.Context, o Item) (Item, error)
	Delete(ctx context.Context, o Item) error
}
