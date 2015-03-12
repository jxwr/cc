package migrate

import (
	"errors"
	"log"

	"github.com/jxwr/cc/topo"
)

var (
	ErrMigrateAlreadyExist = errors.New("mig: task is running on the node")
	ErrMigrateNotExist     = errors.New("mig: no task running on the node")
	ErrReplicatSetNotFound = errors.New("mig: replica set not found")
	ErrNodeNotFound        = errors.New("mig: node not found")
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

	// 只有这两种状态可进行自动切换
	if task.CurrentState() == StateRunning || task.CurrentState() == StateNodeFailure {
		if !fromNode.Fail || !toNode.Fail {
			task.SetState(StateNodeFailure)
		} else {
			task.SetState(StateRunning)
		}
	}
	return nil
}

func (m *MigrateManager) HandleNodeStateChange(cluster *topo.Cluster) {
	for _, task := range m.tasks {
		m.handleTaskChange(task, cluster)
	}
}
