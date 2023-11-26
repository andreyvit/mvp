package mvpmetrics

import (
	"slices"
	"strings"
	"sync"
)

type vector[T any] struct {
	values       []T
	labels       [][]string
	lock         sync.RWMutex
	labelIndices map[string]int
}

func (vec *vector[T]) enum(f func(labels []string, value T)) {
	vec.lock.RLock()
	defer vec.lock.RUnlock()
	for i, value := range vec.values {
		f(vec.labels[i], value)
	}
}

func (vec *vector[T]) acquire(labels []string) (*T, bool) {
	hash := strings.Join(labels, "|") // TODO: maybe use an actual hash

	vec.lock.RLock()
	i, ok := vec.labelIndices[hash]
	if ok {
		return &vec.values[i], false
	}
	vec.lock.RUnlock()
	vec.lock.Lock()

	if vec.labelIndices == nil {
		vec.values = make([]T, 0, 10)
		vec.labels = make([][]string, 0, 10)
		vec.labelIndices = make(map[string]int, 10)
	}

	i, ok = vec.labelIndices[hash]
	if ok {
		return &vec.values[i], true
	}

	i = len(vec.values)
	vec.labelIndices[hash] = i

	if i+1 > cap(vec.values) {
		vec.values = doublecap(vec.values)
		vec.labels = doublecap(vec.labels)
	}
	var zero T
	vec.values = append(vec.values, zero)
	vec.labels = append(vec.labels, slices.Clone(labels))
	return &vec.values[i], true
}

func (vec *vector[T]) release(isWriteLocked bool) {
	if isWriteLocked {
		vec.lock.Unlock()
	} else {
		vec.lock.RUnlock()
	}
}

func doublecap[S ~[]T, T any](slice S) S {
	result := make(S, len(slice), cap(slice)*2)
	copy(result, slice)
	return result
}
