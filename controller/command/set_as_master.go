package command

import (
	"fmt"

	cc "github.com/ksarch-saas/cc/controller"
	"github.com/ksarch-saas/cc/redis"
)

type SetAsMasterCommand struct {
	NodeId string
}

func (self *SetAsMasterCommand) Execute(c *cc.Controller) (cc.Result, error) {
	cs := c.ClusterState
	node := cs.FindNode(self.NodeId)
	if node == nil {
		return nil, ErrNodeNotExist
	}
	if node.IsMaster() {
		return nil, ErrNodeIsMaster
	}
	mm := c.MigrateManager
	if len(mm.AllTasks()) > 0 {
		return nil, fmt.Errorf("Migrate task exists, cancel task to continue.")
	}
	_, err := redis.ClusterFailover(node.Addr())
	if err != nil {
		return nil, err
	}
	for _, ns := range cs.AllNodeStates() {
		_, err = redis.EnableWrite(ns.Addr(), self.NodeId)
		if err == nil {
			return nil, nil
		}
	}
	return nil, err
}
