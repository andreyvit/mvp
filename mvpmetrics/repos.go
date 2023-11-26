package mvpmetrics

import (
	"cmp"
	"fmt"
	"slices"
	"sync"

	"golang.org/x/exp/maps"
)

type Registry struct {
	finOnce       sync.Once
	finalized     bool
	metricsByName map[string]Metric
	metrics       []Metric
}

var DefaultRegistry = &Registry{}

type Metric interface {
	Name() string
	WriteMetricTo(mw *Writer)
}

func (reg *Registry) Add(m Metric) {
	name := m.Name()
	if reg.finalized {
		panic(fmt.Errorf("cannot add metric %s to a finalized registry"))
	}
	if reg.metricsByName == nil {
		reg.metricsByName = make(map[string]Metric)
	}
	if prev := reg.metricsByName[name]; prev != nil {
		panic(fmt.Errorf("metric %s is already registered"))
	}
	reg.metricsByName[name] = m
}

func (reg *Registry) Metrics() []Metric {
	reg.finOnce.Do(reg.finalize)
	return reg.metrics
}

func (reg *Registry) finalize() {
	if reg.finalized {
		return
	}
	reg.metrics = maps.Values(reg.metricsByName)
	slices.SortFunc(reg.metrics, func(a, b Metric) int {
		return cmp.Compare(a.Name(), b.Name())
	})
}

func (reg *Registry) WriteMetricsTo(mw *Writer) {
	for _, m := range reg.Metrics() {
		m.WriteMetricTo(mw)
	}
}
