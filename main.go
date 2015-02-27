package main

import (
	"fmt"

	"github.com/jxwr/cc/topo"
)

func main() {
	fmt.Println("here we go")
	manager := NewClusterManager()
	stateMachine := NewStateMachine(manager)

	e1 := NewBuildEvent()
	manager.PushStateEvent(e1)

	e2 := NewSlaveDeadEvent(topo.NewServer("127.0.0.1", 7000))
	manager.PushStateEvent(e2)

	stateMachine.Run()
}
