package mvpmetrics

// import (
// 	"fmt"
// 	"slices"

// 	"golang.org/x/exp/maps"
// )

// type Registry struct {
// 	finalized bool
// 	metrics   map[string]Metric
// 	names     []string
// }

// var DefaultRegistry = &Registry{}

// type Metric interface {
// 	Name() string
// 	WriteMetricTo(mw *Writer)
// }

// func (reg *Registry) Add(m Metric) {
// 	name := m.Name()
// 	if reg.finalized {
// 		panic(fmt.Errorf("cannot add metric %s to a finalized registry"))
// 	}
// 	if reg.metrics == nil {
// 		reg.metrics = make(map[string]Metric)
// 	}
// 	if prev := reg.metrics[name]; prev != nil {
// 		panic(fmt.Errorf("metric %s is already registered"))
// 	}
// 	reg.metrics[name] = m
// }

// func (reg *Registry) finalize() {
// 	reg.names = maps.Keys(reg.metrics)
// 	slices.Sort(reg.names)
// }
