package crud

import "context"

// Error describes a crud error
type Error struct {
	Field string `json:"field"`
	Label string `json:"label"`
}

// NewError creates a new error
func NewError(field, label string) Error {
	return Error{
		Field: field,
		Label: label,
	}
}

// Service retrieves item
type Service interface {
	Unmarshal(data []byte, contentType string) (interface{}, error)
	Check(ctx context.Context, old, new interface{}) []Error
	List(ctx context.Context, page, pageSize uint, sortKey string, sortDesc bool, filters map[string][]string) ([]interface{}, uint, error)
	Get(ctx context.Context, ID uint64) (interface{}, error)
	Create(ctx context.Context, o interface{}) (interface{}, error)
	Update(ctx context.Context, o interface{}) (interface{}, error)
	Delete(ctx context.Context, o interface{}) error
}
