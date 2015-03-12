package command

import (
	"errors"
)

var (
	ErrNodeNotExist            = errors.New("node not exist")
	ErrClusterSnapshotNotReady = errors.New("cluster snapshot not ready")
)
