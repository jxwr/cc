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

func (self *UpdateRegionCommand) Execute(c *cc.Controller) (cc.Result, error) {
	// 更新Cluster拓扑
	cs := c.ClusterState
	cs.UpdateRegionNodes(self.Region, self.Nodes)

	// 首先更新迁移任务状态，以便发现故障时，在处理故障之前就暂停迁移任务
	cluster := cs.GetClusterSnapshot()
	if cluster != nil {
		mm := c.MigrateManager
		mm.HandleNodeStateChange(cluster)
	}

	// 更新Region内Node的状态机
	for _, node := range cs.AllNodeStates() {
		node.AdvanceFSM(cs, state.CMD_NONE)
	}

	return nil, nil
}
