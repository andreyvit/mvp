package mvphelpers

import (
	"encoding/json"
	"fmt"
	"html/template"
)

func List(v ...any) []any {
	return v
}

func Dict(args ...any) map[string]any {
	n := len(args)
	m := make(map[string]any, n/2)
	for i := 0; i < n; i++ {
		switch arg := args[i].(type) {
		case string:
			if i+1 >= n {
				panic(fmt.Errorf("string key %q not followed by value", arg))
			}
			m[arg] = args[i+1]
			i++
		case map[string]any:
			for k, v := range arg {
				m[k] = v
			}
		case map[string]string:
			for k, v := range arg {
				m[k] = v
			}
		default:
			panic(fmt.Errorf("argument %d must be a string or a map, got %T: %v", i, arg, arg))
		}
	}
	return m
}

func StrDict(args ...any) map[string]string {
	n := len(args)
	m := make(map[string]string, n/2)
	for i := 0; i < n; i++ {
		switch arg := args[i].(type) {
		case string:
			if i+1 >= n {
				panic(fmt.Errorf("string key %q not followed by value", arg))
			}
			m[arg] = fmt.Sprint(args[i+1])
			i++
		case map[string]string:
			for k, v := range arg {
				m[k] = v
			}
		default:
			panic(fmt.Errorf("argument %d must be a string or a map, got %T: %v", i, arg, arg))
		}
	}
	return m
}

func JSONDict(args ...any) template.JS {
	return template.JS(must(json.Marshal(Dict(args...))))
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
