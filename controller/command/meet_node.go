package command

import (
	cc "github.com/ksarch-saas/cc/controller"
	"github.com/ksarch-saas/cc/log"
	"github.com/ksarch-saas/cc/redis"
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
	if target.Fail {
		return nil, ErrNodeIsDead
	}
	if target.Free == false {
		return nil, ErrNodeNotFree
	}
	var err error
	for _, ns := range cs.AllNodeStates() {
		_, err = redis.ClusterMeet(ns.Addr(), target.Ip, target.Port)
		if err == nil {
			log.Eventf(target.Addr(), "Meet.")
			return nil, nil
		}
	}
	return nil, err
}
