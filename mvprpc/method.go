package mvprpc

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/andreyvit/mvp/expandable"
	mvpm "github.com/andreyvit/mvp/mvpmodel"
)

var (
	ObjectSchema = expandable.NewSchema("mvprpc")
	BaseMethod   = expandable.NewBase[Method](ObjectSchema, "method")
)

type API struct {
	Name            string
	methodsByName   map[string]*Method
	methodsByNameCI map[string]*Method
}

func (api *API) init() {
	if api.methodsByName == nil {
		api.methodsByName = make(map[string]*Method)
		api.methodsByNameCI = make(map[string]*Method)
	}
}

type Method struct {
	expandable.Impl
	Name      string
	InType    reflect.Type
	InPtrType reflect.Type
	OutType   reflect.Type
	NewIn     func() any

	StoreAffinity mvpm.StoreAffinity
}

func (api *API) Method(name string, in any, out any, opts ...any) *Method {
	api.init()
	if in == nil {
		in = &struct{}{}
	}
	inPtrType := reflect.TypeOf(in)
	if inPtrType.Kind() != reflect.Ptr || inPtrType.Elem().Kind() != reflect.Struct {
		panic(fmt.Errorf("invalid input type %v, expected pointer to struct", inPtrType))
	}
	inType := inPtrType.Elem()

	meth := &Method{
		Name:      name,
		InType:    inType,
		InPtrType: inPtrType,
		OutType:   reflect.TypeOf(out), // nil when out == nil
		NewIn: func() any {
			return reflect.New(inType).Interface()
		},
	}

	for _, opt := range opts {
		switch opt := opt.(type) {
		case mvpm.StoreAffinity:
			meth.StoreAffinity = opt
		default:
			panic(fmt.Errorf("%s: invalid option %T %v", name, opt, opt))
		}
	}

	lower := strings.ToLower(name)
	if api.methodsByNameCI[lower] != nil {
		panic(fmt.Errorf("method %s is already defined", name))
	}
	api.methodsByName[name] = meth
	api.methodsByNameCI[lower] = meth

	return meth
}
