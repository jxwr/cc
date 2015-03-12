package command

import (
	cc "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/topo"
)

type MigrateCommand struct {
	SourceId string
	TargetId string
	Ranges   []topo.Range
}

func (self MigrateCommand) Execute(c *cc.Controller) (cc.Result, error) {
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
