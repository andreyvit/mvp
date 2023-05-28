package mvp

import (
	"reflect"
	"regexp"
)

var validHTTPMethods = map[string]struct {
	Idempotent bool
}{
	"GET":    {Idempotent: true},
	"POST":   {},
	"PUT":    {},
	"DELETE": {},
	"OPTION": {Idempotent: true},
}

var (
	emptyStructType                 = reflect.TypeOf(struct{}{})
	emptyStructPtrType              = reflect.TypeOf((*struct{})(nil))
	rcPtrType          reflect.Type = reflect.TypeOf((*RC)(nil))
	errorType          reflect.Type = reflect.TypeOf((*error)(nil)).Elem()
	pathParamsRe                    = regexp.MustCompile(`:(\w+)`)
)
