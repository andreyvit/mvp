package forms

// DynamicChild is a container that creates its child during Finalize
// based on a Resolver function. The resolved type name is included in
// the field name path to prevent stale postback issues.
type DynamicChild struct {
	RenderableImpl[DynamicChild]
	Name     string
	Resolver func() (typeName string, child Child)
	child    Child
}

func (dc *DynamicChild) Finalize(state *State) {
	if dc.Resolver == nil {
		return
	}
	typeName, child := dc.Resolver()
	if typeName == "" || child == nil {
		return
	}
	dc.child = child
	state.PushName(dc.Name + ":" + typeName)
}

func (dc *DynamicChild) EnumChildren(f func(Child, ChildFlags)) {
	if dc.child != nil {
		f(dc.child, 0)
	}
}
