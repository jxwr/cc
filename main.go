package main

import (
	"fmt"

	"github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/frontend"
	"github.com/jxwr/cc/spectator"
	"github.com/jxwr/cc/topo"
)

func main() {
	fmt.Println("here we go")

	c := controller.NewController()

	s0 := topo.NewNode("127.0.0.1", 7000)
	s1 := topo.NewNode("127.0.0.1", 7002)

	sp := spectator.NewSpectator([]*topo.Node{s0, s1})
	go sp.ReportRegionSnanshotLoop()

	fe := frontend.NewFrontEnd(c, ":6000")
	fe.Run()
}
