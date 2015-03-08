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
	node := cs.FindNode(self.NodeId)
	if node == nil {
		return nil, ErrNodeNotExist
	}
	_, err := redis.DisableRead(node.Addr(), node.Id)
	return nil, err
}
