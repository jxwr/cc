package fsm

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyStateModel     = errors.New("fsm: empty state model")
	ErrEmptyStateTrasition = errors.New("fsm: empty state transition")
	ErrStateNotExist       = errors.New("fsm: state not exist")
)

type State struct {
	Name    string
	OnEnter func(ctx interface{})
	OnLeave func(ctx interface{})
}

type Transition struct {
	From       string
	To         string
	Input      Input
	Priority   int
	Constraint func(ctx interface{}) bool
	Apply      func(ctx interface{})
}

type Input interface {
	Eq(Input) bool
}

/// StateModel

type StateModel struct {
	States     map[string]*State
	TransTable map[string][]*Transition
}

func NewStateModel() *StateModel {
	m := &StateModel{
		States:     map[string]*State{},
		TransTable: map[string][]*Transition{},
	}
	return m
}

func (m *StateModel) AddState(s *State) {
	m.States[s.Name] = s
}

func (m *StateModel) AddTransition(t *Transition) {
	_, ok := m.States[t.From]
	if !ok {
		panic("no such state")
	}

	ts, ok := m.TransTable[t.From]
	if !ok {
		m.TransTable[t.From] = []*Transition{t}
		return
	}

	// sort by priority, e.g. 1 > 0
	pos := len(ts)
	for i, u := range ts {
		if t.Priority > u.Priority {
			pos = i
			break
		}
	}
	// insert
	m.TransTable[t.From] = append(ts[:pos], append([]*Transition{t}, ts[pos:]...)...)
}

func (m *StateModel) DumpTransitions() {
	for state, transArray := range m.TransTable {
		fmt.Printf("%s(%d):\n", state, len(transArray))
		for _, t := range transArray {
			fmt.Printf("  |____%v____> %s\n", t.Input, t.To)
		}
	}
}

/// StateMachine

type StateMachine struct {
	model   *StateModel
	current string
}

func NewStateMachine(initalState string, model *StateModel) *StateMachine {
	m := &StateMachine{
		model:   model,
		current: initalState,
	}

	return m
}

func (m *StateMachine) CurrentState() string {
	return m.current
}

func (m *StateMachine) Advance(ctx interface{}, input Input) (string, error) {
	model := m.model
	if model == nil {
		return "", ErrEmptyStateModel
	}

	curr := m.CurrentState()
	_, ok := m.model.States[curr]
	if !ok {
		return "", ErrStateNotExist
	}

	ts, ok := m.model.TransTable[curr]
	if !ok {
		return "", ErrEmptyStateTrasition
	}

	// 按状态转换函数的优先级，顺序检查是否可进行状态变换
	for _, t := range ts {
		if t.Input.Eq(input) && (t.Constraint == nil || t.Constraint(ctx)) {
			// 状态转换
			m.model.States[t.From].OnLeave(ctx)
			m.current = t.To
			m.model.States[t.To].OnEnter(ctx)

			if t.Apply != nil {
				t.Apply(ctx)
			}
			return m.CurrentState(), nil
		}
	}

	return m.CurrentState(), nil
}
