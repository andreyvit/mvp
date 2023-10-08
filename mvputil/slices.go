package mvputil

func FilterOut[S ~[]E, E comparable](s S, undesired E) S {
	i := 0
	for k := 0; k < len(s); k++ {
		if s[k] != undesired {
			if i != k {
				s[i] = s[k]
			}
			i++
		}
	}
	return s[:i]
}

func FilterFunc[S ~[]E, E comparable](s S, f func(E) bool) S {
	i := 0
	for k := 0; k < len(s); k++ {
		if f(s[k]) {
			if i != k {
				s[i] = s[k]
			}
			i++
		}
	}
	return s[:i]
}

func Prepend[S ~[]E, E any](s S, elems ...E) S {
	ns, ne := len(s), len(elems)
	result := make(S, ns+ne)
	copy(result, s)
	copy(result[ns:], elems)
	return result
}
