package command

import (
	cc "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/state"
)

type FailoverEndCommand struct {
	NodeId string
	Done   bool
}

func (self *FailoverEndCommand) Execute(c *cc.Controller) (cc.Result, error) {
	cs := c.ClusterState
	node := cs.FindNodeState(self.NodeId)
	if node == nil {
		return nil, ErrNodeNotExist
	}
	err := node.AdvanceFSM(cs, state.CMD_FAILOVER_END_SIGNAL)

	// NB: 触发FAILOVER_BEGIN时将相关的迁移任务暂停了，此处需要将迁移任务继续
	// 该命令由选主任务结束的回调触发，主从切换任务结束有两种可能：
	// 1）同步成功, Done == true
	// 2）任务超时, Done == false
	// 任务超时意味着主从同步遇到了问题或其他罕见情况，此时不恢复迁移任务，留给管理员手动恢复
	if node.CurrentState() != state.StateWaitFailoverEnd && self.Done {
		mig := c.MigrateManager
		tasks := mig.FindTasksByNode(self.NodeId)
		for _, task := range tasks {
			// FIXME: 处理失败
			task.Resume()
		}
	}
	return nil, err
}
