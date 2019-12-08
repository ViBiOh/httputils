package crud

import (
	"context"
	"encoding/json"
	"errors"
)

type testItem struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
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

func (t testService) Check(item Item) []error {
	value := item.(testItem)

	if value.ID == 6000 {
		return []error{
			errors.New("invalid name"),
			errors.New("invalid value"),
		}
	} else if value.Name == "invalid" {
		return []error{
			errors.New("invalid name"),
		}
	}

	return nil
}

func (t testService) List(ctx context.Context, page, pageSize uint, sortKey string, sortDesc bool, filters map[string][]string) ([]Item, uint, error) {
	return nil, 0, nil
}

func (t testService) Get(ctx context.Context, ID uint64) (Item, error) {
	if ID == 8000 {
		return testItem{
			ID:   8000,
			Name: "Test",
		}, nil
	} else if ID == 4000 {
		return nil, errors.New("error with database")
	} else if ID == 2000 {
		return nil, ErrNotFound
	}
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
