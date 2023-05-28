package mvp

import (
	"time"

	"github.com/andreyvit/edb"
	"github.com/andreyvit/mvp/mvpjobs"
)

var (
	builtinDBSchema = &edb.Schema{
		Name: "mvpbuiltin",
	}

	builtinModule = &Module{
		Name:     "mvpbuiltin",
		DBSchema: builtinDBSchema,
	}

	jobsTable = edb.AddTable(builtinDBSchema, "jobs", 2, func(row *mvpjobs.Job, ib *edb.IndexBuilder) {
		ib.Add(jobsByKind, row.Kind)
		if !row.IsAnonymous() {
			ib.Add(jobsByKindName, row.KindName())
		}
		if !row.NextRunTime.IsZero() && row.Status.IsPending() {
			ib.Add(pendingJobsByRunTime, row.NextRunTime)
		}
		if row.Status.IsRunning() {
			ib.Add(runningJobsByStartTime, row.StartTime)
		}
	}, func(tx *edb.Tx, row *mvpjobs.Job, oldVer uint64) {
	}, []*edb.Index{
		jobsByKind,
		jobsByKindName,
		pendingJobsByRunTime,
		runningJobsByStartTime,
	})
	jobsByKind             = edb.AddIndex[string]("by_kind")
	jobsByKindName         = edb.AddIndex[mvpjobs.KindName]("by_kind_name_v2").Unique()
	pendingJobsByRunTime   = edb.AddIndex[time.Time]("pending_by_run_time")
	runningJobsByStartTime = edb.AddIndex[time.Time]("running_by_start_time")
)
