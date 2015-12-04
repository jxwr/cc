package command

import (
	"fmt"

	cc "github.com/ksarch-saas/cc/controller"
	"github.com/ksarch-saas/cc/log"
	"github.com/ksarch-saas/cc/redis"
)

type ReplicateCommand struct {
	ChildId  string
	ParentId string
}

func (self *ReplicateCommand) Execute(c *cc.Controller) (cc.Result, error) {
	cs := c.ClusterState
	child := cs.FindNode(self.ChildId)
	parent := cs.FindNode(self.ParentId)
	if child == nil {
		return nil, fmt.Errorf("Child node not exist %s", self.ChildId)
	}
	if parent == nil {
		return nil, fmt.Errorf("Parent node not exist %s", self.ParentId)
	}
	if parent.Fail || child.Fail {
		return nil, ErrNodeIsDead
	}
	// TODO: more check
	_, err := redis.ClusterReplicate(child.Addr(), parent.Id)
	if err != nil {
		return nil, err
	}
	log.Eventf(child.Addr(), "Reparent to %s(%s).", parent.Addr(), parent.Id)
	return nil, nil
}
