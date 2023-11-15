package mvphelpers

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
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
	case string:
		if str, ok := strings.CutSuffix(v, "px"); ok {
			v, err := strconv.ParseFloat(str, 64)
			if err != nil {
				panic(fmt.Errorf("expected a number, got %T %v", v, v))
			}
			return v
		}
	}

	vv := reflect.ValueOf(v)
	if vv.CanFloat() {
		return vv.Float()
	} else if vv.CanInt() {
		return float64(vv.Int())
	} else if vv.CanUint() {
		return float64(vv.Uint())
	}
	panic(fmt.Errorf("expected a number, got %T %v", v, v))
}

func FuzzyBool(value any) bool {
	switch value := value.(type) {
	case nil:
		return false
	case time.Time:
		return !value.IsZero()
	case bool:
		return value
	case string:
		return len(value) > 0
	}
	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.Invalid:
		return false
	case reflect.String:
		return val.Len() != 0
	case reflect.Pointer:
		return !val.IsNil()
	case reflect.Map:
		return val.Len() != 0
	case reflect.Slice:
		return val.Len() != 0
	case reflect.Bool:
		return val.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return val.Uint() != 0
	default:
		return true
	}
}

// FuzzyHumanBool is similar to strconv.ParseBool, but recognizes more on/off.
func FuzzyHumanBool(str string) (bool, error) {
	switch str {
	case "1", "t", "T", "true", "TRUE", "True", "on", "ON", "On", "y", "Y", "yes", "YES", "Yes":
		return true, nil
	case "0", "f", "F", "false", "FALSE", "False", "off", "OFF", "Off", "n", "N", "no", "NO", "No":
		return false, nil
	}
	return false, fmt.Errorf("invalid bool value %q", str)
}
