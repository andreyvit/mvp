package mvpjobs

import (
	"fmt"

	"github.com/andreyvit/mvp/mvprpc"
)

type Schema struct {
	Name  string
	kinds map[string]*Kind
	byTag map[string][]*Kind
	api   mvprpc.API
}

func (scm *Schema) init() {
	if scm.kinds == nil {
		scm.kinds = make(map[string]*Kind)
		scm.byTag = make(map[string][]*Kind)
		scm.api.Name = scm.Name
	}
}

func (scm *Schema) Include(peer *Schema) {
	scm.init()
	for _, kind := range peer.kinds {
		if existing := scm.kinds[kind.Name]; existing != nil {
			panic(fmt.Errorf("job kind %q defined in both %q and %q", kind.Name, existing.schema.Name, kind.schema.Name))
		}
		scm.addKind(kind)
	}
}

func (scm *Schema) addKind(kind *Kind) {
	scm.kinds[kind.Name] = kind
}
