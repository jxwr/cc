package controller

import (
	"errors"
	"sync"
	"time"

	"github.com/jxwr/cc/meta"
	"github.com/jxwr/cc/migrate"
	"github.com/jxwr/cc/state"
)

var (
	ErrProcessCommandTimedout = errors.New("controller: process command timeout")
	ErrNotClusterLeader       = errors.New("controller: not cluster leader")
	ErrNotRegionLeader        = errors.New("controller: not region leader")
)

type Controller struct {
	mutex          sync.Mutex
	ClusterState   *state.ClusterState
	MigrateManager *migrate.MigrateManager
}

func NewController() *Controller {
	c := &Controller{
		MigrateManager: migrate.NewMigrateManager(),
		ClusterState:   state.NewClusterState(),
		mutex:          sync.Mutex{},
	}
	return c
}

func (c *Controller) ProcessCommand(command Command, timeout time.Duration) (result Result, err error) {
	switch command.Type() {
	case REGION_COMMAND:
		if !meta.IsRegionLeader() {
			return nil, ErrNotRegionLeader
		}
	case CLUSTER_COMMAND:
		if !meta.IsClusterLeader() {
			return nil, ErrNotClusterLeader
		}
	}

	// 一次处理一条命令，也即同一时间只能在做一个状态变换
	c.mutex.Lock()
	defer c.mutex.Unlock()

	resultCh := make(chan Result)
	errorCh := make(chan error)

	//c.ClusterState.DebugDump()

	go func() {
		result, err := command.Execute(c)
		if err != nil {
			errorCh <- err
		} else {
			resultCh <- result
		}
	}()

	select {
	case result = <-resultCh:
	case err = <-errorCh:
	case <-time.After(timeout):
		err = ErrProcessCommandTimedout
	}
	return
}
