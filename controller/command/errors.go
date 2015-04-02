package command

import (
	"errors"
)

var (
	ErrNodeNotFree             = errors.New("this is not a free node")
	ErrNodeIsFree              = errors.New("this is a free node")
	ErrNodeNotExist            = errors.New("node not exist")
	ErrMigrateTaskNotExist     = errors.New("migration task not exist")
	ErrClusterSnapshotNotReady = errors.New("cluster snapshot not ready")
)
