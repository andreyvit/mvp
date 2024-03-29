package mvp

import (
	"log"
	"os"
	"path/filepath"
	"reflect"

	"github.com/andreyvit/edb"
	"github.com/andreyvit/mvp/flake"
	mvpm "github.com/andreyvit/mvp/mvpmodel"
)

var flakeIDType = reflect.TypeOf(flake.ID(0))

func FlakeIDType() reflect.Type {
	return flakeIDType
}

func initAppDB(app *App, opt *AppOptions) {
	app.gen = flake.NewGen(0, 0)

	if app.Settings.DataDir == "" {
		app.Settings.DataDir = must(os.MkdirTemp("", "testdb*"))
	}
	ensure(os.MkdirAll(app.Settings.DataDir, 0755))
	app.db = must(edb.Open(filepath.Join(app.Settings.DataDir, "bolt.db"), &app.DBSchema, edb.Options{
		Logf:      log.Printf,
		Verbose:   app.Settings.VerboseDB,
		IsTesting: false,
	}))
}

func closeAppDB(app *App) {
	app.db.Close()
}

func (app *App) DB() *edb.DB {
	return app.db
}

func (app *App) NewID() flake.ID {
	return app.gen.New()
}

func (rc *RC) DBTx() *edb.Tx {
	if rc.tx == nil {
		rc.tx = rc.app.db.BeginRead()
	}
	return rc.tx
}

func (rc *RC) InTx(affinity mvpm.StoreAffinity, f func() error) error {
	if !affinity.WantsAutomaticTx() {
		return f()
	}
	if rc.tx != nil {
		if affinity.IsWriter() && !rc.tx.IsWritable() {
			rc.DoneReading()
		} else {
			return f()
		}
	}
	isWrite := affinity.IsWriter()
	err := rc.app.db.Tx(isWrite, func(tx *edb.Tx) error {
		rc.tx = tx
		tx.OnChange(rc.app.dbMonitoringOptions, rc.onDBChange)
		defer func() {
			rc.tx = nil
		}()
		return f()
	})
	if isWrite {
		rc.handleWriteTxEnded()
	}
	return err
}

func (rc *RC) IsInWriteTx() bool {
	return rc.tx != nil && rc.tx.IsWritable()
}

func (rc *RC) handleWriteTxEnded() {
	rc.applyDelayedCacheBusting()
}

// func (rc *RC) Commit() error {
// 	if rc.tx == nil {
// 		return nil
// 	}
// 	var errs error
// 	if rc.tx.IsWritable() {
// 		err := rc.tx.Commit()
// 		if err != nil {
// 			errs = err
// 		}
// 	}
// 	rc.tx.Close()
// 	rc.tx = nil
// 	return errs
// }

func (rc *RC) DoneReading() {
	if rc.tx == nil {
		return
	}
	if rc.tx.IsWritable() {
		panic("DoneReading on a writable transaction")
	}
	rc.tx.Close()
	rc.tx = nil
}

func (rc *RC) TryRead(f func() error) error {
	return rc.InTx(mvpm.SafeReader, f)
}
func (rc *RC) TryWrite(f func() error) error {
	return rc.InTx(mvpm.SafeWriter, f)
}
func (rc *RC) MustRead(f func()) {
	ensure(rc.InTx(mvpm.SafeReader, func() error {
		f()
		return nil
	}))
}
func (rc *RC) MustWrite(f func()) {
	ensure(rc.InTx(mvpm.SafeWriter, func() error {
		f()
		return nil
	}))
}

func (app *App) SetNewKeyOnRow(row any) bool {
	tbl := app.DBSchema.TableByRow(row)
	key := runHooksFwd2A(app.Hooks.makeRowKey, app, tbl)
	if key == nil {
		return false
	}
	tbl.SetRowKey(row, key)
	return true
}
