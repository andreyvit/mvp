package mvphelpers

import "html/template"

func FuncMap() template.FuncMap {
	return template.FuncMap{
		"in_groups_of": InGroupsOf,
	}
}

func InGroupsOf(n int, list []any) []Group {
	count := (len(list) + n - 1) / n
	result := make([]Group, count)
	for i := 0; i < count; i++ {
		result[i].Index = i
		result[i].GroupSize = n
		result[i].GroupCount = count
		if len(list) > n {
			result[i].Items, list = list[:n], list[n:]
		} else {
			result[i].Items = list
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
