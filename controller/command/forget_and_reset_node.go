package command

import (
	"fmt"
	"strings"

	cc "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/log"
	"github.com/jxwr/cc/redis"
)

type ForgetAndResetNodeCommand struct {
	NodeId string
}

// 似乎，只有同时进行Forget和Reset才有意义，否则都是一个不一致的状态
func (self *ForgetAndResetNodeCommand) Execute(c *cc.Controller) (cc.Result, error) {
	cs := c.ClusterState
	target := cs.FindNode(self.NodeId)
	if target == nil {
		return nil, ErrNodeNotExist
	}
	if !target.Free == false {
		return nil, ErrNodeIsFree
	}
	if len(target.Ranges) > 0 {
		return nil, ErrNodeNotEmpty
	}
	var err error
	forgetCount := 0
	allForgetDone := true
	// 1. 所有节点发送Forget
	for _, ns := range cs.AllNodeStates() {
		if ns.Id() == target.Id {
			continue
		}
		node := ns.Node()
		_, err = redis.ClusterForget(ns.Addr(), target.Id)
		if !node.Fail && err != nil && !strings.HasPrefix(err.Error(), "ERR Unknown node") {
			allForgetDone = false
			log.Warningf(target.Addr(), "Forget node %s(%s) failed, %v", ns.Addr(), ns.Id(), err)
			continue
		}
		log.Eventf(target.Addr(), "Forget by %s(%s).", ns.Addr(), ns.Id())
		forgetCount++
	}
	if !allForgetDone {
		return nil, fmt.Errorf("Not all forget done, only (%d/%d) success",
			forgetCount, len(cs.AllNodeStates())-1)
	}
	// 2. 重置
	if !target.Fail {
		_, err = redis.ClusterReset(target.Addr(), false)
		if err != nil {
			return nil, fmt.Errorf("Reset node %s(%s) failed, %v", target.Id, target.Addr(), err)
		}
		log.Eventf(target.Addr(), "Reset.")
	}
	return nil, nil
}
