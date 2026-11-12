package mvp

import (
	"time"

	"github.com/andreyvit/edb"
	"github.com/andreyvit/edb/edbindexing"
	"github.com/andreyvit/edb/kvo"
	"github.com/andreyvit/edb/values"
	"github.com/andreyvit/mvp/mvpjobs"
	mvpm "github.com/andreyvit/mvp/mvpmodel"
)

var (
	builtinDBSchema = &edb.Schema{
		Name: "mvpbuiltin",
	}

	builtinModule = &Module{
		Name:     "mvpbuiltin",
		DBSchema: builtinDBSchema,
	}

	jobsTable = edb.DefineTable(builtinDBSchema, "jobs", jobModel, func(b *edb.TableBuilder[mvpjobs.JobID]) {
		b.SetSchemaVersion(2)

		jobsByKind = b.AddIndex("by_kind", edbindexing.ByField(jobKind))
		jobsByKindName = b.AddIndex[mvpjobs.KindName]("by_kind_name_v2", func(_ *edb.Tx, _ values.Value, row kvo.Record, yield func(mvpjobs.KindName)) {
			if name := row.Get(jobName); name != "" {
				yield(mvpjobs.KindName{Kind: row.Get(jobKind), Name: name})
			}
		}, edb.IndexOptDeprecatedUnique)
		pendingJobsByRunTime = b.AddIndex[time.Time]("pending_by_run_time", func(_ *edb.Tx, _ values.Value, row kvo.Record, yield func(time.Time)) {
			if next := row.Get(jobNextRunTime); !next.IsZero() && row.Get(jobStatus).IsPending() {
				yield(next)
			}
		})
		runningJobsByStartTime = b.AddIndex[time.Time]("running_by_start_time", func(_ *edb.Tx, _ values.Value, row kvo.Record, yield func(time.Time)) {
			if row.Get(jobStatus).IsRunning() {
				yield(row.Get(jobStartTime))
			}
		})
	})
	jobsByKind             *edb.Index[string]
	jobsByKindName         *edb.Index[mvpjobs.KindName]
	pendingJobsByRunTime   *edb.Index[time.Time]
	runningJobsByStartTime *edb.Index[time.Time]

	migrationsTable = edb.DefineTable(builtinDBSchema, "migrations", migrationModel, func(b *edb.TableBuilder[mvpm.MigrationID]) {
	})
)
