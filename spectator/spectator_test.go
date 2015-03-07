package spectator

import (
	"fmt"
	"testing"

	"github.com/jxwr/cc/topo"
)

func TestBuildClusterTopo(t *testing.T) {
	s0 := topo.NewNode("127.0.0.1", 7000)
	s1 := topo.NewNode("127.0.0.1", 7002)

	sp := NewSpectator([]*topo.Node{s0, s1})
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

func TestFetchReploff(t *testing.T) {
	s0 := topo.NewNode("127.0.0.1", 7000)
	s1 := topo.NewNode("127.0.0.1", 7002)

	sp := NewSpectator([]*topo.Node{s0, s1})
	sp.BuildClusterTopo()
	cluster, _ := sp.BuildClusterTopo()

	rs := cluster.FindReplicaSetByNode("12eaf28f625c24735c45c631cd82822d658c26b8")
	if rs != nil {
		rmap := sp.FetchReplOffsetInReplicaSet(rs)
		for nodeId, offset := range rmap {
			fmt.Println(nodeId, offset)
		}
	}

	nodeId, err := sp.MaxReploffSlibing("12eaf28f625c24735c45c631cd82822d658c26b8", true)
	fmt.Println(nodeId, err)
}
