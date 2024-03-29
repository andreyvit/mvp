package mvp

import (
	"github.com/andreyvit/edb"
)

func (rc *RC) onDBChange(tx *edb.Tx, chg *edb.Change) {
	runHooksFwd2(rc.app.Hooks.dbChange, rc, chg)
}
