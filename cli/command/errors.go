package command

import (
	"errors"
)

var (
	ErrInvalidParameter     = errors.New("Command failed: invalid parameter")
	ErrNoNodesFound         = errors.New("Command failed: no nodes found")
	ErrMoreThanOneNodeFound = errors.New("Command failed: more than one node found")
)
