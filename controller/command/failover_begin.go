package command

import (
	cc "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/state"
)

type FailoverBeginCommand struct {
	NodeId string
}

func (self *FailoverBeginCommand) Execute(c *cc.Controller) (cc.Result, error) {
	cs := c.ClusterState
	node := cs.FindNode(self.NodeId)
	if node == nil {
		return nil, ErrNodeNotExist
	}

	// 避免迁移过程中进行failover造成少量数据丢失。出现可能性很低。
	// 假如故障节点一息尚存，迁移任务还在缓慢进行此时如果进行了Failover，迁移
	// 进程仍然会向原Master(已FAIL)搬迁数据。
	mig := c.MigrateManager
	tasks := mig.FindTasksByNode(self.NodeId)
	for _, task := range tasks {
		task.Pause()
	}
	err := node.AdvanceFSM(cs, state.CMD_FAILOVER_BEGIN_SIGNAL)

	// 如果不处于WaitFailoverEnd状态，说明状态变换失败，恢复迁移过程
	if node.CurrentState() != state.StateWaitFailoverEnd {
		for _, task := range tasks {
			task.Resume()
		}
	}
	return nil, err
}
