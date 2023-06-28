package forms

import (
	"fmt"
	"html/template"
	"log"
	"strings"

	"golang.org/x/exp/maps"
)

var (
	debugLogList = true
)

type List[T any] struct {
	Name     string
	Template string
	TemplateStyle
	Identity
	Styles     []*Style
	WrapperTag TagOpts
	*Binding[[]T]

	MinCount int
	MaxCount int

	NewItem    func(name, typ string) (T, bool)
	DeleteItem func(item T)
	ItemName   func(item T) string
	ItemType   func(item T) string
	RenderItem func(item T) *Group
	Sort       func(items []T)
	Empty      Child
	TopArea    func() Children
	BottomArea func() Children

	children  Children
	items     []T
	InnerHTML template.HTML
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

	existingItems := list.Binding.Value

	var items []T
	if state.Data == nil {
		items = existingItems
	} else {
		existingItemsByName := make(map[string]T)
		for _, item := range existingItems {
			name := list.FullItemName(item)
			existingItemsByName[name] = item
		}

		prefix := list.Identity.FullName + "["
		itemNames := make(map[string]struct{})
		for k := range state.Data.Values {
			if suffix, ok := strings.CutPrefix(k, prefix); ok {
				if name, _, ok := strings.Cut(suffix, "]"); ok {
					itemNames[name] = struct{}{}
				}
			}
		}
		if debugLogList {
			log.Printf("forms: %T: itemNames = %v", list, maps.Keys(itemNames))
		}

		items = make([]T, 0, len(existingItemsByName)+10)
		for fullName := range itemNames {
			remove := JoinNames(list.Identity.FullName, fullName, "remove")
			if state.Data.Action == remove {
				continue
			}
			item, found := existingItemsByName[fullName]
			if !found {
				if list.NewItem == nil {
					continue
				}
				name, typ := list.SplitFullItemName(fullName)
				var ok bool
				item, ok = list.NewItem(name, typ)
				if !ok {
					if debugLogList {
						log.Printf("forms: %T: failed to add %q %q", list, typ, name)
					}
					continue
				}
				if debugLogList {
					log.Printf("forms: %T: added %q %q (full %q)", list, typ, name, fullName)
				}
			} else {
				delete(existingItemsByName, fullName)
			}
			items = append(items, item)
		}
		if list.NewItem != nil {
			addName := JoinNames(list.Identity.FullName, "add")
			if state.Data.Action == addName {
				if item, ok := list.NewItem("", ""); ok {
					items = append(items, item)
				}
			} else if suffix, ok := strings.CutPrefix(state.Data.Action, addName); ok {
				comps := SplitName(suffix)
				if len(comps) == 1 {
					typ := comps[0]
					if item, ok := list.NewItem("", typ); ok {
						items = append(items, item)
					}
				}
			}
		}

		for _, item := range existingItemsByName {
			if list.DeleteItem != nil {
				list.DeleteItem(item)
			}
		}
	}
	if list.Sort != nil {
		list.Sort(items)
	}
	list.items = items

	children := make(Children, 0, len(items)+10)
	if list.TopArea != nil {
		children = append(children, list.TopArea()...)
	}
	for _, item := range items {
		group := list.RenderItem(item)
		group.Name = list.FullItemName(item)
		children = append(children, group)
	}
	if list.BottomArea != nil {
		children = append(children, list.BottomArea()...)
	}
	list.children = children
}

func (list *List[T]) FullItemName(item T) string {
	name := list.ItemName(item)
	if list.ItemType == nil {
		return name
	} else {
		return list.ItemType(item) + ":" + name
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

func (*List[T]) EnumFields(f func(*Field)) {}

func (list *List[T]) EnumChildren(f func(Child)) {
	f(list.children)
}

func (list *List[T]) Process(fd *FormData) {
	list.Binding.Set(list.items)
}

func (list *List[T]) RenderInto(buf *strings.Builder, r *Renderer) {
	list.InnerHTML = r.Render(list.children)
	r.RenderWrapperTemplateInto(buf, list.Template, list, list.InnerHTML)
}
