package forms

func uniqifyInPlace[T comparable](items []T) []T {
	n := len(items)
	set := make(map[T]struct{}, n)
	o := 0
	for i, item := range items {
		if _, found := set[item]; found {
			continue
		}
		set[item] = struct{}{}
		if o < i {
			items[o] = items[i]
		}
		o++
	}
	return items[:o]
}
