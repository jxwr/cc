package state

import (
	"log"

	"github.com/jxwr/cc/fsm"
	"github.com/jxwr/cc/redis"
)

const (
	StateRunning           = "RUNNING"
	StateWaitFailoverBegin = "WAIT_FAILOVER_BEGIN"
	StateWaitFailoverEnd   = "WAIT_FAILOVER_END"
	StateOffline           = "OFFLINE"
)

var (
	RunningState = &fsm.State{
		Name: StateRunning,
		OnEnter: func(ctx interface{}) {
			log.Println("enter RUNNING state")
		},
		OnLeave: func(ctx interface{}) {
			log.Println("leave RUNNING state")
		},
	}

	WaitFailoverBeginState = &fsm.State{
		Name: StateWaitFailoverBegin,
		OnEnter: func(ctx interface{}) {
			log.Println("enter WAIT_FAILOVE_BEGIN state")
		},
		OnLeave: func(ctx interface{}) {
			log.Println("leave WAIT_FAILOVER_BEGIN state")
		},
	}

	WaitFailoverEndState = &fsm.State{
		Name: StateWaitFailoverEnd,
		OnEnter: func(ctx interface{}) {
			log.Println("enter WAIT_FAILOVE_END state")
		},
		OnLeave: func(ctx interface{}) {
			log.Println("leave WAIT_FAILOVER_END state")
		},
	}

	OfflineState = &fsm.State{
		Name: StateOffline,
		OnEnter: func(ctx interface{}) {
			log.Println("enter OFFLINE state")
		},
		OnLeave: func(ctx interface{}) {
			log.Println("leave OFFLINE state")
		},
	}

	RedisNodeStateModel = fsm.NewStateModel()
)

func init() {
	RedisNodeStateModel.AddState(RunningState)
	RedisNodeStateModel.AddState(WaitFailoverBeginState)
	RedisNodeStateModel.AddState(WaitFailoverEndState)
	RedisNodeStateModel.AddState(OfflineState)

	/// State: (WaitFailoverRunning)

	RedisNodeStateModel.AddTransition(&fsm.Transition{
		From:       StateRunning,
		To:         StateOffline,
		Input:      Input{F, F, ANY, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply:      nil,
	})

	RedisNodeStateModel.AddTransition(&fsm.Transition{
		From:       StateRunning,
		To:         StateWaitFailoverBegin,
		Input:      Input{T, ANY, FAIL, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply:      nil,
	})

	RedisNodeStateModel.AddTransition(&fsm.Transition{
		From:       StateRunning,
		To:         StateWaitFailoverBegin,
		Input:      Input{ANY, T, FAIL, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply:      nil,
	})

	RedisNodeStateModel.AddTransition(&fsm.Transition{
		From:       StateRunning,
		To:         StateWaitFailoverEnd,
		Input:      Input{T, ANY, FAIL, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply: func(ctx interface{}) {
			log.Println("apply")
		},
	})

	RedisNodeStateModel.AddTransition(&fsm.Transition{
		From:       StateRunning,
		To:         StateWaitFailoverEnd,
		Input:      Input{ANY, T, FAIL, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply: func(ctx interface{}) {
			log.Println("apply")
		},
	})

	/// State: (WaitFailoverBegin)

	RedisNodeStateModel.AddTransition(&fsm.Transition{
		From:       StateWaitFailoverBegin,
		To:         StateRunning,
		Input:      Input{ANY, ANY, FINE, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply:      nil,
	})

	RedisNodeStateModel.AddTransition(&fsm.Transition{
		From:       StateWaitFailoverBegin,
		To:         StateWaitFailoverEnd,
		Input:      Input{ANY, ANY, FAIL, M, CMD_FAILOVER_BEGIN_SIGNAL},
		Priority:   0,
		Constraint: nil,
		Apply: func(ctx interface{}) {
			log.Println("failover master")
		},
	})

	RedisNodeStateModel.AddTransition(&fsm.Transition{
		From:       StateWaitFailoverBegin,
		To:         StateWaitFailoverEnd,
		Input:      Input{ANY, ANY, FAIL, S, CMD_FAILOVER_BEGIN_SIGNAL},
		Priority:   0,
		Constraint: nil,
		Apply: func(ctx interface{}) {
			log.Println("failover slave")
		},
	})

	RedisNodeStateModel.AddTransition(&fsm.Transition{
		From:       StateWaitFailoverBegin,
		To:         StateOffline,
		Input:      Input{ANY, ANY, FAIL, S, ANY},
		Priority:   1,
		Constraint: nil,
		Apply: func(ictx interface{}) {
			ctx := ictx.(StateContext)
			cs := ctx.ClusterState
			node := ctx.NodeState

			for _, s := range cs.AllNodes() {
				resp, err := redis.DisableRead(s.Addr(), node.Id())
				if err == nil {
					log.Println("disable slave apply", resp, node.Id())
					break
				}
			}
		},
	})

	/// State: (WaitFailoverEnd)

	RedisNodeStateModel.AddTransition(&fsm.Transition{
		From:       StateWaitFailoverEnd,
		To:         StateOffline,
		Input:      Input{ANY, ANY, ANY, ANY, CMD_FAILOVER_END_SIGNAL},
		Priority:   1,
		Constraint: nil,
		Apply:      nil,
	})

	/// State: (Offline)

	RedisNodeStateModel.AddTransition(&fsm.Transition{
		From:       StateOffline,
		To:         StateRunning,
		Input:      Input{T, ANY, ANY, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply:      nil,
	})

	RedisNodeStateModel.AddTransition(&fsm.Transition{
		From:       StateOffline,
		To:         StateRunning,
		Input:      Input{ANY, T, ANY, ANY, ANY},
		Priority:   0,
		Constraint: nil,
		Apply:      nil,
	})

	RedisNodeStateModel.AddTransition(&fsm.Transition{
		From:       StateOffline,
		To:         StateWaitFailoverBegin,
		Input:      Input{F, F, FAIL, M, ANY},
		Priority:   0,
		Constraint: nil,
		Apply:      nil,
	})
}

type StateContext struct {
	Input        Input
	ClusterState *ClusterState
	NodeState    *NodeState
}
