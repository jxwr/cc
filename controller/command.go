package controller

type Result interface{}

type Command interface {
	Execute(*Controller) (Result, error)
}
