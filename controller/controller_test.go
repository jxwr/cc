package controller_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/controller/command"
	"github.com/jxwr/cc/inspector"
	"github.com/jxwr/cc/topo"
)

func TestUpdateRegion(t *testing.T) {
	c := controller.NewController()

	s0 := topo.NewNode("127.0.0.1", 7000)
	s1 := topo.NewNode("127.0.0.1", 7002)

	sp := inspector.NewInspector([]*topo.Node{s0, s1})

	go func() {
		time.Sleep(10 * time.Second)
		cmd := &command.MigrateCommand{
			"1c11d8d88e7d2ac9e0bb9bb1a1208e06468cd9e0",
			"5f674075196119c0d94037583b8a4a9a0e902dd5",
			[]topo.Range{topo.Range{6010, 6020}},
		}
		fmt.Println("=====", "migrate command", "=====")
		c.ProcessCommand(cmd, 2*time.Second)
	}()

	for {
		time.Sleep(1 * time.Second)
		clusterTopo, err := sp.BuildClusterTopo()
		if err != nil {
			fmt.Println(err)
			continue
		}
		ss := clusterTopo.LocalRegionNodes()
		fmt.Println("=================", clusterTopo.Region())
		cmd := command.UpdateRegionCommand{clusterTopo.Region(), ss}
		c.ProcessCommand(cmd, 5*time.Second)
	}
}
