package mvp

import (
	"context"
	"fmt"
	"sync"

	"github.com/andreyvit/mvp/flogger"
	"github.com/andreyvit/mvp/mvpjobs"
)

type EphemeralJobQueue struct {
	mut sync.Mutex
	m   map[string]bool
	q   chan *EphemeralJob
}

type EphemeralJob struct {
	Kind *mvpjobs.Kind
	Key  string
	F    func(rc *RC) error
}

func (app *App) initEphemeralJobs() {
	app.ephemeralJobQueue.m = make(map[string]bool)
	app.ephemeralJobQueue.q = make(chan *EphemeralJob, app.Settings.EphemeralQueueMaxSize)
}

func (app *App) EnqueueEphemeral(kind *mvpjobs.Kind, name string, f func(rc *RC) error) {
	if kind.Persistence != mvpjobs.Ephemeral {
		panic("EnqueueEphemeral requires an ephemeral job")
	}

	var key string
	if name == "" {
		key = kind.Name
	} else {
		key = fmt.Sprintf("%s:%s", kind.Name, name)
	}

	if !app.startEphemeralJob(kind, key) {
		return
	}
	job := &EphemeralJob{kind, key, f}
	if app.Settings.IsTesting {
		rc := NewRC(context.Background(), app, "ejobs:inline")
		defer rc.Close()
		// TODO: better context?
		app.runEphemeralJob(rc, job)
	} else {
		app.ephemeralJobQueue.q <- job
	}
}

func (app *App) StartEphemeralJobWorkers(ctx context.Context, count int, quitf func(err error)) {
	if count == 0 {
		return
	}

	wg := new(sync.WaitGroup)
	for i := 1; i <= count; i++ {
		wg.Add(1)
		go app.runEphemeralJobsContinuously(ctx, i, count, wg)
	}
	go func() {
		wg.Wait()
		quitf(nil)
	}()
}

func (app *App) runEphemeralJobsContinuously(ctx context.Context, workerIdx, workerCount int, wg *sync.WaitGroup) {
	defer wg.Done()
	rc := NewRC(ctx, app, fmt.Sprintf("ejobs:w%d", workerIdx))
	defer rc.Close()
	for ctx.Err() == nil {
		select {
		case job := <-app.ephemeralJobQueue.q:
			app.runEphemeralJob(rc, job)
		case <-ctx.Done():
			break
		}
	}
}

func (app *App) runEphemeralJob(rc *RC, job *EphemeralJob) {
	defer app.finishEphemeralJob(job)

	rc.RequestID = job.Key
	defer func() { rc.RequestID = "" }()

	flogger.Log(rc, "ephemeral job starting")
	err := rc.InTx(job.Kind.Method.StoreAffinity, func() error {
		return job.F(rc)
	})
	if err != nil {
		flogger.Log(rc, "ERROR: ephemeral job failed: %v", err)
	} else {
		flogger.Log(rc, "ephemeral job finished")
	}
}

func (app *App) startEphemeralJob(kind *mvpjobs.Kind, key string) bool {
	app.ephemeralJobQueue.mut.Lock()
	defer app.ephemeralJobQueue.mut.Unlock()
	if again, found := app.ephemeralJobQueue.m[key]; found {
		if !again && kind.Behavior.IsRepeatable() {
			app.ephemeralJobQueue.m[key] = true
		}
		return false
	}
	app.ephemeralJobQueue.m[key] = false
	return true
}

func (app *App) finishEphemeralJob(job *EphemeralJob) {
	app.ephemeralJobQueue.mut.Lock()
	defer app.ephemeralJobQueue.mut.Unlock()

	if job.Kind.Behavior.IsRepeatable() && app.ephemeralJobQueue.m[job.Key] {
		go func() {
			app.ephemeralJobQueue.q <- job
		}()
	}

	delete(app.ephemeralJobQueue.m, job.Key)
}
