package state

import (
	"log"
	"testing"
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
		Input:      Msg{F, F, ANY, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply: func() {
			log.Println("apply")
		},
	})

	// RUNNING >>>(T,*,FAIL,*,*)>>> WAIT_FAILOVER_BEGIN
	model.AddTransition(&Transition{
		From:       StateRunning,
		To:         StateWaitFailoverBegin,
		Input:      Msg{T, ANY, FAIL, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply: func() {
			log.Println("apply")
		},
	})

	// RUNNING >>>(*,T,FAIL,*,*)>>> WAIT_FAILOVER_BEGIN
	model.AddTransition(&Transition{
		From:       StateRunning,
		To:         StateWaitFailoverBegin,
		Input:      Msg{ANY, T, FAIL, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply: func() {
			log.Println("apply")
		},
	})

	// RUNNING >>>(T,*,FAIL,*,*)&&Constraint>>> WAIT_FAILOVER_END
	model.AddTransition(&Transition{
		From:       StateRunning,
		To:         StateWaitFailoverEnd,
		Input:      Msg{T, ANY, FAIL, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply: func() {
			log.Println("apply")
		},
	})

	// RUNNING >>>(*,T,FAIL,*,*)&&Constraint>>> WAIT_FAILOVER_END
	model.AddTransition(&Transition{
		From:       StateRunning,
		To:         StateWaitFailoverEnd,
		Input:      Msg{ANY, T, FAIL, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply: func() {
			log.Println("apply")
		},
	})

	/// State: (WaitFailoverBegin)

	// WAIT_FAILOVER_BEGIN >>>(*,*,FINE,*,*)>>> RUNNING
	model.AddTransition(&Transition{
		From:       StateWaitFailoverBegin,
		To:         StateRunning,
		Input:      Msg{ANY, ANY, FINE, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply: func() {
			log.Println("apply")
		},
	})

	// WAIT_FAILOVER_BEGIN >>>(*,*,FAIL,M,CMD_FAILOVER)>>> WAIT_FAILOVER_END
	model.AddTransition(&Transition{
		From:       StateWaitFailoverBegin,
		To:         StateWaitFailoverEnd,
		Input:      Msg{ANY, ANY, FAIL, M, CMD_FAILOVER},
		Priority:   0,
		Constraint: nil,
		Apply: func() {
			log.Println("apply")
		},
	})

	// WAIT_FAILOVER_BEGIN >>>(*,*,FAIL,S,CMD_FAILOVER)>>> WAIT_FAILOVER_END
	model.AddTransition(&Transition{
		From:       StateWaitFailoverBegin,
		To:         StateWaitFailoverEnd,
		Input:      Msg{ANY, ANY, FAIL, S, CMD_FAILOVER},
		Priority:   0,
		Constraint: nil,
		Apply: func() {
			log.Println("apply")
		},
	})

	model.AddTransition(&Transition{
		From:       StateWaitFailoverBegin,
		To:         StateOffline,
		Input:      Msg{ANY, ANY, FAIL, S, ANY},
		Priority:   1,
		Constraint: nil,
		Apply: func() {
			log.Println("apply")
		},
	})

	/// State: (WaitFailoverEnd)

	model.AddTransition(&Transition{
		From:       StateWaitFailoverEnd,
		To:         StateOffline,
		Input:      Msg{ANY, ANY, ANY, ANY, CMD_FAILOVER_EXIT},
		Priority:   1,
		Constraint: nil,
		Apply: func() {
			log.Println("apply")
		},
	})

	/// State: (Offline)

	model.AddTransition(&Transition{
		From:       StateOffline,
		To:         StateRunning,
		Input:      Msg{T, ANY, ANY, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply: func() {
			log.Println("apply")
		},
	})

	model.AddTransition(&Transition{
		From:       StateOffline,
		To:         StateRunning,
		Input:      Msg{ANY, T, ANY, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply: func() {
			log.Println("apply")
		},
	})

	model.AddTransition(&Transition{
		From:       StateOffline,
		To:         StateWaitFailoverBegin,
		Input:      Msg{F, F, FAIL, M, ANY},
		Priority:   0,
		Constraint: nil,
		Apply: func() {
			log.Println("apply")
		},
	})

	/// Init state machine

	fsm := NewStateMachine(StateRunning, model)

	log.Println(fsm.CurrentState())

	msgQueue := []Msg{
		Msg{T, T, FAIL, M, CMD_NONE},
		Msg{T, T, FAIL, M, CMD_NONE},
		Msg{T, T, FAIL, M, CMD_FAILOVER},
		Msg{T, T, FAIL, M, CMD_FAILOVER},
		Msg{T, T, FAIL, M, CMD_FAILOVER_EXIT},
		Msg{T, T, FAIL, M, CMD_FAILOVER_EXIT},
		Msg{T, T, FAIL, M, CMD_FAILOVER_EXIT},
		Msg{T, T, FAIL, M, CMD_FAILOVER_EXIT},
		Msg{F, F, FAIL, S, CMD_NONE},
		Msg{F, F, FAIL, S, CMD_NONE},
		Msg{F, F, FAIL, M, CMD_NONE},
		Msg{F, F, FINE, M, CMD_NONE},
		Msg{F, F, FINE, M, CMD_NONE},
		Msg{F, F, FINE, M, CMD_NONE},
	}

	for _, msg := range msgQueue {
		fsm.Next(msg)
	}

	model.DumpTransitions()
}
