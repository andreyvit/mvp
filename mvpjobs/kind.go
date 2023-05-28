package mvpjobs

import (
	"fmt"
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
}

func (k *Kind) IsCron() bool {
	return k.Behavior == Cron
}

func (k *Kind) Retry(bo backoff.Backoff) *Kind {
	k.Backoff = bo
	return k
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

func (scm *Schema) Define(kindName string, in any, behavior Behavior, opts ...any) *Kind {
	scm.init()
	kind := &Kind{
		schema:   scm,
		Behavior: behavior,
		Name:     kindName,
		Method:   scm.api.Method("Job"+kindName, in, nil),
	}
	for _, opt := range opts {
		switch opt := opt.(type) {
		case Persistence:
			kind.Persistence = opt
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

func (scm *Schema) Cron(kindName string, interval time.Duration, opts ...any) *Kind {
	kind := scm.Define(kindName, &NoParams{}, Cron, opts...)
	kind.RepeatInterval = interval
	return kind
}
