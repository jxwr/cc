package command

import (
	cc "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/state"
)

type FailoverBeginCommand struct {
	NodeId string
}

func (self *FailoverBeginCommand) Execute(c *cc.Controller) (cc.Result, error) {
	cs := c.ClusterState
	node := cs.FindNode(self.NodeId)
	if node == nil {
		return nil, ErrNodeNotExist
	}
	err := node.AdvanceFSM(cs, state.CMD_FAILOVER_BEGIN_SIGNAL)
	return nil, err
}
