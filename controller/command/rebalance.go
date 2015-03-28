package command

import (
	cc "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/migrate"
)

type RebalanceCommand struct {
	Method       string
	TargetIds    []string
	ShowPlanOnly bool
}

// Rebalance任务同时只能有一个
func (self *RebalanceCommand) Execute(c *cc.Controller) (cc.Result, error) {
	mm := c.MigrateManager
	cs := c.ClusterState
	cluster := cs.GetClusterSnapshot()

	if self.Method == "" {
		self.Method = "default"
	}

	plans, err := migrate.GenerateRebalancePlan(self.Method, cluster, self.TargetIds)
	if err != nil {
		return nil, err
	}

	// 是否立即执行？
	if !self.ShowPlanOnly && len(plans) > 0 {
		err = mm.RunRebalanceTask(plans, cluster)
		if err != nil {
			return nil, err
		}
	}

	return plans, nil
}
