// Package director launches a bunch of components and gets them restarted on failure.
package director

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

var ErrDisabled = errors.New("disabled")
var ErrFinished = errors.New("finished successfully")

type StartFunc func(ctx context.Context, quitf func(err error)) error

type Component struct {
	Name         string
	Critical     bool
	RestartDelay time.Duration
}

type compState struct {
	Comp     *Component
	Start    StartFunc
	QuitC    chan error
	Failures int
}

func (cs *compState) OnQuit(err error) {
	cs.QuitC <- err
}

type Director struct {
	doneWG sync.WaitGroup
}

func New() *Director {
	return &Director{}
}

func (dr *Director) Start(ctx context.Context, comp *Component, startf StartFunc) error {
	cs := &compState{
		Comp:  comp,
		Start: startf,
		QuitC: make(chan error, 1),
	}
	log.Printf("[start] %s", cs.Comp.Name)
	err := startf(ctx, cs.OnQuit)
	if err == ErrDisabled {
		return nil
	} else if err != nil {
		return err
	}

	dr.doneWG.Add(1)
	go dr.run(ctx, cs)
	return nil
}

func (dr *Director) Wait() {
	dr.doneWG.Wait()
}

func (dr *Director) run(ctx context.Context, cs *compState) {
	defer dr.doneWG.Done()
	for {
		err := <-cs.QuitC
		if err == ErrDisabled {
			return
		}
		if err == nil {
			if ctx.Err() != nil {
				log.Printf("[done] %s", cs.Comp.Name)
				return
			} else {
				err = ErrFinished
			}
		}
		cs.Failures++
		log.Printf("ERROR: [failed] %s: %v", cs.Comp.Name, err)

		select {
		case <-time.After(cs.Comp.RestartDelay):
			break
		case <-ctx.Done():
			return
		}

		log.Printf("[restart] %s", cs.Comp.Name)
		err = cs.Start(ctx, cs.OnQuit)
		go cs.OnQuit(err)
	}
}
