package command

import (
	"fmt"

	cc "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/migrate"
)

type RebalanceCommand struct {
	Method       string
	TargetIds    []string
	ShowPlanOnly bool
}

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

	if !self.ShowPlanOnly {
		err = mm.RunRebalanceTask(plans, cluster)
		if err != nil {
			return nil, err
		}
	}

	fmt.Println(plans)
	return plans, nil
}
