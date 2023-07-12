package mvp

import (
	"fmt"
	"sort"
	"time"

	"github.com/andreyvit/edb"
	"github.com/andreyvit/mvp/flake"
	"github.com/andreyvit/mvp/flogger"
	mvpm "github.com/andreyvit/mvp/mvpmodel"
)

func collectMigrations(app *App, lc flogger.Context) []*migration {
	var migb = MigrationBuilder{
		lc:         lc,
		migrations: make([]*migration, 0, 100),
		ids:        make(map[flake.ID]string, 100),
	}
	runHooksFwd2(app.Hooks.migrate, app, &migb)
	return migb.build()
}

func executeMigrations(allMigrations []*migration, rc *RC) {
	pendingMigrations := make([]*migration, 0, len(allMigrations))
	for _, m := range allMigrations {
		if !rc.DBTx().Exists(migrationsTable, m.id) {
			pendingMigrations = append(pendingMigrations, m)
		}
	}

	if len(pendingMigrations) > 0 {
		flogger.Log(rc, "Running %d migrations...", len(pendingMigrations))
		for _, m := range pendingMigrations {
			executeMigration(m, rc)
		}
	}
}

func executeMigration(m *migration, rc *RC) {
	rc.RequestID = fmt.Sprintf("init:migration:%s", m.name)
	defer func() { rc.RequestID = "init" }()

	flogger.Log(rc, "running...")
	start := time.Now()
	m.handler(rc)
	elapsed := time.Since(start)
	edb.Put(rc, &mvpm.MigrationRecord{
		ID:                m.id,
		Name:              m.name,
		ExecutionTime:     rc.BaseApp().Now(),
		ExecutionDuration: elapsed,
	})
	flogger.Log(rc, "finished in %d µs (%d ms)", elapsed.Microseconds(), elapsed.Milliseconds())
}

type migration struct {
	id      mvpm.MigrationID
	name    string
	handler func(rc *RC)
}

type MigrationBuilder struct {
	lc         flogger.Context
	migrations []*migration
	ids        map[mvpm.MigrationID]string
}

func (b *MigrationBuilder) Run(id mvpm.MigrationID, name string, handler func(rc *RC)) {
	b.checkDup(id, name)
	b.migrations = append(b.migrations, &migration{id, name, handler})
}
func (b *MigrationBuilder) Draft(id mvpm.MigrationID, name string, handler func(rc *RC)) {
	b.checkDup(id, name)
	flogger.Log(b.lc, "skipping draft migration %s", name)
}
func (b *MigrationBuilder) Stale(id mvpm.MigrationID, name string, handler func(rc *RC)) {
	b.checkDup(id, name)
}
func (b *MigrationBuilder) checkDup(id mvpm.MigrationID, name string) {
	if name == "" {
		name = id.String()
	}
	if prev := b.ids[id]; prev != "" {
		panic(fmt.Errorf("migrations %q and %q use the same ID 0x%X", name, prev, id))
	}
	b.ids[id] = name
}
func (b *MigrationBuilder) build() []*migration {
	sort.Slice(b.migrations, func(i, j int) bool {
		return b.migrations[i].id < b.migrations[j].id
	})
	return b.migrations
}
