package mvputil

import (
	"reflect"
	"runtime"
	"strings"
)

func FuncName(f interface{}) string {
	return PureFuncName(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name())
}

func PureFuncName(s string) string {
	if i := strings.LastIndexByte(s, '.'); i >= 0 {
		s = s[i+1:]
	}
	s = strings.TrimSuffix(s, "-fm")
	return s
}
