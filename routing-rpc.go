package mvp

import (
	"fmt"
	"reflect"

	"github.com/andreyvit/mvp/mvpjobs"
	"github.com/andreyvit/mvp/mvprpc"
)

type MethodImpl struct {
	*mvprpc.Method
	call func(rc *RC, in any) (result any, err error)
}

func (app *App) doCall(rc *RC, m *MethodImpl, in any) (any, error) {
	var out any
	callErr := app.InTx(rc, m.StoreAffinity, func() error {
		var err error
		out, err = m.call(rc, in)
		return err
	})
	return out, callErr
}

type JobRegistry interface {
	JobImpl(kind *mvpjobs.Kind, impl any)
}

type CallRegistry interface {
	CallImpl(method *mvprpc.Method, impl any)
}

func (app *App) JobImpl(kind *mvpjobs.Kind, impl any) {
	if app.jobsByKind == nil {
		app.jobsByKind = make(map[*mvpjobs.Kind]*JobImpl)
	}
	if app.jobsByKind[kind] != nil {
		panic(fmt.Errorf("job %s already has an impl defined", kind.Name))
	}
	app.jobsByKind[kind] = &JobImpl{
		Kind:           kind,
		RepeatInterval: kind.RepeatInterval,
	}
	app.MethodImpl(kind.Method, impl)
}

func (app *App) MethodImpl(method *mvprpc.Method, impl any) {
	// Cautionary tale. This could, and have been, written
	// in a shorter way using generics. But it wasn't flexible,
	// and error messages were worse, and when I found myself
	// adding separate methods for various signatures, I finally
	// remembered that generics are almost always a bad idea.
	// Don't repeat that mistake.
	name := method.Name

	fv := reflect.ValueOf(impl)
	ft := fv.Type()
	if ft.Kind() != reflect.Func {
		panic(fmt.Sprintf("%s: impl must be a func", name))
	}
	if ft.NumIn() < 1 {
		panic(fmt.Sprintf("%s: impl must accept *RC as first param", name))
	}
	rcFacet := BaseRC.FacetByPtrType(ft.In(0))
	if rcFacet == nil {
		panic(fmt.Sprintf("%s: impl must accept *RC as first param", name))
	}

	var hasIn bool
	if ft.NumIn() == 2 && ft.In(1) == method.InPtrType {
		hasIn = true // ok
	} else if method.InPtrType == emptyStructPtrType && ft.NumIn() == 1 {
		hasIn = false // ok
	} else {
		panic(fmt.Sprintf("%s: impl must accept (*RC, %v)", name, method.InPtrType))
	}

	var resultIdx, errIdx int
	if ft.NumOut() == 1 && ft.Out(0) == errorType {
		resultIdx, errIdx = -1, 0
	} else if ft.NumOut() == 2 && ft.Out(1) == errorType {
		resultIdx, errIdx = 0, 1
	} else {
		panic(fmt.Sprintf("%s: impl must return error or (whatever, error)", name))
	}

	call := func(rc *RC, in any) (result any, err error) {
		args := make([]reflect.Value, ft.NumIn())
		args[0] = reflect.ValueOf(rcFacet.AnyFrom(rc))
		if hasIn {
			args[1] = reflect.ValueOf(in)
		}
		results := fv.Call(args)
		if resultIdx >= 0 {
			result = results[resultIdx].Interface()
		}
		if errVal := results[errIdx].Interface(); errVal != nil {
			err = errVal.(error)
		}
		return
	}

	m := &MethodImpl{
		Method: method,
		call:   call,
	}
	if app.methodsByName == nil {
		app.methodsByName = make(map[string]*MethodImpl)
	}
	if app.methodsByName[name] != nil {
		panic(fmt.Errorf("%s method already has an impl defined", name))
	}
	app.methodsByName[name] = m
}
