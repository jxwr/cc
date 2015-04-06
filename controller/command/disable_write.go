package command

import (
	cc "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/redis"
)

type DisableWriteCommand struct {
	NodeId string
}

func (self *DisableWriteCommand) Execute(c *cc.Controller) (cc.Result, error) {
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
		_, err = redis.DisableWrite(ns.Addr(), target.Id)
		if err == nil {
			return nil, nil
		}
	}
	return nil, err
}
