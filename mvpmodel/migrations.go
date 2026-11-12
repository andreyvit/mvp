package mvpm

import (
	"time"

	"github.com/andreyvit/mvp/flake"
)

type (
	MigrationID = flake.ID

	MigrationRecord struct {
		ID                MigrationID   `msgpack:"-" kvo:"-"`
		Name              string        `msgpack:"n" kvo:"name"`
		ExecutionTime     time.Time     `msgpack:"@" kvo:"execution_time"`
		ExecutionDuration time.Duration `msgpack:"d" kvo:"execution_duration"`
	}
)
