package state

import (
	"fmt"
	"time"

	"github.com/jxwr/cc/fsm"
	"github.com/jxwr/cc/topo"
)

type NodeState struct {
	node       *topo.Node
	updateTime time.Time
	version    int64
	fsm        *fsm.StateMachine
}

func NewNodeState(node *topo.Node, version int64) *NodeState {
	ns := &NodeState{
		version: version,
		node:    node,
		fsm:     fsm.NewStateMachine(StateRunning, RedisNodeStateModel),
	}
	return ns
}

func (ns *NodeState) Addr() string {
	return ns.node.Addr()
}

func (ns *NodeState) Id() string {
	return ns.node.Id()
}

func (ns *NodeState) CurrentState() string {
	return ns.fsm.CurrentState()
}

func (ns *NodeState) AdvanceFSM(cs *ClusterState, cmd InputField) error {
	// 构造Input五元组
	s := ns.node
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
	s := ns.node
	fmt.Printf("%s %d %s %v (%v,%v,%v,%v))\n",
		s.Id(), ns.version, s.Addr(), ns.fsm.CurrentState(),
		s.Readable(), s.Writable(), s.Fail(), s.Role())
}
