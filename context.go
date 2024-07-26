package mvp

import (
	"context"
	"fmt"
)

var (
	AppContextKey = appContextKeyType{}
	RCContextKey  = rcContextKeyType{}
)

type (
	appContextKeyType struct{}
	rcContextKeyType  struct{}
)

func (appContextKeyType) String() string {
	return "mvp.AppContextKey"
}
func (rcContextKeyType) String() string {
	return "mvp.RCContextKey"
}

func AppFrom(ctx context.Context) *App {
	if v := ctx.Value(AppContextKey); v != nil {
		return v.(*App)
	}
	panic(fmt.Errorf("context does not have App in its chain: %v", ctx))
}

func RCFrom(ctx context.Context) *RC {
	if rc := OptionalRCFrom(ctx); rc != nil {
		return rc
	}
	panic(fmt.Errorf("context does not have RC in its chain: %v", ctx))
}

func OptionalRCFrom(ctx context.Context) *RC {
	if v := ctx.Value(RCContextKey); v != nil {
		return v.(*RC)
	}
	return nil
}
