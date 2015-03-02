package main

import (
	"log"
)

const (
	StateStarting          ClusterState = 1 + iota // 初始状态，具体状态待判定
	StatePrebuild                                  // 集群未初始化
	StateBuilding                                  // 自动构建集群状态
	StateRunning                                   // 正常运行
	StateFailover                                  // 正在进行Failover
	StateSlaveDead                                 // 存在Slave处于FAIL状态
	StateSlaveAlive                                // 某处于FAIL状态的Slave重新可用
	StateMasterDead                                // 存在Master处于FAIL状态
	StateAutoRebalancing                           // 正在进行自动Rebalance
	StateManualRebalancing                         // 正在进行人工Rebalance
)

//type StateInput
//type StateOutput
//type StateTransition

type ClusterState int
type TransFunc func(*ClusterStateMachine)

type ClusterStateMachine struct {
	transTable map[ClusterState]TransFunc
	manager    *ClusterManager
	state      ClusterState
	event      StateEvent
}

func NewStateMachine(manager *ClusterManager) *ClusterStateMachine {
	sm := &ClusterStateMachine{
		manager:    manager,
		state:      StateStarting,
		transTable: map[ClusterState]TransFunc{},
	}

	sm.transTable[StateStarting] = (*ClusterStateMachine).OnStarting
	sm.transTable[StatePrebuild] = (*ClusterStateMachine).OnPrebuild
	sm.transTable[StateBuilding] = (*ClusterStateMachine).OnBuilding
	sm.transTable[StateRunning] = (*ClusterStateMachine).OnRunning
	sm.transTable[StateSlaveDead] = (*ClusterStateMachine).OnSlaveDead
	sm.transTable[StateSlaveAlive] = (*ClusterStateMachine).OnSlaveAlive
	sm.transTable[StateMasterDead] = (*ClusterStateMachine).OnMasterDead
	sm.transTable[StateFailover] = (*ClusterStateMachine).OnFailover
	sm.transTable[StateAutoRebalancing] = (*ClusterStateMachine).OnAutoRebalancing

	return sm
}

func (self *ClusterStateMachine) Run() {
	for {
		transFunc := self.transTable[self.state]
		if transFunc == nil {
			log.Fatalf("nil trans func %v\n", self.state)
		}
		transFunc(self)
	}
}

func (self *ClusterStateMachine) OnStarting() {
	log.Println("Enter StateStarting")
	self.state = StateRunning
}

func (self *ClusterStateMachine) OnPrebuild() {
	event := self.manager.PollStateEvent()
	if event.Type() == EventBuild {
		self.state = StateBuilding
	} else {
		log.Printf("invalid state transformation PRBUILD -> %v", event.Type())
	}
}

func (self *ClusterStateMachine) OnBuilding() {
	err := self.manager.BuildCluster()
	if err != nil {
		log.Printf("build cluster failed %v", err)
		self.state = StatePrebuild
	} else {
		self.state = StateRunning
	}
}

func (self *ClusterStateMachine) OnRunning() {
	log.Println("Enter StateRunning")

	event := self.manager.PollStateEvent()
	self.event = event

	switch event.Type() {
	case EventSlaveDead:
		self.state = StateSlaveDead
	case EventMasterDead:
		self.state = StateMasterDead
	case EventSlaveAlive:
		self.state = StateSlaveAlive
	case EventAutoRebalance:
		self.state = StateAutoRebalancing
	default:
		log.Printf("invalid state transformation RUNNING -> %v", event.Type())
	}
}

func (self *ClusterStateMachine) OnSlaveDead() {
	e := self.event.(*SlaveDeadEvent)
	self.manager.DisableServerRead(e.Server)
}

func (self *ClusterStateMachine) OnSlaveAlive() {
	e := self.event.(*SlaveAliveEvent)
	self.manager.EnableServerRead(e.Server)
}

func (self *ClusterStateMachine) OnMasterDead() {
	e := self.event.(*MasterDeadEvent)
	self.manager.DisableServerRead(e.Server)

	// 不自动触发Failover，要么该故障的Master自动恢复，要么外部执行Failover
	event := self.manager.PollStateEvent()

	if event.Type() == EventMasterAlive {
		self.manager.EnableServerRead(e.Server)
		self.state = StateRunning
		return
	}

	if event.Type() == EventFailover {
		self.state = StateFailover
	}
}

func (self *ClusterStateMachine) OnFailover() {
	e := self.event.(*MasterDeadEvent)
	err := self.manager.HandleFailover(e.Server)
	if err != nil {
		log.Printf("failover failed %v\n", err)
	} else {
		self.state = StateRunning
	}
}

func (self *ClusterStateMachine) OnAutoRebalancing() {
	err := self.manager.AutoRebalance()
	if err != nil {
		log.Printf("auto rebalance failed %v\n", err)
	} else {
		self.state = StateRunning
	}
}
