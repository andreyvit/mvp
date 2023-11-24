package mvpmetrics

import (
	"sync"
	"sync/atomic"
)

type ValueVector[K comparable] struct {
	lock         sync.RWMutex
	labelIndices map[K]int
	values       []int64
}

func (vv *ValueVector[K]) init() {
	vv.labelIndices = make(map[K]int)
	vv.values = make([]int64, 0, 10)
}

func (vv *ValueVector[K]) Inc(labels K) {
	v, wl := acquire(&vv.lock, vv.labelIndices, &vv.values, labels)
	atomic.AddInt64(v, 1)
	release(&vv.lock, wl)
}

func (vv *ValueVector[K]) Add(delta int64, labels K) {
	v, wl := acquire(&vv.lock, vv.labelIndices, &vv.values, labels)
	atomic.AddInt64(v, delta)
	release(&vv.lock, wl)
}

func acquire[K comparable, T any](lock *sync.RWMutex, labelIndices map[K]int, values *[]T, labels K) (*T, bool) {
	lock.RLock()
	i, ok := labelIndices[labels]
	if ok {
		return &(*values)[i], false
	}
	lock.RUnlock()
	lock.Lock()

	i, ok = labelIndices[labels]
	vals := *values
	if ok {
		return &vals[i], true
	}

	i = len(vals)
	labelIndices[labels] = i

	if i+1 <= cap(vals) {
		vals = vals[:i+1]
	} else {
		vals = make([]T, i+1, cap(vals)*2)
		copy(vals, *values)
	}
	*values = vals
	return &vals[i], true
}

func release(lock *sync.RWMutex, isWriteLocked bool) {
	if isWriteLocked {
		lock.Unlock()
	} else {
		lock.RUnlock()
	}
}

// import "fmt"

// type Help string

// type Value struct {
// 	name   string
// 	labels []string
// 	help   string
// }

// func New(name string, labels []string, opts ...any) *Value {
// 	m := &Value{
// 		name:   name,
// 		labels: labels,
// 	}
// 	reg := DefaultRegistry
// 	for _, opt := range opts {
// 		switch opt := opt.(type) {
// 		case Help:
// 			m.help = string(opt)
// 		case *Registry:
// 			reg = opt
// 		default:
// 			panic(fmt.Errorf("invalid option %T %v", opt, opt))
// 		}
// 	}
// 	reg.Add(m)
// 	return m
// }

// func (m *Value) Name() string {
// }

// func (m *Value) Append(mw *Writer) {
// }

// func (m *Value) WriteMetricTo(mw *Writer) {
// }

// func (m *Value) Inc(labelVals ...string) {
// 	if len(labelVals) != len(m.labels) {
// 		panic(fmt.Errorf("mvpmetrics.Value.Inc: invalid number of label values, got %d, wanted %d", len(labelVals), len(m.labels)))
// 	}
// }
