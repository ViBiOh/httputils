package memory

import "time"

type Item[K comparable] struct {
	id         K
	expiration time.Time
}

type ExpirationQueue[K comparable] []Item[K]

func (eq ExpirationQueue[K]) Len() int {
	return len(eq)
}

func (eq ExpirationQueue[K]) Less(i, j int) bool {
	return eq[i].expiration.Before(eq[j].expiration)
}

func (eq ExpirationQueue[K]) Swap(i, j int) {
	eq[i], eq[j] = eq[j], eq[i]
}

func (eq ExpirationQueue[K]) Index(id K) int {
	for index, item := range eq {
		if item.id == id {
			return index
		}
	}

	return -1
}

func (eq *ExpirationQueue[K]) Push(x any) {
	*eq = append(*eq, x.(Item[K]))
}

func (eq *ExpirationQueue[K]) Pop() any {
	old := *eq

	n := len(old)
	item := old[n-1]
	*eq = old[0 : n-1]

	return item
}
