package state

import (
	"fmt"

	"github.com/jxwr/cc/topo"
)

type ClusterState struct {
	version    int64
	nodeStates map[string]*NodeState
}

func NewClusterState() *ClusterState {
	cs := &ClusterState{
		version:    0,
		nodeStates: map[string]*NodeState{},
	}
	return cs
}

func (cs *ClusterState) AllNodeStates() map[string]*NodeState {
	return cs.nodeStates
}

func (cs *ClusterState) UpdateRegionNodes(region string, nodes []*topo.Node) {
	cs.version++

	// 添加不存在的节点，版本号+1
	for _, n := range nodes {
		if n.Region() != region {
			continue
		}
		nodeState := cs.nodeStates[n.Id()]
		if nodeState == nil {
			nodeState = NewNodeState(n, cs.version)
			cs.nodeStates[n.Id()] = nodeState
		} else {
			nodeState.version = cs.version
			nodeState.node = n
		}
	}

	// 删除已经下线的节点
	for id, n := range cs.nodeStates {
		if n.node.Region() != region {
			continue
		}
		nodeState := cs.nodeStates[id]
		if nodeState.version != cs.version {
			delete(cs.nodeStates, id)
		}
	}
}

func (cs *ClusterState) FindNode(nodeId string) *NodeState {
	return cs.nodeStates[nodeId]
}

func (cs *ClusterState) DebugDump() {
	fmt.Println("Cluster Debug Information:")
	for _, ns := range cs.nodeStates {
		fmt.Print("  ")
		ns.DebugDump()
	}
}
