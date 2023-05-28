package mvpjobs

import (
	"encoding/json"
	"time"

	"github.com/andreyvit/mvp/flake"
)

type JobID = flake.ID

type Job struct {
	ID        JobID           `msgpack:"-"`
	Kind      string          `msgpack:"k"`
	Name      string          `msgpack:"n"`
	RawParams json.RawMessage `msgpack:"p"`
	// Streams   []string        `msgpack:"str,omitempty"`

	Status      Status    `msgpack:"s2,omitempty"`
	NextRunTime time.Time `msgpack:"rt,omitempty"`
	Step        string    `msgpack:"sp,omitempty"`
	RawState    []byte    `msgpack:"st,omitempty"`

	EnqueueTime     time.Time `msgpack:"tq,omitempty"`
	StartTime       time.Time `msgpack:"ts,omitempty"`
	LastAttemptTime time.Time `msgpack:"ta,omitempty"`
	FinishTime      time.Time `msgpack:"tf,omitempty"`
	LastSuccessTime time.Time `msgpack:"tls,omitempty"`
	LastFailureTime time.Time `msgpack:"tlf,omitempty"`

	Attempt        int           `msgpack:"a,omitempty"`
	ConsecFailures int           `msgpack:"failc,omitempty"`
	TotalFailures  int           `msgpack:"failt,omitempty"`
	LastErr        string        `msgpack:"errl,omitempty"`
	LastDuration   time.Duration `msgpack:"durl,omitempty"`
	TotalDuration  time.Duration `msgpack:"durt,omitempty"`
}

func (j *Job) KindName() KindName {
	return KindName{j.Kind, j.Name}
}

func (j *Job) IsAnonymous() bool {
	return j.Name == ""
}

type KindName struct {
	Kind string
	Name string
}

func (kn KindName) IsAnonymous() bool {
	return kn.Name == ""
}

type Set struct {
	schema *Schema
	name   string
	kinds  map[string]*Kind
}

// func (schema *Schema) AddSet(setName string) *Set {
// 	set := &Set{
// 		schema: schema,
// 		name:   setName,
// 		kinds:  make(map[string]*Kind),
// 	}
// 	if schema.sets[setName] != nil {
// 		panic(fmt.Errorf("duplicate set %q", setName))
// 	}
// 	schema.sets[setName] = set
// 	return set
// }

func (schema *Schema) KindByName(kindName string) *Kind {
	return schema.kinds[kindName]
}
