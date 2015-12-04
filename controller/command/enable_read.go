package command

import (
	cc "github.com/ksarch-saas/cc/controller"
	"github.com/ksarch-saas/cc/redis"
)

type EnableReadCommand struct {
	NodeId string
}

func (self *EnableReadCommand) Execute(c *cc.Controller) (cc.Result, error) {
	cs := c.ClusterState
	target := cs.FindNode(self.NodeId)
	if target == nil {
		return nil, ErrNodeNotExist
	}
	if target.Fail {
		return nil, ErrNodeIsDead
	}
	var err error
	for _, ns := range cs.AllNodeStates() {
		_, err = redis.EnableRead(ns.Addr(), target.Id)
		if err == nil {
			return nil, nil
		}
	}
	return nil, err
}
