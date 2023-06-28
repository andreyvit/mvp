package mvp

import (
	"fmt"
	"net/http"
	"reflect"
)

const (
	MwSlotRateLimitAdjustment = "ratelimitconf"
	MwSlotRateLimiter         = "ratelimit"
)

func (app *App) addBuiltinMiddleware(b *RouteBuilder) {
	b.UseIn(MwSlotRateLimitAdjustment, nil)
	b.UseIn(MwSlotRateLimiter, app.enforceRateLimit)
}

type middlewareFunc = func(rc *RC) (any, error)

type middlewareSlot struct {
	name string
	f    middlewareFunc
}

type middlewareSlotList []middlewareSlot

func (mwlist middlewareSlotList) Index(name string) int {
	if name == "" {
		return -1
	}
	for i, mw := range mwlist {
		if mw.name == name {
			return i
		}
	}
	return -1
}

func (mwlist middlewareSlotList) Clone() middlewareSlotList {
	return append(middlewareSlotList(nil), mwlist...)
}

func (mwlist *middlewareSlotList) Add(name string, f middlewareFunc) {
	i := mwlist.Index(name)
	if i >= 0 {
		(*mwlist)[i].f = f
	} else {
		*mwlist = append(*mwlist, middlewareSlot{name: name, f: f})
	}
}

func adaptMiddleware(f any) func(rc *RC) (any, error) {
	switch f := f.(type) {
	case nil:
		return nil
	case func(rc *RC) (any, error):
		return f
	case func(w http.ResponseWriter, r *http.Request) bool:
		return func(rc *RC) (any, error) {
			if f(rc.RespWriter, rc.Request.Request) {
				return ResponseHandled{}, nil
			}
			return nil, nil
		}
	default:
		fv := reflect.ValueOf(f)
		if fv.Kind() != reflect.Func {
			panic(fmt.Errorf("unsupported middleware type %T", f))
		}
		ft := fv.Type()
		if ft.NumIn() == 1 && ft.NumOut() == 2 && ft.Out(1) == errorType {
			if rcf := BaseRC.FacetByPtrType(ft.In(0)); rcf != nil {
				return func(rc *RC) (any, error) {
					out := fv.Call([]reflect.Value{
						reflect.ValueOf(rcf.AnyFrom(rc)),
					})
					return out[0].Interface(), errFromAny(out[1].Interface())
				}
			}
		}

		panic(fmt.Errorf("unknown middleware signature %T", f))
	}
}
