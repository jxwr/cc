package command

import (
	"errors"
)

var (
	ErrInvalidParameter = errors.New("Command failed: invalid parameter")
)
