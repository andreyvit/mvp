package mvp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/andreyvit/edb"
	"github.com/andreyvit/mvp/backoff"
	"github.com/andreyvit/mvp/flogger"
	"github.com/andreyvit/mvp/mvpjobs"
	"github.com/andreyvit/mvp/mvprpc"
)

var (
	errJobCrashed = errors.New("process crashed")
)

type JobImpl struct {
	Method         *mvprpc.Method
	Kind           *mvpjobs.Kind
	RepeatInterval time.Duration
}

func (app *App) Enqueue(rc *RC, kind *mvpjobs.Kind, in mvpjobs.Params) *mvpjobs.Job {
	name := in.JobName()
	if j := app.Job(rc, kind, name); j != nil {
		app.Reenqueue(rc, kind, j, in, false)
		return j
	}

	j := &mvpjobs.Job{
		ID:          app.NewID(),
		Kind:        kind.Name,
		Name:        in.JobName(),
		RawParams:   mvpjobs.EncodeParams(in),
		Status:      mvpjobs.StatusQueued,
		NextRunTime: rc.Now,
		EnqueueTime: rc.Now,
	}
	edb.Put(rc, j)
	return j
}

func (app *App) Reenqueue(rc *RC, kind *mvpjobs.Kind, j *mvpjobs.Job, in mvpjobs.Params, force bool) {
	if kind.Behavior == mvpjobs.Repeatable {
		if j.Status == mvpjobs.StatusRunning {
			j.Status = mvpjobs.StatusRunningPending
			if in != nil {
				j.RawParams = mvpjobs.EncodeParams(in)
			}
			j.NextRunTime = rc.Now
			edb.Put(rc, j)
		} else if j.Status.IsTerminal() {
			j.Status = mvpjobs.StatusQueued
			if in != nil {
				j.RawParams = mvpjobs.EncodeParams(in)
			}
			j.NextRunTime = rc.Now
			j.EnqueueTime = rc.Now
			edb.Put(rc, j)
		} else if j.Status.IsPending() && j.NextRunTime.After(rc.Now) {
			j.NextRunTime = rc.Now
			if in != nil {
				j.RawParams = mvpjobs.EncodeParams(in)
			}
			edb.Put(rc, j)
		}
	} else if force {
		if j.Status.IsTerminal() {
			j.Status = mvpjobs.StatusQueued
			if in != nil {
				j.RawParams = mvpjobs.EncodeParams(in)
			}
			j.NextRunTime = rc.Now
			j.EnqueueTime = rc.Now
			edb.Put(rc, j)
		} else if j.Status.IsPending() && j.NextRunTime.After(rc.Now) {
			j.NextRunTime = rc.Now
			if in != nil {
				j.RawParams = mvpjobs.EncodeParams(in)
			}
			edb.Put(rc, j)
		}
	}
}

func (app *App) failRunningJobs(ctx context.Context) {
	rc := NewRC(ctx, app, "jobs")

	var jobsToFail []*mvpjobs.Job
	app.MustRead(rc, func() {
		for c := edb.FullIndexScan[mvpjobs.Job](rc, runningJobsByStartTime); c.Next(); {
			j := c.Row()
			if !j.Status.IsRunning() {
				panic("job not running in running index: " + j.ID.String())
			}
			if app.JobSchema.KindByName(j.Kind) == nil {
				continue
			}
			jobsToFail = append(jobsToFail, j)
		}
	})
	if len(jobsToFail) > 0 {
		app.MustWrite(rc, func() {
			for _, j := range jobsToFail {
				kind := app.JobSchema.KindByName(j.Kind)
				app.markJobCompleted(rc, kind, j, errJobCrashed, -1)
			}
		})
		flogger.Log(rc, "recovered %d crashed jobs", len(jobsToFail))
	}
}

func (app *App) StartJobWorkers(ctx context.Context, count int, quitf func(err error)) {
	app.failRunningJobs(ctx)

	if count == 0 {
		return
	}

	donec := make(chan struct{}, count)
	for i := 1; i <= count; i++ {
		go app.runJobsContinuously(ctx, i, count, donec)
	}
	go func() {
		for i := 1; i <= count; i++ {
			<-donec
		}
		quitf(nil)
	}()
}

func (app *App) runJobsContinuously(ctx context.Context, workerIdx, workerCount int, donec chan<- struct{}) {
	defer sendSignal(donec)
	rc := NewRC(ctx, app, fmt.Sprintf("jobs:w%d", workerIdx))
	for ctx.Err() == nil {
		c := app.runPendingJobsOnce(rc, workerIdx, workerCount)
		if c == 0 {
			select {
			case <-time.After(5 * time.Second):
				break
			case <-ctx.Done():
				return
			}
		}
	}
}

func (app *App) RunPendingJobs(ctx context.Context) int {
	rc := NewRC(ctx, app, "jobs")
	return app.runPendingJobsOnce(rc, 0, 0)
}

func (app *App) RunJob(ctx context.Context, kind *mvpjobs.Kind, params mvpjobs.Params) error {
	return app.executeJob(ctx, app.NewID(), kind, params.JobName(), mvpjobs.EncodeParams(params), 1, 0, 0)
}

