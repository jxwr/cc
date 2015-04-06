package command

import (
	cc "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/redis"
)

type DisableReadCommand struct {
	NodeId string
}

func (self *DisableReadCommand) Execute(c *cc.Controller) (cc.Result, error) {
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
		_, err = redis.DisableRead(ns.Addr(), target.Id)
		if err == nil {
			return nil, nil
		}
	}
	return nil, err
}
