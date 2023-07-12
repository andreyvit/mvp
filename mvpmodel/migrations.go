package mvpm

import (
	"time"

	"github.com/andreyvit/mvp/flake"
)

type (
	MigrationID = flake.ID

	MigrationRecord struct {
		ID                MigrationID   `msgpack:"-"`
		Name              string        `msgpack:"n"`
		ExecutionTime     time.Time     `msgpack:"@"`
		ExecutionDuration time.Duration `msgpack:"d"`
	}
)
