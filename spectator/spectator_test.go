package spectator

import (
	"fmt"
	"testing"

	"github.com/jxwr/cc/topo"
)

func TestBuildClusterTopo(t *testing.T) {
	s0 := topo.NewServer("127.0.0.1", 7000)
	s1 := topo.NewServer("127.0.0.1", 7002)

	sp := NewSpectator([]*topo.Server{s0, s1})
	sp.BuildClusterTopo()
	cluster, err := sp.BuildClusterTopo()

	if err == nil {
		ss := cluster.FailureServers()
		fmt.Println(cluster, len(ss), len(cluster.RegionServers()), err)
	} else {
		fmt.Println(err)
	}

	sp.Run()
}
