package mvphelpers

import (
	"fmt"
	"reflect"
)

func FuzzyList(v any) []any {
	if list, ok := v.([]any); ok {
		return list
	}

	vv := reflect.ValueOf(v)
	switch vv.Kind() {
	case reflect.Slice, reflect.Array:
		n := vv.Len()
		result := make([]any, n)
		for i := 0; i < n; i++ {
			result[i] = vv.Index(i).Interface()
		}
		return result
	}
	panic(fmt.Errorf("expected a list (slice or array), got %T %v", v, v))
}

func FuzzyInt(v any) int {
	r, ok := TryFuzzyInt(v)
	if !ok {
		panic(fmt.Errorf("expected an integer, got %T %v", v, v))
	}
	return r
}

func TryFuzzyInt(v any) (int, bool) {
	switch v := v.(type) {
	case int:
		return v, true
	case int8:
		return int(v), true
	case int16:
		return int(v), true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case uint8:
		return int(v), true
	case uint16:
		return int(v), true
	case uint32:
		return int(v), true
	case uint64:
		return int(v), true
	}

	vv := reflect.ValueOf(v)
	if vv.CanInt() {
		return int(vv.Int()), true
	} else if vv.CanUint() {
		return int(vv.Uint()), true
	}
	return 0, false
}

func FuzzyInt64(v any) int64 {
	r, ok := TryFuzzyInt64(v)
	if !ok {
		panic(fmt.Errorf("expected an integer, got %T %v", v, v))
	}
	return r
}
func TryFuzzyInt64(v any) (int64, bool) {
	switch v := v.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		return int64(v), true
	}

	vv := reflect.ValueOf(v)
	if vv.CanInt() {
		return vv.Int(), true
	} else if vv.CanUint() {
		return int64(vv.Uint()), true
	}
	return 0, false
}

func FuzzyFloat64(v any) float64 {
	switch v := v.(type) {
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return float64(v)
	}

	vv := reflect.ValueOf(v)
	if vv.CanFloat() {
		return vv.Float()
	} else if vv.CanInt() {
		return float64(vv.Int())
	} else if vv.CanUint() {
		return float64(vv.Uint())
	}
	panic(fmt.Errorf("expected an integer, got %T %v", v, v))
}

func FuzzyBool(value any) bool {
	if value == nil || value == "" || value == false {
		return false
	}
	return true
}
