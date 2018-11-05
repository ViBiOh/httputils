package crud

// Item describe item
type Item interface {
	ID() string
}

// ItemService retrieves item
type ItemService interface {
	Empty() Item
	List(page, pageSize uint, sortKey string, sortAsc bool, filters map[string][]string) ([]Item, error)
	Get(ID string) (Item, error)
	Create(o Item) (Item, error)
	Update(ID string, o Item) (Item, error)
	Delete(ID string) error
}
