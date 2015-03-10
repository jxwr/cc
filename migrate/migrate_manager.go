package migrate

import (
	"errors"

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

func (m *MigrateManager) FindTaskBySource(nodeId string) *MigrateTask {
	for _, t := range m.tasks {
		if t.From.Id == nodeId {
			return t
		}
	}
	return nil
}

func (m *MigrateManager) FindTasksByTarget(nodeId string) []*MigrateTask {
	ts := []*MigrateTask{}

	for _, t := range m.tasks {
		if t.To.Id == nodeId {
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
	t := m.FindTaskBySource(task.From.Id)
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

func (m *MigrateManager) Create(fromNode, toNode *topo.Node, ranges []Range) error {
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

func (m *MigrateManager) Pause(nodeId string) error {
	task := m.FindTaskBySource(nodeId)
	if task == nil {
		return ErrMigrateNotExist
	}
	err := task.Pause()
	return err
}

func (m *MigrateManager) Resume(nodeId string) error {
	task := m.FindTaskBySource(nodeId)
	if task == nil {
		return ErrMigrateNotExist
	}
	err := task.Resume()
	return err
}

func (m *MigrateManager) Cancel(nodeId string) error {
	task := m.FindTaskBySource(nodeId)
	if task == nil {
		return ErrMigrateNotExist
	}
	err := task.Cancel()
	if err != nil {
		return err
	}
	m.removeTask(task)
	return nil
}

func (m *MigrateManager) Reset(nodeId string) error {
	task := m.FindTaskBySource(nodeId)
	if task == nil {
		return ErrMigrateNotExist
	}
	err := task.Reset()
	return err
}

func (m *MigrateManager) CancelAll(nodeId string) error {
	tasks := m.AllTasks()
	for _, task := range tasks {
		err := task.Cancel()
		if err != nil {
			return err
		}
		m.removeTask(task)
	}
	return nil
}
