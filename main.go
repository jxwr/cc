package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	//	"github.com/jxwr/cc/topo"
)

func handleTopSnapshot(c *gin.Context) {
	c.String(200, "pong")
}

func main() {
	fmt.Println("here we go")
	/*
		        manager := NewClusterManager()
				stateMachine := NewStateMachine(manager)

				e1 := NewBuildEvent()
				manager.PushStateEvent(e1)

				e2 := NewSlaveDeadEvent(topo.NewServer("127.0.0.1", 7000))
				manager.PushStateEvent(e2)
	*/
	r := gin.Default()

	r.POST("/topo/snapshort", handleTopSnapshot)

	r.Run(":8080")
}
