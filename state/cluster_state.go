package state

import (
	"fmt"

	"github.com/jxwr/cc/topo"
)

type ClusterState struct {
	version int64
	nodes   map[string]*NodeState
}

func NewClusterState() *ClusterState {
	cs := &ClusterState{
		version: 0,
		nodes:   map[string]*NodeState{},
	}
	return cs
}

func (cs *ClusterState) AllNodes() map[string]*NodeState {
	return cs.nodes
}

func (cs *ClusterState) UpdateRegionNodes(region string, servers []*topo.Server) {
	cs.version++

	// 添加不存在的节点，版本号+1
	for _, s := range servers {
		if s.Region() != region {
			continue
		}
		node := cs.nodes[s.Id()]
		if node == nil {
			node = NewNodeState(s, cs.version)
			cs.nodes[s.Id()] = node
		} else {
			node.version = cs.version
			node.server = s
		}
	}

	// 删除已经下线的节点
	for id, n := range cs.nodes {
		if n.server.Region() != region {
			continue
		}
		node := cs.nodes[id]
		if node.version != cs.version {
			delete(cs.nodes, id)
		}
	}
}

func (cs *ClusterState) FindNode(nodeId string) *NodeState {
	return cs.nodes[nodeId]
}

func (cs *ClusterState) DebugDump() {
	fmt.Println("Cluster Debug Information:")
	for _, node := range cs.nodes {
		fmt.Print("  ")
		node.DebugDump()
	}
}
