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

func (t *testItem) SetID(id uint64) {
	t.ID = id
}

type testService struct {
}

func (t testService) Unmarsall(data []byte) (Item, error) {
	var item testItem

	err := json.Unmarshal(data, &item)
	return &item, err
}

func (t testService) Check(old, new Item) []error {
	var value *testItem
	if new != nil {
		value = new.(*testItem)
	} else {
		value = old.(*testItem)
	}

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
	if page == 2 {
		return nil, 0, errors.New("error while reading")
	} else if page == 3 {
		return nil, 10, nil
	} else {
		return []Item{
			&testItem{ID: 1, Name: "First"},
			&testItem{ID: 2, Name: "First"},
		}, 10, nil
	}
}

func (t testService) Get(ctx context.Context, ID uint64) (Item, error) {
	if ID == 8000 || ID == 6000 || ID == 7000 {
		return &testItem{
			ID:   ID,
			Name: "Test",
		}, nil
	} else if ID == 4000 {
		return nil, errors.New("error with database")
	} else if ID == 2000 {
		return nil, ErrNotFound
	}
	return nil, nil
}

func (t testService) Create(ctx context.Context, o Item) (Item, uint64, error) {
	value := o.(*testItem)

	if value.Name == "error" {
		return nil, 0, errors.New("error while creating")
	}

	return &testItem{
		ID:   1,
		Name: value.Name,
	}, 1, nil
}

func (t testService) Update(ctx context.Context, o Item) (Item, error) {
	value := o.(*testItem)

	if value.Name == "error" {
		return nil, errors.New("error while updating")
	}

	return o, nil
}

func (t testService) Delete(ctx context.Context, o Item) error {
	value := o.(*testItem)

	if value.ID == 8000 {
		return errors.New("error while deleting")
	}

	return nil
}
