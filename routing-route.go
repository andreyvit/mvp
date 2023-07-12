package mvp

import (
	"reflect"

	"github.com/andreyvit/mvp/expandable"
	mvpm "github.com/andreyvit/mvp/mvpmodel"
)

type Route struct {
	desc           string
	routeName      string
	method         string
	path           string
	bodyParamNames []string
	funcVal        reflect.Value
	rcFacet        expandable.Any[RC]
	inType         reflect.Type
	idempotent     bool
	storeAffinity  mvpm.StoreAffinity
	pathParams     []string
	routingContext
}

// RouteName returns the name passed to RouteBuilder.Route.
func (r *Route) RouteName() string {
	return r.routeName
}

// Method returns the HTTP method.
func (r *Route) Method() string {
	return r.method
}

// RouteName returns the HTTP method.
func (r *Route) Path() string {
	return r.path
}

// Description returns "callName method path"
func (r *Route) Description() string {
	return r.desc
}

// IsIdempotent whether the route is idempotent, which generally means the GET
// method, although some GET routes might be manually marked as non-idempotent
// via Route.Mutator().
func (r *Route) IsIdempotent() bool {
	return r.idempotent
}

func (r *Route) PathParams() []string {
	return r.pathParams
}

func (r *Route) BodyParamNames() []string {
	return r.bodyParamNames
}

func (route *Route) Mutator() {
	route.idempotent = false
}
