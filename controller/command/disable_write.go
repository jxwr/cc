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
	node := cs.FindNode(self.NodeId)
	if node == nil {
		return nil, ErrNodeNotExist
	}
	_, err := redis.DisableWrite(node.Addr(), node.Id())
	return nil, err
}
