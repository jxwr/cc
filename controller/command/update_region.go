package command

import (
	ctl "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/state"
	"github.com/jxwr/cc/topo"
)

type UpdateRegionCommand struct {
	Region string
	Nodes  []*topo.Node
}

func (self UpdateRegionCommand) Execute(c *ctl.Controller) (ctl.Result, error) {
	cs := c.ClusterState

	cs.UpdateRegionNodes(self.Region, self.Nodes)

	for _, node := range cs.AllNodes() {
		node.AdvanceFSM(cs, state.CMD_NONE)
	}

	return nil, nil
}
