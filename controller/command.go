package controller

type Result interface{}

type CommandType int

const (
	REGION_COMMAND CommandType = iota
	CLUSTER_COMMAND
)

type Command interface {
	Type() CommandType
	Execute(*Controller) (Result, error)
}
