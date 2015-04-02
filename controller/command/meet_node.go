package command

import (
	cc "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/redis"
)

type MeetNodeCommand struct {
	NodeId string
}

func (self *MeetNodeCommand) Execute(c *cc.Controller) (cc.Result, error) {
	cs := c.ClusterState
	target := cs.FindNode(self.NodeId)
	if target == nil {
		return nil, ErrNodeNotExist
	}
	if target.Free == false {
		return nil, ErrNodeNotFree
	}
	var err error
	for _, ns := range cs.AllNodeStates() {
		_, err = redis.ClusterMeet(ns.Addr(), target.Ip, target.Port)
		if err == nil {
			return nil, nil
		}
	}
	return nil, err
}
