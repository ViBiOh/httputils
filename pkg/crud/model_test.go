package crud

import (
	"context"
	"encoding/json"
)

type testItem struct {
	ID   uint64
	Name string
}

func (t testItem) SetID(id uint64) {
	t.ID = id
}

type testService struct {
}

func (t testService) Unmarsall(data []byte) (Item, error) {
	var item testItem

	err := json.Unmarshal(data, &item)
	return item, err
}

func (t testService) List(ctx context.Context, page, pageSize uint, sortKey string, sortDesc bool, filters map[string][]string) ([]Item, uint, error) {
	return nil, 0, nil
}

func (t testService) Get(ctx context.Context, ID uint64) (Item, error) {
	return nil, nil
}

func (t testService) Create(ctx context.Context, o Item) (Item, error) {
	return nil, nil
}

func (t testService) Update(ctx context.Context, o Item) (Item, error) {
	return nil, nil
}

func (t testService) Delete(ctx context.Context, o Item) error {
	return nil
}
