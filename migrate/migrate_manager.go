package migrate

import (
	"errors"
	"log"

	"github.com/jxwr/cc/redis"
	"github.com/jxwr/cc/topo"
)

var (
	ErrMigrateAlreadyExist = errors.New("mig: task is running on the node")
	ErrMigrateNotExist     = errors.New("mig: no task running on the node")
	ErrReplicatSetNotFound = errors.New("mig: replica set not found")
	ErrNodeNotFound        = errors.New("mig: node not found")
	ErrSourceNodeFail      = errors.New("mig: source node failure")
	ErrTargetNodeFail      = errors.New("mig: target node failure")
	ErrCanNotRecover       = errors.New("mig: can not recover")
)

/// Migrate

type MigrateManager struct {
	tasks []*MigrateTask
}

func NewMigrateManager() *MigrateManager {
	m := &MigrateManager{
		tasks: []*MigrateTask{},
	}
	return m
}

func (m *MigrateManager) CreateTask(sourceId, targetId string, ranges []topo.Range, cluster *topo.Cluster) (*MigrateTask, error) {
	task := m.FindTaskBySource(sourceId)
	if task != nil {
		return nil, ErrMigrateAlreadyExist
	}
	sourceRS := cluster.FindReplicaSetByNode(sourceId)
	targetRS := cluster.FindReplicaSetByNode(targetId)
	if sourceRS == nil || targetRS == nil {
		return nil, ErrReplicatSetNotFound
	}
	task = NewMigrateTask(sourceRS, targetRS, ranges)
	m.AddTask(task)
	return task, nil
}

func (m *MigrateManager) AddTask(task *MigrateTask) error {
	t := m.FindTaskBySource(task.SourceNode().Id)
	if t != nil {
		return ErrMigrateAlreadyExist
	}
	m.tasks = append(m.tasks, task)
	return nil
}

func (m *MigrateManager) RemoveTask(task *MigrateTask) {
	pos := -1
	for i, t := range m.tasks {
		if t == task {
			pos = i
		}
	}
	if pos != -1 {
		m.tasks = append(m.tasks[:pos], m.tasks[pos+1:]...)
	}
}

func (m *MigrateManager) FindTasksByNode(nodeId string) []*MigrateTask {
	ts := m.FindTasksByTarget(nodeId)
	t := m.FindTaskBySource(nodeId)
	if t != nil {
		ts = append(ts, t)
	}
	return ts
}

func (m *MigrateManager) AllTasks() []*MigrateTask {
	return m.tasks
}

func (m *MigrateManager) FindTasksByTarget(nodeId string) []*MigrateTask {
	ts := []*MigrateTask{}

	for _, t := range m.tasks {
		if t.TargetNode().Id == nodeId {
			ts = append(ts, t)
		}
	}
	return ts
}

func (m *MigrateManager) FindTaskBySource(nodeId string) *MigrateTask {
	for _, t := range m.tasks {
		if t.SourceNode().Id == nodeId {
			return t
		}
	}
	return nil
}

// 更新任务状态机
func (m *MigrateManager) handleTaskChange(task *MigrateTask, cluster *topo.Cluster) error {
	fromNode := cluster.FindNode(task.SourceNode().Id)
	toNode := cluster.FindNode(task.TargetNode().Id)

	if fromNode == nil {
		log.Printf("mig: source node not exist\n")
		return ErrNodeNotFound
	}
	if toNode == nil {
		log.Printf("mig: target node not exist\n")
		return ErrNodeNotFound
	}

	// 角色变化说明该分片进行了主从切换，需要修正Task结构
	if !fromNode.IsMaster() {
		rs := cluster.FindReplicaSetByNode(fromNode.Id)
		if rs == nil {
			log.Printf("mig: %s role changed, but new replica set not found\n", fromNode.Id)
			return ErrReplicatSetNotFound
		}
		task.ReplaceSourceReplicaSet(rs)
	}
	if !toNode.IsMaster() {
		rs := cluster.FindReplicaSetByNode(toNode.Id)
		if rs == nil {
			log.Printf("mig: %s role changed, but new replica set not found\n", toNode.Id)
			return ErrReplicatSetNotFound
		}
		task.ReplaceTargetReplicaSet(rs)
	}

	// 如果是源节点挂了，直接取消，等待主从切换之后重建任务
	if fromNode.Fail {
		log.Printf("mig: cancel migration task %s\n", task.TaskName())
		task.SetState(StateCancelling)
		return ErrSourceNodeFail
	}
	// 如果目标节点挂了，需要记录当前的ReplicaSet，观察等待主从切换
	if toNode.Fail {
		if task.CurrentState() == StateRunning {
			task.SetState(StateTargetNodeFailure)
			task.SetBackupReplicaSet(task.TargetReplicaSet())
			return ErrTargetNodeFail
		}
	} else {
		task.SetState(StateRunning)
		task.SetBackupReplicaSet(nil)
	}
	// 如果目标节点已经进行了Failover(重新选主)，我们需要找到对应的新主
	// 方法是从BackupReplicaSet里取一个从来查找
	if toNode.IsStandbyMaster() {
		brs := task.BackupReplicaSet()
		if brs == nil {
			task.SetState(StateCancelling)
			log.Println("mig: no backup replicaset found, controller maybe restarted after target master failure, can not do recovery.")
			return ErrCanNotRecover
		}
		slaves := brs.Slaves()
		if len(slaves) == 0 {
			task.SetState(StateCancelling)
			log.Println("mig: the dead target master has no slave, cannot do recovery.")
			return ErrCanNotRecover
		} else {
			rs := cluster.FindReplicaSetByNode(slaves[0].Id)
			if rs == nil {
				task.SetState(StateCancelling)
				log.Println("mig: no replicaset for slave of dead target master found")
				return ErrCanNotRecover
			}
			task.ReplaceTargetReplicaSet(rs)
			log.Printf("mig: recover dead target node to %s()\n",
				rs.Master().Id, rs.Master().Addr())
		}
	}
	return nil
}

func (m *MigrateManager) HandleNodeStateChange(cluster *topo.Cluster) {
	// 处理主节点的迁移任务重建
	for _, node := range cluster.AllNodes() {
		if node.IsMaster() && !node.Fail && len(node.Migrating) != 0 {
			// 如果已经存在该节点的迁移任务，先跳过，等结束后再处理
			task := m.FindTaskBySource(node.Id)
			if task != nil {
				continue
			}

			log.Printf("Will recover migrating task for %s\n", node.Id)

			for id, slots := range node.Migrating {
				// 根据slot生成ranges
				ranges := []topo.Range{}
				for _, slot := range slots {
					// 如果是自己
					if id == node.Id {
						redis.SetSlot(node.Addr(), slot, redis.SLOT_STABLE, "")
					} else {
						ranges = append(ranges, topo.Range{Left: slot, Right: slot})
					}
				}

				rs := cluster.FindReplicaSetByNode(id)
				if rs.FindNode(id).IsStandbyMaster() {
					continue
				}

				task, err := m.CreateTask(node.Id, rs.Master().Id, ranges, cluster)
				if err != nil {
					log.Println("Can not recover migrate task,", err)
				} else {
					go func(t *MigrateTask) {
						t.Run()
						m.RemoveTask(t)
					}(task)
				}
			}
		}
	}

	for _, task := range m.tasks {
		m.handleTaskChange(task, cluster)
	}
}
