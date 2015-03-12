package migrate

import (
	"errors"
	"log"

	"github.com/jxwr/cc/topo"
)

var (
	ErrMigrateAlreadyExist = errors.New("mig: task is running on the node")
	ErrMigrateNotExist     = errors.New("mig: no task running on the node")
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

func (m *MigrateManager) handleTaskChange(task *MigrateTask, cluster *topo.Cluster) error {
	fromNode := cluster.FindNode(task.SourceNode().Id)
	toNode := cluster.FindNode(task.TargetNode().Id)

	if fromNode == nil {
		log.Printf("mig: source node not exist\n")
		return nil
	}
	if toNode == nil {
		log.Printf("mig: target node not exist\n")
		return nil
	}

	if !fromNode.IsMaster() {
		rs := cluster.FindReplicaSetByNode(fromNode.Id)
		fromNode = rs.Master()
		if fromNode == nil {
			log.Println("mig: no source master node found")
			return nil
		}
		task.ReplaceSourceNode(fromNode)
	}

	if !toNode.IsMaster() {
		rs := cluster.FindReplicaSetByNode(toNode.Id)
		toNode = rs.Master()
		if toNode == nil {
			log.Println("mig: no target master node found")
			return nil
		}
		task.ReplaceSourceNode(toNode)
	}

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

func (m *MigrateManager) FindTaskBySource(nodeId string) *MigrateTask {
	for _, t := range m.tasks {
		if t.SourceNode().Id == nodeId {
			return t
		}
	}
	return nil
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

func (m *MigrateManager) addTask(task *MigrateTask) error {
	t := m.FindTaskBySource(task.SourceNode().Id)
	if t != nil {
		return ErrMigrateAlreadyExist
	}
	m.tasks = append(m.tasks, task)
	return nil
}

func (m *MigrateManager) removeTask(task *MigrateTask) {
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

func (m *MigrateManager) CreateTask(fromNode, toNode *topo.Node, ranges []topo.Range) error {
	task := m.FindTaskBySource(fromNode.Id)
	if task != nil {
		return ErrMigrateAlreadyExist
	}
	task = NewMigrateTask(fromNode, toNode, ranges)
	m.addTask(task)
	return nil
}

func (m *MigrateManager) RunTask(nodeId string) {
	task := m.FindTaskBySource(nodeId)
	task.Run()
	m.removeTask(task)
}
