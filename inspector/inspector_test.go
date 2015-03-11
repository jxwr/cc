package inspector

import (
	"fmt"
	"testing"

	"github.com/jxwr/cc/topo"
)

func TestBuildClusterTopo(t *testing.T) {
	s0 := topo.NewNode("127.0.0.1", 7000)
	s1 := topo.NewNode("127.0.0.1", 7002)

	sp := NewInspector([]*topo.Node{s0, s1})
	sp.BuildClusterTopo()
	cluster, err := sp.BuildClusterTopo()

	if err == nil {
		ss := cluster.FailureNodes()
		fmt.Println(cluster, len(ss), len(cluster.LocalRegionNodes()), err)
	} else {
		fmt.Println(err)
	}

	//go sp.Run()
}
