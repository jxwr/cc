package command

import (
	"fmt"

	cc "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/redis"
)

type ReplicateCommand struct {
	ChildId  string
	ParentId string
}

func (self *ReplicateCommand) Execute(c *cc.Controller) (cc.Result, error) {
	cs := c.ClusterState
	child := cs.FindNode(self.ChildId)
	parent := cs.FindNode(self.ParentId)
	if child == nil || parent == nil {
		return nil, ErrNodeNotExist
	}
	if parent.Fail || child.Fail {
		return nil, ErrNodeIsDead
	}
	if child.IsMaster() {
		return nil, fmt.Errorf("child node cannot reparent")
	}
	resp, err := redis.ClusterReplicate(child.Addr(), parent.Id)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
