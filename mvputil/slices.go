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
