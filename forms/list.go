package forms

import (
	"fmt"
	"html/template"
	"log"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/exp/maps"
)

var (
	debugLogList = false
)

type List[T comparable] struct {
	RenderableImpl[List[T]]
	Name     string
	Template string
	TemplateStyle
	Identity
	Styles     []*Style
	WrapperTag TagOpts
	*Binding[[]T]

	MinCount int
	MaxCount int

	UseIndicesAsKeys bool

	NewItem       func(name, typ string, index int) (T, bool)
	DeleteItem    func(item T)
	ItemName      func(item T, index int) string
	ItemType      func(item T, index int) string
	RenderItem    func(item T, index int) *Group
	RenderItemPtr func(item *T, index int) *Group
	Sort          func(items []T)
	Empty         Child
	TopArea       func() Children
	BottomArea    func() Children
	ItemActions   map[string]func(item T, index int, action *ItemAction)

	children   Children
	childFlags []ChildFlags
	items      []T
	InnerHTML  template.HTML
}

type ItemAction struct {
	ItemName string
	Key      string
	Remove   bool
	Save     bool
}

func (*List[T]) DefaultTemplate() string { return "list" }

func (list *List[T]) Finalize(state *State) {
	state.PushName(list.Name)
	state.PushStyles(list.Styles)
	state.AssignIdentity(&list.Identity)
	if list.FullName == "" {
		panic(fmt.Errorf("%T must have a name", list))
	}
	if list.Template == "" {
		list.Template = list.DefaultTemplate()
	}

	if list.children != nil {
		return
	}

	if list.UseIndicesAsKeys {
		if list.NewItem == nil {
			list.NewItem = func(name, typ string, index int) (T, bool) {
				var zero T
				return zero, true
			}
		}
		if list.ItemName == nil {
			list.ItemName = func(item T, index int) string {
				return strconv.Itoa(index)
			}
		}
	}

	if list.RenderItem == nil && list.RenderItemPtr == nil {
		panic("need either RenderItem or RenderItemPtr")
	}

	existingItems := list.Binding.Value()

	var items []T
	newlyAddedItems := make(map[T]struct{})
	if state.Data == nil {
		items = existingItems
	} else {
		newItemNameSet, action := list.computeListItemNames(state.Data)
		if debugLogList {
			log.Printf("forms: %T: newItemNameSet = %v", list, maps.Keys(newItemNameSet))
		}

		items = make([]T, 0, len(existingItems)+10)
		for i, item := range existingItems {
			itemName := list.FullItemName(item, i)
			if action != nil && action.ItemName == itemName {
				list.ItemActions[action.Key](item, i, action)
				if action.Remove {
					delete(newItemNameSet, itemName)
					continue
				}
			}
			if _, found := newItemNameSet[itemName]; found {
				items = append(items, item)
				delete(newItemNameSet, itemName)
			} else {
				if list.DeleteItem != nil {
					list.DeleteItem(item)
				}
			}
		}

		if list.NewItem != nil {
			addedItemNames := maps.Keys(newItemNameSet)
			sort.Strings(addedItemNames)
			for _, itemName := range addedItemNames {
				proposedName, typ := list.SplitFullItemName(itemName)
				if item, ok := list.NewItem(proposedName, typ, len(items)); ok {
					items = append(items, item)
					if debugLogList {
						log.Printf("forms: %T: added %q %q (full %q)", list, typ, proposedName, itemName)
					}
				} else {
					if debugLogList {
						log.Printf("forms: %T: failed to add %q %q", list, typ, proposedName)
					}
				}
			}

			addName := JoinNames(list.Identity.FullName, "add")
			if state.Data.Action == addName {
				if item, ok := list.NewItem("", "", len(items)); ok {
					items = append(items, item)
					newlyAddedItems[item] = struct{}{}
				}
			} else if suffix, ok := strings.CutPrefix(state.Data.Action, addName); ok {
				comps := SplitName(suffix)
				if len(comps) == 1 {
					typ := comps[0]
					if item, ok := list.NewItem("", typ, len(items)); ok {
						items = append(items, item)
						newlyAddedItems[item] = struct{}{}
					}
				}
			}
		}
	}
	if list.Sort != nil {
		list.Sort(items)
	}
	list.items = items

	children := make(Children, 0, len(items)+10)
	childFlags := make([]ChildFlags, 0, len(items)+10)
	if list.TopArea != nil {
		for _, child := range list.TopArea() {
			children = append(children, child)
			childFlags = append(childFlags, 0)
		}
	}
	for i, item := range items {
		var group *Group
		if list.RenderItem != nil {
			group = list.RenderItem(item, i)
		} else if list.RenderItemPtr != nil {
			group = list.RenderItemPtr(&items[i], i)
		}
		group.Name = list.FullItemName(item, i)
		var flags ChildFlags
		if _, found := newlyAddedItems[item]; found {
			flags = ChildFlagSkipProcessing
		}
		children = append(children, group)
		childFlags = append(childFlags, flags)
	}
	if list.BottomArea != nil {
		for _, child := range list.BottomArea() {
			children = append(children, child)
			childFlags = append(childFlags, 0)
		}
	}
	list.children = children
	list.childFlags = childFlags
}

func (list *List[T]) computeListItemNames(data *FormData) (map[string]struct{}, *ItemAction) {
	fullListName := list.Identity.FullName
	result := make(map[string]struct{})
	prefix := fullListName + "["
	var action *ItemAction

	for k := range data.Values {
		if suffix, ok := strings.CutPrefix(k, prefix); ok {
			if name, _, ok := strings.Cut(suffix, "]"); ok {
				removeAction := JoinNames(fullListName, name, "remove")
				isRemoving := (data.Action == removeAction)
				if !isRemoving {
					result[name] = struct{}{}

					for actionKey := range list.ItemActions {
						if data.Action == JoinNames(fullListName, name, actionKey) {
							action = &ItemAction{
								ItemName: name,
								Key:      actionKey,
							}
							break
						}
					}
				}

			}
		}
	}
	return result, action
}

func (list *List[T]) FullItemName(item T, index int) string {
	name := list.ItemName(item, index)
	if list.ItemType == nil {
		return name
	} else {
		return list.ItemType(item, index) + ":" + name
	}
}

func (list *List[T]) SplitFullItemName(fullName string) (name, typ string) {
	if list.ItemType == nil {
		return fullName, ""
	} else {
		typ, name, ok := strings.Cut(fullName, ":")
		if !ok {
			typ = ""
			name = ""
		}
		return name, typ
	}
}

func (*List[T]) EnumFields(f func(name string, field *Field)) {}

func (list *List[T]) EnumChildren(f func(Child, ChildFlags)) {
	for i, child := range list.children {
		f(child, list.childFlags[i])
	}
}

func (list *List[T]) ShouldProcessChild(data *FormData, child Child) bool {
	return false
}

func (list *List[T]) Process(fd *FormData) {
	if list.Sort != nil {
		list.Sort(list.items)
	}
	list.Binding.Set(list.items)
}

func (list *List[T]) RenderInto(buf *strings.Builder, r *Renderer) {
	list.InnerHTML = r.Render(list.children)
	r.RenderWrapperTemplateInto(buf, list.Template, list, list.InnerHTML)
}
