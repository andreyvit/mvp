package mvp

import (
	"reflect"
	"time"

	mvpm "github.com/andreyvit/mvp/mvpmodel"

	"github.com/andreyvit/edb/kvo"
	"github.com/andreyvit/edb/kvo/codec/msgpack"
	"github.com/andreyvit/edb/values"
	"github.com/andreyvit/mvp/mvpjobs"
)

var builtinModels = kvo.NewSchema()
var jobKind = kvo.NewStringProp[string](builtinModels, 1, "job:kind", kvo.TMachineString)
var jobName = kvo.NewStringProp[string](builtinModels, 2, "job:name", kvo.TMachineString)
var jobStatus = kvo.NewSignedProp[mvpjobs.Status](builtinModels, 3, "job:status", kvo.NewIntStringerType[mvpjobs.Status]("job_status"))
var jobNextRunTime = kvo.NewTimeProp(builtinModels, 4, "job:next_run_time")
var jobStartTime = kvo.NewTimeProp(builtinModels, 5, "job:start_time")

// The projection excludes job parameters/state; indexing never decodes them.
var jobModel = func() *kvo.Model {
	// Status has a text MessagePack representation; retain its parser's rules.
	kvo.RegisterMessagePackScalar(reflect.TypeFor[mvpjobs.Status](), func(v msgpack.View) values.Value {
		if v.Nil() {
			return values.Value{}
		}
		status, err := mvpjobs.ParseStatus(values.Bytes(v.Bytes()).AsString())
		if err != nil {
			panic(err)
		}
		return values.Int(int64(status))
	})
	model := kvo.NewModel(builtinModels, kvo.NewEntityType(builtinModels, "Job"), func(b *kvo.ModelBuilder) {
		b.Prop(jobKind)
		b.Prop(jobName)
		b.Prop(jobStatus)
		b.Prop(jobNextRunTime)
		b.Prop(jobStartTime)
	})
	model.RegisterNativeType(reflect.TypeFor[mvpjobs.Job](), map[kvo.PropCode]string{
		jobKind.Code(): "Kind", jobName.Code(): "Name", jobStatus.Code(): "Status", jobNextRunTime.Code(): "NextRunTime", jobStartTime.Code(): "StartTime",
	})
	return model
}()

var migrationName = kvo.NewStringProp[string](builtinModels, 6, "migration:name", kvo.TMachineString)
var migrationTime = kvo.NewTimeProp(builtinModels, 7, "migration:execution_time")
var migrationDuration = kvo.NewSignedProp[time.Duration](builtinModels, 8, "migration:execution_duration", kvo.TInt64)
var migrationModel = kvo.NewModel(builtinModels, kvo.NewEntityType(builtinModels, "MigrationRecord", reflect.TypeFor[mvpm.MigrationRecord]()), func(b *kvo.ModelBuilder) {
	b.Prop(migrationName)
	b.Prop(migrationTime)
	b.Prop(migrationDuration)
})
