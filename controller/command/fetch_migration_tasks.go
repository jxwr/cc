package command

import (
	cc "github.com/ksarch-saas/cc/controller"
	"github.com/ksarch-saas/cc/migrate"
)

type FetchMigrationTasksCommand struct{}

type FetchMigrationTasksResult struct {
	Plans []*migrate.MigratePlan
}

func (self *FetchMigrationTasksCommand) Execute(c *cc.Controller) (cc.Result, error) {
	mm := c.MigrateManager
	tasks := mm.AllTasks()
	var plans []*migrate.MigratePlan
	for _, t := range tasks {
		plans = append(plans, t.ToPlan())
	}
	result := FetchMigrationTasksResult{
		Plans: plans,
	}
	return result, nil
}
