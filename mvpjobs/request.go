package mvpjobs

import (
	"encoding/json"
	"time"
)

type Request struct {
	Kind   string          `msgpack:"k"`
	Name   string          `msgpack:"n"`
	Params json.RawMessage `msgpack:"p"`
	EnqueueOptions
}

type EnqueueOptions struct {
	RunTime time.Time `msgpack:"rt,omitempty"`
}
