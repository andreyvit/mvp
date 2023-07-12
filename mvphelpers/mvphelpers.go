package mvphelpers

import "html/template"

func FuncMap() template.FuncMap {
	return template.FuncMap{
		"in_groups_of": InGroupsOf,
		"is_even":      IsEven,
		"is_odd":       IsOdd,
		"dict":         Dict,
		"list":         List,
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
