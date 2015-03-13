package command

import (
	"errors"
)

var (
	ErrNodeNotExist            = errors.New("node not exist")
	ErrMigrateTaskNotExist     = errors.New("migration task not exist")
	ErrClusterSnapshotNotReady = errors.New("cluster snapshot not ready")
)
