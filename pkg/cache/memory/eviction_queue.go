package memory

import "time"

type Item[K comparable] struct {
	id         K
	expiration time.Time
	index      int
}

type ExpirationQueue[K comparable] struct {
	index map[K]int
	items []Item[K]
}

func NewExpirationQueue[K comparable]() *ExpirationQueue[K] {
	return &ExpirationQueue[K]{
		index: make(map[K]int),
	}
}

func (eq *ExpirationQueue[K]) Len() int {
	return len(eq.items)
}

func (eq *ExpirationQueue[K]) Less(i, j int) bool {
	return eq.items[i].expiration.Before(eq.items[j].expiration)
}

func (eq *ExpirationQueue[K]) Swap(i, j int) {
	eq.items[i], eq.items[j] = eq.items[j], eq.items[i]

	eq.items[i].index = i
	eq.items[j].index = j
	eq.index[eq.items[i].id] = i
	eq.index[eq.items[j].id] = j
}

func (eq *ExpirationQueue[K]) Index(id K) int {
	if idx, ok := eq.index[id]; ok {
		return idx
	}

	return -1
}

func (eq *ExpirationQueue[K]) Push(x any) {
	item := x.(Item[K])
	item.index = len(eq.items)
	eq.items = append(eq.items, item)
	eq.index[item.id] = item.index
}

func (eq *ExpirationQueue[K]) Pop() any {
	old := eq.items

	n := len(old)
	item := old[n-1]
	eq.items = old[0 : n-1]
	delete(eq.index, item.id)

	return item
}
