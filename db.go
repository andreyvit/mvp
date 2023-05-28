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

func (app *App) InTx(rc *RC, affinity mvpm.StoreAffinity, f func() error) error {
	if affinity == mvpm.DBUnused {
		return f()
	}
	if rc.tx != nil {
		if affinity.IsWriter() && !rc.tx.IsWritable() {
			panic("cannot initiate a mutating tx from read-only one")
		}
		return f()
	} else {
		return app.db.Tx(affinity.IsWriter(), func(tx *edb.Tx) error {
			rc.tx = tx
			defer func() {
				rc.tx = nil
			}()
			return f()
		})
	}
}

func (app *App) Read(rc *RC, f func() error) error {
	return app.InTx(rc, mvpm.SafeReader, f)
}
func (app *App) Write(rc *RC, f func() error) error {
	return app.InTx(rc, mvpm.SafeWriter, f)
}
func (app *App) MustRead(rc *RC, f func()) {
	ensure(app.InTx(rc, mvpm.SafeReader, func() error {
		f()
		return nil
	}))
}
func (app *App) MustWrite(rc *RC, f func()) {
	ensure(app.InTx(rc, mvpm.SafeWriter, func() error {
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
