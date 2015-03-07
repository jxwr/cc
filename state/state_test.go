package state

import (
	"fmt"
	"log"
	"testing"

	"github.com/jxwr/cc/topo"
)

var (
	runningState = &State{
		Name: StateRunning,
		OnEnter: func() {
			log.Println("enter RUNNING state")
		},
		OnLeave: func() {
			log.Println("leave RUNNING state")
		},
	}

	waitFailoverBeginState = &State{
		Name: StateWaitFailoverBegin,
		OnEnter: func() {
			log.Println("enter WAIT_FAILOVE_BEGIN state")
		},
		OnLeave: func() {
			log.Println("leave WAIT_FAILOVER_BEGIN state")
		},
	}

	waitFailoverEndState = &State{
		Name: StateWaitFailoverEnd,
		OnEnter: func() {
			log.Println("enter WAIT_FAILOVE_END state")
		},
		OnLeave: func() {
			log.Println("leave WAIT_FAILOVER_END state")
		},
	}

	offlineState = &State{
		Name: StateOffline,
		OnEnter: func() {
			log.Println("enter OFFLINE state")
		},
		OnLeave: func() {
			log.Println("leave OFFLINE state")
		},
	}
)

func TestAddTransition(t *testing.T) {
	model := NewStateModel()

	model.AddState(runningState)
	model.AddState(waitFailoverBeginState)
	model.AddState(waitFailoverEndState)
	model.AddState(offlineState)

	/// State: (WaitFailoverRunning)

	// RUNNING >>>(F,F,*,*,*)>>> OFFLINE
	model.AddTransition(&Transition{
		From:       StateRunning,
		To:         StateOffline,
		Input:      Input{F, F, ANY, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply: func(node *NodeState, cs *ClusterState) {
			log.Println("apply")
		},
	})

	// RUNNING >>>(T,*,FAIL,*,*)>>> WAIT_FAILOVER_BEGIN
	model.AddTransition(&Transition{
		From:       StateRunning,
		To:         StateWaitFailoverBegin,
		Input:      Input{T, ANY, FAIL, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply: func(node *NodeState, cs *ClusterState) {
			log.Println("apply")
		},
	})

	// RUNNING >>>(*,T,FAIL,*,*)>>> WAIT_FAILOVER_BEGIN
	model.AddTransition(&Transition{
		From:       StateRunning,
		To:         StateWaitFailoverBegin,
		Input:      Input{ANY, T, FAIL, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply: func(node *NodeState, cs *ClusterState) {
			log.Println("apply")
		},
	})

	// RUNNING >>>(T,*,FAIL,*,*)&&Constraint>>> WAIT_FAILOVER_END
	model.AddTransition(&Transition{
		From:       StateRunning,
		To:         StateWaitFailoverEnd,
		Input:      Input{T, ANY, FAIL, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply: func(node *NodeState, cs *ClusterState) {
			log.Println("apply")
		},
	})

	// RUNNING >>>(*,T,FAIL,*,*)&&Constraint>>> WAIT_FAILOVER_END
	model.AddTransition(&Transition{
		From:       StateRunning,
		To:         StateWaitFailoverEnd,
		Input:      Input{ANY, T, FAIL, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply: func(node *NodeState, cs *ClusterState) {
			log.Println("apply")
		},
	})

	/// State: (WaitFailoverBegin)

	// WAIT_FAILOVER_BEGIN >>>(*,*,FINE,*,*)>>> RUNNING
	model.AddTransition(&Transition{
		From:       StateWaitFailoverBegin,
		To:         StateRunning,
		Input:      Input{ANY, ANY, FINE, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply: func(node *NodeState, cs *ClusterState) {
			log.Println("apply")
		},
	})

	model.AddTransition(&Transition{
		From:       StateWaitFailoverBegin,
		To:         StateWaitFailoverEnd,
		Input:      Input{ANY, ANY, FAIL, M, CMD_FAILOVER_BEGIN_SIGNAL},
		Priority:   0,
		Constraint: nil,
		Apply: func(node *NodeState, cs *ClusterState) {
			log.Println("apply")
		},
	})

	model.AddTransition(&Transition{
		From:       StateWaitFailoverBegin,
		To:         StateWaitFailoverEnd,
		Input:      Input{ANY, ANY, FAIL, S, CMD_FAILOVER_BEGIN_SIGNAL},
		Priority:   0,
		Constraint: nil,
		Apply: func(node *NodeState, cs *ClusterState) {
			log.Println("apply")
		},
	})

	model.AddTransition(&Transition{
		From:       StateWaitFailoverBegin,
		To:         StateOffline,
		Input:      Input{ANY, ANY, FAIL, S, ANY},
		Priority:   1,
		Constraint: nil,
		Apply: func(node *NodeState, cs *ClusterState) {
			log.Println("apply")
		},
	})

	/// State: (WaitFailoverEnd)

	model.AddTransition(&Transition{
		From:       StateWaitFailoverEnd,
		To:         StateOffline,
		Input:      Input{ANY, ANY, ANY, ANY, CMD_FAILOVER_END_SIGNAL},
		Priority:   1,
		Constraint: nil,
		Apply: func(node *NodeState, cs *ClusterState) {
			log.Println("apply")
		},
	})

	/// State: (Offline)

	model.AddTransition(&Transition{
		From:       StateOffline,
		To:         StateRunning,
		Input:      Input{T, ANY, ANY, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply: func(node *NodeState, cs *ClusterState) {
			log.Println("apply")
		},
	})

	model.AddTransition(&Transition{
		From:       StateOffline,
		To:         StateRunning,
		Input:      Input{ANY, T, ANY, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply: func(node *NodeState, cs *ClusterState) {
			log.Println("apply")
		},
	})

	model.AddTransition(&Transition{
		From:       StateOffline,
		To:         StateWaitFailoverBegin,
		Input:      Input{F, F, FAIL, M, ANY},
		Priority:   0,
		Constraint: nil,
		Apply: func(node *NodeState, cs *ClusterState) {
			log.Println("apply")
		},
	})

	/// Init state machine

	fsm := NewStateMachine(StateRunning, model)

	log.Println(fsm.CurrentState())

	inputQueue := []Input{
		Input{T, T, FAIL, M, CMD_NONE},
		Input{T, T, FAIL, M, CMD_NONE},
		Input{T, T, FAIL, M, CMD_FAILOVER_BEGIN_SIGNAL},
		Input{T, T, FAIL, M, CMD_FAILOVER_BEGIN_SIGNAL},
		Input{T, T, FAIL, M, CMD_FAILOVER_END_SIGNAL},
		Input{T, T, FAIL, M, CMD_FAILOVER_END_SIGNAL},
		Input{T, T, FAIL, M, CMD_FAILOVER_END_SIGNAL},
		Input{T, T, FAIL, M, CMD_FAILOVER_END_SIGNAL},
		Input{F, F, FAIL, S, CMD_NONE},
		Input{F, F, FAIL, S, CMD_NONE},
		Input{F, F, FAIL, M, CMD_NONE},
		Input{F, F, FINE, M, CMD_NONE},
		Input{F, F, FINE, M, CMD_NONE},
		Input{F, F, FINE, M, CMD_NONE},
	}

	for _, input := range inputQueue {
		fsm.Advance(nil, nil, input)
	}

	model.DumpTransitions()
}

func TestClusterUpdateRegionNodes(t *testing.T) {
	cs := NewClusterState()

	ss := []*topo.Server{
		topo.NewServer("127.0.0.1", 7000).SetId("7000").SetRegion("bj").SetRole("master"),
		topo.NewServer("127.0.0.1", 7001).SetId("7001").SetRegion("bj").SetRole("slave"),
	}
	cs.UpdateRegionNodes("bj", ss)

	cmds := []InputField{
		CMD_NONE,
		CMD_FAILOVER_BEGIN_SIGNAL,
		CMD_NONE,
		CMD_NONE,
		CMD_FAILOVER_BEGIN_SIGNAL,
		CMD_NONE,
		CMD_NONE,
		CMD_FAILOVER_END_SIGNAL,
	}

	for _, cmd := range cmds {
		for _, s := range ss {
			node := cs.FindNode(s.Id())
			node.AdvanceFSM(cs, cmd)
		}
		fmt.Println("========= Handle", cmd)
		cs.DebugDump()
	}
}
