package command

import (
	cc "github.com/ksarch-saas/cc/controller"
	"github.com/ksarch-saas/cc/state"
)

type FailoverEndCommand struct {
	NodeId string
	Done   bool
}

func (self *FailoverEndCommand) Execute(c *cc.Controller) (cc.Result, error) {
	cs := c.ClusterState
	node := cs.FindNodeState(self.NodeId)
	if node == nil {
		return nil, ErrNodeNotExist
	}
	err := node.AdvanceFSM(cs, state.CMD_FAILOVER_END_SIGNAL)
	return nil, err
}

func (self *FailoverEndCommand) clusterCommand() {}
