package mvpjobs

import (
	"fmt"
	"reflect"
	"time"

	"github.com/andreyvit/mvp/backoff"
	"github.com/andreyvit/mvp/mvprpc"
)

type Kind struct {
	schema         *Schema
	Name           string
	Behavior       Behavior
	Persistence    Persistence
	Method         *mvprpc.Method
	Set            string
	RepeatInterval time.Duration
	Backoff        backoff.Backoff
	Enabled        bool
	Handler        any
}

func (k *Kind) IsCron() bool {
	return k.Behavior == Cron
}

func (k *Kind) IsPersistent() bool {
	return k.Persistence == Persistent
}

func (k *Kind) AllowNames() bool {
	return !k.IsCron()
}

func (scm *Schema) Kinds() []*Kind {
	kinds := make([]*Kind, 0, len(scm.kinds))
	for _, kind := range scm.kinds {
		kinds = append(kinds, kind)
	}
	return kinds
}

func (scm *Schema) PersistentKindNames() []string {
	names := make([]string, 0, len(scm.kinds))
	for _, kind := range scm.kinds {
		if kind.IsPersistent() {
			names = append(names, kind.Name)
		}
	}
	return names
}
func (scm *Schema) EphemeralKindNames() []string {
	names := make([]string, 0, len(scm.kinds))
	for _, kind := range scm.kinds {
		if !kind.IsPersistent() {
			names = append(names, kind.Name)
		}
	}
	return names
}

func isStructPtr(typ reflect.Type) bool {
	return typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Struct
}

func (scm *Schema) Define(kindName string, inOrFunc any, behavior Behavior, opts ...any) *Kind {
	var handler any
	var in any
	if inOrFunc == nil {
		inOrFunc = &NoParams{}
	}
	inTyp := reflect.TypeOf(inOrFunc)
	if inTyp.Kind() == reflect.Func && inTyp.NumIn() == 2 && isStructPtr(inTyp.In(1)) {
		handler = inOrFunc
		in = reflect.New(inTyp.In(1).Elem()).Interface()
	} else if isStructPtr(inTyp) {
		in = inOrFunc
	} else {
		panic(fmt.Errorf("invalid value for Define inOrFunc: %T %v", inOrFunc, inOrFunc))
	}

	scm.init()
	kind := &Kind{
		schema:      scm,
		Behavior:    behavior,
		Name:        kindName,
		Method:      scm.api.Method("Job"+kindName, in, nil),
		Enabled:     true,
		Persistence: Persistent,
		Handler:     handler,
	}
	for _, opt := range opts {
		switch opt := opt.(type) {
		case Persistence:
			kind.Persistence = opt
		case Behavior:
			kind.Behavior = opt
		case backoff.Backoff:
			kind.Backoff = opt
		case WithRepeatInterval:
			kind.RepeatInterval = time.Duration(opt)
		case WithEnabled:
			kind.Enabled = bool(opt)
		default:
			panic(fmt.Errorf("%s: unknown options %T %v", kindName, opt, opt))
		}
	}
	// for _, tag := range strings.Fields(tags) {
	// 	scm.byTag[tag] = append(scm.byTag[tag], kind)
	// }

	scm.addKind(kind)

	return kind
}

// func (scm *Schema) Cron(kindName string, interval time.Duration, opts ...any) *Kind {
// 	kind := scm.Define(kindName, &NoParams{}, Cron, opts...)
// 	kind.RepeatInterval = interval
// 	return kind
// }
