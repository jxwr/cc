package state

import (
	"fmt"

	"github.com/jxwr/cc/fsm"
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

/// Node State

type NodeState struct {
	version int64
	server  *topo.Server
	fsm     *fsm.StateMachine
}

func NewNodeState(server *topo.Server, version int64) *NodeState {
	ns := &NodeState{
		version: version,
		server:  server,
		fsm:     fsm.NewStateMachine(StateRunning, RedisNodeStateModel),
	}
	return ns
}

func (ns *NodeState) Addr() string {
	return ns.server.Addr()
}

func (ns *NodeState) Id() string {
	return ns.server.Id()
}

func (ns *NodeState) AdvanceFSM(cs *ClusterState, cmd InputField) error {
	// 构造Input五元组
	s := ns.server
	r := F
	if s.Readable() {
		r = T
	}
	w := F
	if s.Writable() {
		w = T
	}
	fail := FINE
	if s.Fail() {
		fail = FAIL
	}
	role := S
	if s.IsMaster() {
		role = M
	}
	input := Input{r, w, fail, role, cmd}

	// 创建状态转换Context
	ctx := StateContext{
		Input:        input,
		ClusterState: cs,
		NodeState:    ns,
	}
	ns.fsm.Advance(ctx, input)
	ns.DebugDump()
	return nil
}

func (ns *NodeState) DebugDump() {
	s := ns.server
	fmt.Printf("%s %d %s %v (%v,%v,%v,%v))\n",
		s.Id(), ns.version, s.Addr(), ns.fsm.CurrentState(),
		s.Readable(), s.Writable(), s.Fail(), s.Role())
}
