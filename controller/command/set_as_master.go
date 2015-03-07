package command

import (
	cc "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/redis"
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
	_, err := redis.ClusterFailover(node.Addr())
	return nil, err
}
