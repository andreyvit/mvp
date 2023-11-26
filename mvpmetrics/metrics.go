package mvpmetrics

import (
	"fmt"
	"sync/atomic"
)

type (
	Help  string
	Scale float64
)

type desc struct {
	name       string
	labelNames []string
	help       string
	scale      float64
}

func (d *desc) Name() string {
	return d.name
}

func (d *desc) verifyCorrectLabels(labelValues []string) {
	if len(labelValues) != len(d.labelNames) {
		panic(fmt.Errorf("invalid number of label values, got %d, wanted %d", len(labelValues), len(d.labelNames)))
	}
}

type Counter struct {
	desc
	values vector[uint64]
}

func NewCounter(name string, labelNames []string, opts ...any) *Counter {
	m := &Counter{
		desc: desc{
			name:       name,
			labelNames: labelNames,
		},
	}
	reg := DefaultRegistry
	for _, opt := range opts {
		switch opt := opt.(type) {
		case Help:
			m.help = string(opt)
		case Scale:
			m.scale = float64(opt)
		case *Registry:
			reg = opt
		default:
			panic(fmt.Errorf("invalid option %T %v", opt, opt))
		}
	}
	reg.Add(m)
	return m
}

func (m *Counter) WriteMetricTo(mw *Writer) {
	m.values.enum(func(labelValues []string, value uint64) {
		if m.scale == 0 {
			mw.WriteUint(m.name, m.labelNames, labelValues, value)
		} else {
			mw.WriteFloat(m.name, m.labelNames, labelValues, float64(value)/m.scale)
		}
	})
}

func (m *Counter) Inc(labelValues ...string) {
	m.Add(1, labelValues...)
}

func (m *Counter) Add(delta uint64, labelValues ...string) {
	m.desc.verifyCorrectLabels(labelValues)
	v, meta := m.values.acquire(labelValues)
	atomic.AddUint64(v, delta)
	m.values.release(meta)
}

func ClampCounterDelta(delta int64) uint64 {
	if delta >= 0 {
		return uint64(delta)
	} else {
		return 0
	}
}

type Gauge struct {
	desc
	values vector[int64]
}

func NewGauge(name string, labelNames []string, opts ...any) *Gauge {
	m := &Gauge{
		desc: desc{
			name:       name,
			labelNames: labelNames,
		},
	}
	reg := DefaultRegistry
	for _, opt := range opts {
		switch opt := opt.(type) {
		case Help:
			m.help = string(opt)
		case Scale:
			m.scale = float64(opt)
		case *Registry:
			reg = opt
		default:
			panic(fmt.Errorf("invalid option %T %v", opt, opt))
		}
	}
	reg.Add(m)
	return m
}

func (m *Gauge) WriteMetricTo(mw *Writer) {
	m.values.enum(func(labelValues []string, value int64) {
		if m.scale == 0 {
			mw.WriteInt(m.name, m.labelNames, labelValues, value)
		} else {
			mw.WriteFloat(m.name, m.labelNames, labelValues, float64(value)/m.scale)
		}
	})
}

func (m *Gauge) Set(value int64, labelValues ...string) {
	m.desc.verifyCorrectLabels(labelValues)
	v, meta := m.values.acquire(labelValues)
	atomic.StoreInt64(v, value)
	m.values.release(meta)
}

func (m *Gauge) Add(delta int64, labelValues ...string) {
	m.desc.verifyCorrectLabels(labelValues)
	v, meta := m.values.acquire(labelValues)
	atomic.AddInt64(v, delta)
	m.values.release(meta)
}
