package mvphelpers

import (
	"fmt"
)

func List(v ...any) []any {
	return v
}

func Dict(args ...any) map[string]any {
	n := len(args)
	if n%2 != 0 {
		panic(fmt.Errorf("odd number of arguments %d: %v", n, args))
	}
	m := make(map[string]any, n/2)
	for i := 0; i < n; i += 2 {
		key, value := args[i], args[i+1]
		if keyStr, ok := key.(string); ok {
			m[keyStr] = value
		} else {
			panic(fmt.Errorf("argument %d must be a string, got %T: %v", i, key, key))
		}
	}
	return m
}
