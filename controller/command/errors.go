package command

import (
	"errors"
)

var (
	ErrNodeNotFree             = errors.New("this is not a free node")
	ErrNodeIsFree              = errors.New("this is a free node")
	ErrNodeNotExist            = errors.New("node not exist")
	ErrNodeNotEmpty            = errors.New("node not empty")
	ErrNodeNotMaster           = errors.New("node is not master")
	ErrMigrateTaskNotExist     = errors.New("migration task not exist")
	ErrClusterSnapshotNotReady = errors.New("cluster snapshot not ready")
)
