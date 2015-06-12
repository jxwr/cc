package command

import (
	cc "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/log"
	"github.com/jxwr/cc/meta"
	"github.com/jxwr/cc/redis"
	"github.com/jxwr/cc/state"
	"github.com/jxwr/cc/topo"
)

type UpdateRegionCommand struct {
	Region string
	Nodes  []*topo.Node
}

func (self *UpdateRegionCommand) Execute(c *cc.Controller) (cc.Result, error) {
	if len(self.Nodes) == 0 {
		return nil, nil
	}

	// 更新Cluster拓扑
	cs := c.ClusterState
	cs.UpdateRegionNodes(self.Region, self.Nodes)

	// 首先更新迁移任务状态，以便发现故障时，在处理故障之前就暂停迁移任务
	cluster := cs.GetClusterSnapshot()
	if cluster != nil {
		mm := c.MigrateManager
		mm.HandleNodeStateChange(cluster)
	}

	for _, ns := range cs.AllNodeStates() {
		node := ns.Node()
		// Slave auto enable read ?
		if !node.IsMaster() && !node.Fail && !node.Readable && node.MasterLinkStatus == "up" {
			if meta.GetAppConfig().AutoEnableSlaveRead {
				redis.EnableRead(node.Addr(), node.Id)
			}
		}
		// Master auto enable write ?
		if node.IsMaster() && !node.Fail && !node.Writable {
			if meta.GetAppConfig().AutoEnableMasterWrite {
				redis.EnableWrite(node.Addr(), node.Id)
			}
		}
		// Fix chained replication: slave's parent is slave.
		if !node.IsMaster() {
			parent := cs.FindNode(node.ParentId)
			// Parent is not master?
			if parent != nil && !parent.IsMaster() {
				grandpa := cs.FindNode(parent.ParentId)
				if grandpa != nil {
					_, err := redis.ClusterReplicate(node.Addr(), grandpa.Addr())
					if err == nil {
						log.Warningf(node.Addr(), "Fix chained replication, (%s->%s->%s)=>(%s->%s)",
							node, parent, grandpa, node, grandpa)
					}
				} else {
					log.Warningf(node.Addr(), "Found chained replication, (%s->%s->nil), cannot fix.",
						node, parent)
				}
			}
		}
		// 更新Region内Node的状态机
		ns.AdvanceFSM(cs, state.CMD_NONE)
	}

	return nil, nil
}