func (app *App) runPendingJobsOnce(rc *RC, workerIdx, workerCount int) int {
	var count int
	for {
		rc.Now = app.Now()
		if app.runSinglePendingJob(rc, workerIdx, workerCount) {
			count++
			// } else ...TODO: cron... { ...
		} else {
			break
		}
	}
	return count
}

func (app *App) runSinglePendingJob(rc *RC, workerIdx, workerCount int) bool {
	j := app.dequeuePendingJob(rc)
	if j == nil {
		return false
	}
	kind := app.JobSchema.KindByName(j.Kind)
	if kind == nil {
		panic(fmt.Errorf("dequeued unknown kind %q", j.Kind))
	}
	jobErr := app.executeJob(rc, j.ID, kind, j.Name, j.RawParams, j.Attempt, workerIdx, workerCount)
	dur := time.Since(j.StartTime)

	if jobErr != nil {
		log.Printf("** WARNING: job failed: %s %v %s: %v", j.Kind, j.ID, j.RawParams, jobErr)
	}

	app.MustWrite(rc, func() {
		j := edb.Reload(rc, j)
		app.markJobCompleted(rc, kind, j, jobErr, dur)
	})
	return true
}

func (app *App) markJobCompleted(rc *RC, kind *mvpjobs.Kind, j *mvpjobs.Job, jobErr error, dur time.Duration) {
	if dur >= 0 {
		j.LastDuration = dur
		j.TotalDuration += dur
	}
	now := app.Now()
	jobImpl := app.jobsByKind[kind]
	if jobImpl == nil {
		panic(fmt.Errorf("jobImpl not found for job %s", kind.Name))
	}

	if jobErr != nil {
		j.LastFailureTime = now
		j.LastErr = jobErr.Error()
		j.ConsecFailures++
		j.TotalFailures++
		delay := kind.Backoff.DelayAfter(j.ConsecFailures)
		if delay >= backoff.InfiniteDelay {
			if jobImpl.RepeatInterval > 0 {
				// cron jobs never fail, they fall back to their repeat interval
				j.Status = mvpjobs.StatusQueued
				j.NextRunTime = now.Add(jobImpl.RepeatInterval)
			} else {
				j.Status = mvpjobs.StatusFailed
				j.NextRunTime = time.Time{}
			}
		} else {
			j.Status = mvpjobs.StatusRetrying
			j.NextRunTime = now.Add(delay)
		}
	} else {
		j.LastSuccessTime = now
		j.LastErr = ""
		j.ConsecFailures = 0
		if jobImpl.RepeatInterval > 0 {
			j.Status = mvpjobs.StatusQueued
			j.NextRunTime = now.Add(jobImpl.RepeatInterval)
		} else {
			j.Status = mvpjobs.StatusDone
		}
	}
	edb.Put(rc, j)
}

func (app *App) executeJob(ctx context.Context, jid mvpjobs.JobID, kind *mvpjobs.Kind, name string, rawParams []byte, attempt int, workerIdx, workerCount int) error {
	in := kind.Method.NewIn().(mvpjobs.Params)
	err := json.Unmarshal(rawParams, in)
	if err != nil {
		return fmt.Errorf("failed to unmarshal job params into %v: %w", kind.Method.InType, err)
	}
	in.SetJobName(name)

	rc := NewRC(ctx, app, fmt.Sprintf("jobs:w%d:%s:%v:%d", workerIdx, kind.Name, jid, attempt))
	// rc.Auth = bm.Auth{
	// 	Type: bm.ActorTypeAdmin,
	// }

	m := app.methodsByName[kind.Method.Name]
	if m == nil {
		panic(fmt.Errorf("no impl registered for job %v", kind.Name))
	}

	_, err = app.doCall(rc, m, in)
	return err
}

func (app *App) dequeuePendingJob(rc *RC) *mvpjobs.Job {
	var j *mvpjobs.Job
	app.MustWrite(rc, func() {
		c := edb.IndexScan[mvpjobs.Job](rc, pendingJobsByRunTime, edb.FullScan())
		if c.Next() {
			j = c.Row()
			if j.NextRunTime.After(rc.Now) {
				// flogger.Log(rc, "dequeued job %s %v, which is in the future", j.Kind, j.ID)
				j = nil
			} else {
				app.markJobStarted(rc, j)
			}
		}
	})
	return j
}

func (app *App) markJobStarted(rc *RC, j *mvpjobs.Job) {
	now := app.Now()
	j.Attempt++
	j.Status = mvpjobs.StatusRunning
	j.StartTime = now
	j.LastAttemptTime = now
	edb.Put(rc, j)
}

func (app *App) Job(txish edb.Txish, kind *mvpjobs.Kind, name string) *mvpjobs.Job {
	if kind.AllowNames() {
		if name == "" {
			return nil
		}
		return edb.Lookup[mvpjobs.Job](txish, jobsByKindName, mvpjobs.KindName{Kind: kind.Name, Name: name})
	} else {
		if name != "" {
			return nil
		}
		return edb.Lookup[mvpjobs.Job](txish, jobsByKind, kind.Name)
	}
}
