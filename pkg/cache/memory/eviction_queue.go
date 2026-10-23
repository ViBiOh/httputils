package memory

type expirationQueue[K comparable, V any] []*entry[K, V]

func (eq expirationQueue[K, V]) Len() int {
	return len(eq)
}

func (eq expirationQueue[K, V]) Less(first, second int) bool {
	return eq[first].expiration.Before(eq[second].expiration)
}

func (eq expirationQueue[K, V]) Swap(first, second int) {
	eq[first], eq[second] = eq[second], eq[first]

	eq[first].index = first
	eq[second].index = second
}

func (eq *expirationQueue[K, V]) Push(item any) {
	value := item.(*entry[K, V])

	value.index = len(*eq)
	*eq = append(*eq, value)
}

func (eq *expirationQueue[K, V]) Pop() any {
	old := *eq
	last := len(old) - 1

	value := old[last]
	old[last] = nil
	value.index = -1

	*eq = old[:last]

	return value
}
