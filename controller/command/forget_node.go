package command

import (
	"fmt"
	"log"
	"strings"

	cc "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/redis"
)

type ForgetNodeCommand struct {
	NodeId string
}

func (self *ForgetNodeCommand) Execute(c *cc.Controller) (cc.Result, error) {
	cs := c.ClusterState
	target := cs.FindNode(self.NodeId)
	if target == nil {
		return nil, ErrNodeNotExist
	}
	if !target.Free == false {
		return nil, ErrNodeIsFree
	}
	var err error
	forgetCount := 0
	allForgetDone := true
	// 1. 所有节点发送Forget
	for _, ns := range cs.AllNodeStates() {
		if ns.Id() == target.Id {
			continue
		}
		_, err = redis.ClusterForget(ns.Addr(), target.Id)
		if err != nil && !strings.HasPrefix(err.Error(), "ERR Unknown node") {
			allForgetDone = false
			log.Printf("Forget node failed, %v", err)
			continue
		}
		forgetCount++
	}
	if !allForgetDone {
		return nil, fmt.Errorf("Not all forget done, only (%d/%d) success",
			forgetCount, len(cs.AllNodeStates())-1)
	}
	// 2. 重置
	_, err = redis.ClusterReset(target.Addr(), false)
	if err != nil {
		return nil, fmt.Errorf("Reset node %s(%s) failed, %v", target.Id, target.Addr(), err)
	}
	return nil, nil
}
