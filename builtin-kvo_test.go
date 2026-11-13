package mvp

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/andreyvit/edb"
	"github.com/andreyvit/edb/kvo"
	"github.com/andreyvit/mvp/mvpjobs"
	mvpm "github.com/andreyvit/mvp/mvpmodel"
)

func TestJobRecordIndexing(t *testing.T) {
	db, err := edb.Open(filepath.Join(t.TempDir(), "jobs.db"), builtinDBSchema, edb.Options{IsTesting: true})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, time.June, 12, 12, 0, 0, 123456789, time.UTC)
	projection := kvo.Project(reflect.TypeFor[mvpjobs.Job](), jobModel, nil)
	if !projection.DirectMessagePack() {
		t.Fatal("job indexing requires a decoded row")
	}
	db.Write(edb.TxMetadata{}, func(tx *edb.Tx) {
		row := &mvpjobs.Job{ID: 1, Kind: "email", Name: "welcome", Status: mvpjobs.StatusQueued, NextRunTime: now, StartTime: now.Add(-time.Hour), RawParams: json.RawMessage(`{"keep":[1,2,3]}`), RawState: []byte("opaque")}
		check := func() {
			t.Helper()
			assertCount := func(idx *edb.IndexDef, key any, want int) {
				count := 0
				for c := tx.IndexScan(idx, edb.ExactScan(key)); c.Next(); {
					count++
					if c.UintKey() != 1 {
						t.Fatal("wrong job key")
					}
				}
				if count != want {
					t.Fatalf("%s: got %d, want %d", idx.FullName(), count, want)
				}
			}
			assertCount(jobsByKind.IndexDef(), row.Kind, 1)
			named := 0
			if !row.IsAnonymous() {
				named = 1
			}
			assertCount(jobsByKindName.IndexDef(), row.KindName(), named)
			pending := 0
			if !row.NextRunTime.IsZero() && row.Status.IsPending() {
				pending = 1
			}
			assertCount(pendingJobsByRunTime.IndexDef(), row.NextRunTime, pending)
			running := 0
			if row.Status.IsRunning() {
				running = 1
			}
			assertCount(runningJobsByStartTime.IndexDef(), row.StartTime, running)
			got := edb.Get[mvpjobs.Job](tx, row.ID)
			// MessagePack preserves instants, not time.Location metadata.
			got.NextRunTime = got.NextRunTime.UTC()
			got.StartTime = got.StartTime.UTC()
			if !reflect.DeepEqual(got, row) {
				t.Fatalf("job payload changed: got %#v, want %#v", got, row)
			}
		}
		for _, status := range []mvpjobs.Status{mvpjobs.StatusQueued, mvpjobs.StatusRunning, mvpjobs.StatusRunningPending, mvpjobs.StatusRetrying, mvpjobs.StatusDone} {
			row.Status = status
			edb.Put(tx, row)
			check()
			tx.Reindex(jobsTable, nil)
			check()
			rec := tx.GetRecord(jobsTable, row.ID)
			if n := testing.AllocsPerRun(100, func() { jobStatus.Get(rec); jobNextRunTime.Get(rec); jobName.Get(rec) }); n != 0 {
				t.Fatalf("job access: %g allocations", n)
			}
		}
		edit := tx.OpenRecordEdit(jobsTable, row.ID)
		jobName.Set(edit.Editor(), "")
		tx.SaveRecord(edit)
		row.Name = ""
		check()

	})
}

func TestBuiltinTableModels(t *testing.T) {
	for _, table := range builtinDBSchema.Tables() {
		if table.Model() == nil {
			t.Fatalf("%s has no model", table.Name())
		}
	}
	db, err := edb.Open(filepath.Join(t.TempDir(), "migrations.db"), builtinDBSchema, edb.Options{IsTesting: true})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, time.November, 11, 12, 0, 0, 123456789, time.UTC)
	db.Write(edb.TxMetadata{}, func(tx *edb.Tx) {
		row := &mvpm.MigrationRecord{ID: 1, Name: "upgrade", ExecutionTime: now, ExecutionDuration: 123 * time.Nanosecond}
		edb.Put(tx, row)
		r := tx.GetRecord(migrationsTable, row.ID)
		if r.MessagePack() == nil || r.Get(migrationName) != "upgrade" || !r.Get(migrationTime).Equal(now) || r.Get(migrationDuration) != row.ExecutionDuration {
			t.Fatal("migration record access")
		}
		if n := testing.AllocsPerRun(100, func() { r.Get(migrationName); r.Get(migrationTime); r.Get(migrationDuration) }); n != 0 {
			t.Fatal("allocations", n)
		}
		e := tx.OpenRecordEdit(migrationsTable, row.ID)
		e.Editor().Set(migrationDuration, 456*time.Nanosecond)
		tx.SaveRecord(e)
		got := edb.Get[mvpm.MigrationRecord](tx, row.ID)
		if got.ID != row.ID || got.Name != row.Name || !got.ExecutionTime.Equal(now) || got.ExecutionDuration != 456*time.Nanosecond {
			t.Fatal(got)
		}
	})
}
