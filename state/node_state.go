package state

import (
	"fmt"
	"sync"
	"time"

	"github.com/ksarch-saas/cc/fsm"
	"github.com/ksarch-saas/cc/streams"
	"github.com/ksarch-saas/cc/topo"
)

type NodeState struct {
	node       *topo.Node        // 节点静态信息
	updateTime time.Time         // 最近一次更新时间
	version    int64             // 更新次数
	fsm        *fsm.StateMachine // 节点状态机
	mutex      *sync.Mutex
}

func NewNodeState(node *topo.Node, version int64) *NodeState {
	ns := &NodeState{
		version: version,
		node:    node,
		fsm:     fsm.NewStateMachine(StateRunning, RedisNodeStateModel),
		mutex:   &sync.Mutex{},
	}
	return ns
}

func (ns *NodeState) Addr() string {
	return ns.node.Addr()
}

func (ns *NodeState) Id() string {
	return ns.node.Id
}

func (ns *NodeState) Region() string {
	return ns.node.Region
}

func (ns *NodeState) Tag() string {
	return ns.node.Tag
}

func (ns *NodeState) Role() string {
	return ns.node.Role
}

func (ns *NodeState) Ranges() []topo.Range {
	return ns.node.Ranges
}

func (ns *NodeState) CurrentState() string {
	return ns.fsm.CurrentState()
}

func (ns *NodeState) Node() *topo.Node {
	return ns.node
}

func (ns *NodeState) AdvanceFSM(cs *ClusterState, cmd InputField) error {
	ns.mutex.Lock()
	defer ns.mutex.Unlock()

	// 构造Input五元组
	s := ns.node
	r := F
	if s.Readable {
		r = T
	}
	w := F
	if s.Writable {
		w = T
	}
	fail := FINE
	if s.Fail {
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

	// 发送状态用于记录和实时观察
	data := &streams.NodeStateStreamData{
		ns.node,
		ns.CurrentState(),
		ns.version,
	}
	streams.NodeStateStream.Pub(data)
	return nil
}

func (ns *NodeState) DebugDump() {
	s := ns.node
	fmt.Printf("%s %d %s %v (%v,%v,%v,%v))\n",
		s.Id, ns.version, s.Addr(), ns.fsm.CurrentState(),
		s.Readable, s.Writable, s.Fail, s.Role)
}
