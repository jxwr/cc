package command

import (
	cc "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/redis"
)

type EnableReadCommand struct {
	NodeId string
}

func (self *EnableReadCommand) Execute(c *cc.Controller) (cc.Result, error) {
	cs := c.ClusterState
	node := cs.FindNode(self.NodeId)
	if node == nil {
		return nil, ErrNodeNotExist
	}
	_, err := redis.EnableRead(node.Addr(), node.Id())
	return nil, err
}
