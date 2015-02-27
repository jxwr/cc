package main

import (
	"github.com/jxwr/cc/topo"
)

const (
	EventBuild = 1 + iota
	EventFailover
	EventSlaveDead
	EventSlaveAlive
	EventMasterDead
	EventMasterAlive
	EventAutoRebalance
)

type EventType int

type StateEvent interface {
	Type() EventType
}

type BuildEvent struct {
	Typ EventType
}

func NewBuildEvent() *BuildEvent {
	return &BuildEvent{EventBuild}
}

type FailoverEvent struct {
	Typ EventType
}

func NewFailoverEvent() *FailoverEvent {
	return &FailoverEvent{EventFailover}
}

type AutoRebalanceEvent struct {
	Typ EventType
}

func NewAutoRebalanceEvent() *AutoRebalanceEvent {
	return &AutoRebalanceEvent{EventAutoRebalance}
}

type SlaveDeadEvent struct {
	Typ    EventType
	Server *topo.Server
}

func NewSlaveDeadEvent(server *topo.Server) *SlaveDeadEvent {
	return &SlaveDeadEvent{EventSlaveDead, server}
}

type SlaveAliveEvent struct {
	Typ    EventType
	Server *topo.Server
}

func NewSlaveAliveEvent(server *topo.Server) *SlaveAliveEvent {
	return &SlaveAliveEvent{EventSlaveAlive, server}
}

type MasterDeadEvent struct {
	Typ    EventType
	Server *topo.Server
}

func NewMasterDeadEvent(server *topo.Server) *MasterDeadEvent {
	return &MasterDeadEvent{EventMasterDead, server}
}

type MasterAliveEvent struct {
	Typ    EventType
	Server *topo.Server
}

func NewMasterAliveEvent(server *topo.Server) *MasterAliveEvent {
	return &MasterAliveEvent{EventMasterAlive, server}
}

func (e *BuildEvent) Type() EventType         { return e.Typ }
func (e *FailoverEvent) Type() EventType      { return e.Typ }
func (e *SlaveDeadEvent) Type() EventType     { return e.Typ }
func (e *SlaveAliveEvent) Type() EventType    { return e.Typ }
func (e *MasterDeadEvent) Type() EventType    { return e.Typ }
func (e *MasterAliveEvent) Type() EventType   { return e.Typ }
func (e *AutoRebalanceEvent) Type() EventType { return e.Typ }
