package controller_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/controller/command"
	"github.com/jxwr/cc/spectator"
	"github.com/jxwr/cc/topo"
)

func TestUpdateRegion(t *testing.T) {
	c := controller.NewController()

	s0 := topo.NewServer("127.0.0.1", 7000)
	s1 := topo.NewServer("127.0.0.1", 7002)

	sp := spectator.NewSpectator([]*topo.Server{s0, s1})

	go func() {
		for {
			cmd := &command.FailoverBeginCommand{"8e05f3ec5ab3b21da8337bb6519124847a93fc3f"}
			fmt.Println("=====", "send failover begin", "=====")
			time.Sleep(1 * time.Second)
			c.ProcessCommand(cmd, 2*time.Second)
		}
	}()

	for {
		time.Sleep(1 * time.Second)
		clusterTopo, err := sp.BuildClusterTopo()
		if err != nil {
			fmt.Println(err)
			continue
		}
		ss := clusterTopo.RegionServers()
		fmt.Println("=================", clusterTopo.Region())
		for _, s := range ss {
			fmt.Println(s.Id(), s.Addr(), s.Fail(), s.Readable(), s.Writable(), s.Role())
		}

		cmd := command.UpdateRegionCommand{clusterTopo.Region(), ss}
		c.ProcessCommand(cmd, 5*time.Second)
	}
}
