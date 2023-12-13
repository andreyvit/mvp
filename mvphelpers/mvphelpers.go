package mvphelpers

import (
	"fmt"
	"html/template"
	"log"
	"runtime/debug"
	"strings"
)

func FuncMap() template.FuncMap {
	return template.FuncMap{
		"in_groups_of": InGroupsOf,
		"is_even":      IsEven,
		"is_odd":       IsOdd,
		"dict":         Dict,
		"jsondict":     JSONDict,
		"list":         List,
		"subst":        Subst,
	}
}

func Subst(str string, keysAndValues ...any) string {
	n := len(keysAndValues)
	if n%2 != 0 {
		panic(fmt.Errorf("odd number of arguments %d: %v", n, keysAndValues))
	}
	for i := 0; i < n; i += 2 {
		key, value := keysAndValues[i], keysAndValues[i+1]
		if keyStr, ok := key.(string); ok {
			str = strings.ReplaceAll(str, keyStr, Stringify(value))
		} else {
			panic(fmt.Errorf("argument %d must be a string, got %T: %v", i, key, key))
		}
	}
	return str
}

// ExposeHelperPanic helps to debug panics inside view helpers.
// Add the following call at the start of a panicing helper:
//
//	defer func() { mvphelpers.ExposeHelperPanic(recover()) }()
func ExposeHelperPanic(e any) {
	if e != nil {
		log.Printf("** ERROR: helper panic: %v", e)
		debug.PrintStack()
		panic(e)
	}
}

func IsEven(n any) bool {
	return FuzzyInt(n)%2 == 0
}

func IsOdd(n any) bool {
	return FuzzyInt(n)%2 == 1
}

func InGroupsOf(n int, list any) []Group {
	allItems := FuzzyList(list)
	count := (len(allItems) + n - 1) / n
	result := make([]Group, count)
	for i := 0; i < count; i++ {
		result[i].Index = i
		result[i].GroupSize = n
		result[i].GroupCount = count
		if len(allItems) > n {
			result[i].Items, allItems = allItems[:n], allItems[n:]
		} else {
			result[i].Items = allItems
		}
	}
	return result
}

type Group struct {
	Index      int
	GroupSize  int
	GroupCount int
	Items      []any
}

func (group Group) IsFirst() bool {
	return group.Index == 0
}

func (group Group) IsLast() bool {
	return group.Index == group.GroupCount-1
}

func (group Group) PlaceholderCount() int {
	return group.GroupSize - len(group.Items)
}

func (group Group) Placeholders() []struct{} {
	return make([]struct{}, group.PlaceholderCount())
}

func Stringify(v any) string {
	switch v := v.(type) {
	case nil:
		return ""
	case string:
		return v
	case template.HTML:
		return string(v)
	default:
		return fmt.Sprint(v)
	}
}
