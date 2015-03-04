package state

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyStateModel     = errors.New("state_machine: empty state model")
	ErrEmptyStateTrasition = errors.New("state_machine: empty state transition")
	ErrStateNotExist       = errors.New("state_machine: state not exist")
)

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
	// insert at right pos
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

func (m *StateMachine) Next(msg Msg) (string, error) {
	model := m.model
	if model == nil {
		return "", ErrEmptyStateModel
	}

	curr := m.current
	_, ok := m.model.States[curr]
	if !ok {
		return "", ErrStateNotExist
	}

	ts, ok := m.model.TransTable[curr]
	if !ok {
		return "", ErrEmptyStateTrasition
	}

	for _, t := range ts {
		if t.Input.Eq(msg) && (t.Constraint == nil || t.Constraint()) {
			m.model.States[t.From].OnLeave()
			m.current = t.To
			m.model.States[t.To].OnEnter()
			t.Apply()
			return m.CurrentState(), nil
		}
	}

	return "", nil
}
