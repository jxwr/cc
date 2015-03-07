package command

import (
	cc "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/state"
	"github.com/jxwr/cc/topo"
)

type UpdateRegionCommand struct {
	Region string
	Nodes  []*topo.Node
}

func (self UpdateRegionCommand) Execute(c *cc.Controller) (cc.Result, error) {
	cs := c.ClusterState

	// 更新ClusterState
	cs.UpdateRegionNodes(self.Region, self.Nodes)

	// 更新Region内Node的状态机
	for _, node := range cs.AllNodeStates() {
		node.AdvanceFSM(cs, state.CMD_NONE)
	}

	return nil, nil
}
