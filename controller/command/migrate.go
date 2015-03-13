package command

import (
	cc "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/migrate"
	"github.com/jxwr/cc/topo"
)

type MigrateCommand struct {
	SourceId string
	TargetId string
	Ranges   []topo.Range
}

func (self *MigrateCommand) Execute(c *cc.Controller) (cc.Result, error) {
	cs := c.ClusterState
	cluster := cs.GetClusterSnapshot()
	if cluster != nil {
		mm := c.MigrateManager
		task, err := mm.CreateTask(self.SourceId, self.TargetId, self.Ranges, cluster)
		if err != nil {
			return nil, err
		}
		go func() {
			task.Run()
			mm.RemoveTask(task)
		}()
	} else {
		return nil, ErrClusterSnapshotNotReady
	}
	return nil, nil
}

type MigratePauseCommand struct {
	SourceId string
}

func (self *MigratePauseCommand) Execute(c *cc.Controller) (cc.Result, error) {
	mm := c.MigrateManager
	task := mm.FindTaskBySource(self.SourceId)
	if task == nil {
		return nil, ErrMigrateTaskNotExist
	}
	task.SetState(migrate.StatePausing)
	return nil, nil
}

type MigrateResumeCommand struct {
	SourceId string
}

func (self *MigrateResumeCommand) Execute(c *cc.Controller) (cc.Result, error) {
	mm := c.MigrateManager
	task := mm.FindTaskBySource(self.SourceId)
	if task == nil {
		return nil, ErrMigrateTaskNotExist
	}
	task.SetState(migrate.StateRunning)
	return nil, nil
}

type MigrateCancelCommand struct {
	SourceId string
}

func (self *MigrateCancelCommand) Execute(c *cc.Controller) (cc.Result, error) {
	mm := c.MigrateManager
	task := mm.FindTaskBySource(self.SourceId)
	if task == nil {
		return nil, ErrMigrateTaskNotExist
	}
	task.SetState(migrate.StateCancelling)
	return nil, nil
}
