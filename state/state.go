package state

import (
	"fmt"
)

type MsgField int

const (
	T    MsgField = iota + 1 // True
	F                        // False
	FAIL                     // Fail?
	FINE                     // Not Fail
	S                        // Slave
	M                        // Master
	/// CMD
	CMD_NONE
	CMD_FAILOVER
	CMD_FAILOVER_EXIT
	ANY // *
)

var CmdNames = map[MsgField]string{
	CMD_NONE:          "NONE",
	CMD_FAILOVER:      "FL",
	CMD_FAILOVER_EXIT: "FE",
}

func (f MsgField) String() string {
	switch f {
	case T:
		return "T"
	case F:
		return "F"
	case FAIL:
		return "FAIL"
	case FINE:
		return "FINE"
	case S:
		return "S"
	case M:
		return "M"
	case ANY:
		return "*"
	default:
		return CmdNames[f]
	}
}

type Msg struct {
	Read    MsgField
	Write   MsgField
	Fail    MsgField
	Role    MsgField
	Command MsgField
}

func (s Msg) String() string {
	return fmt.Sprintf("(%v,%v,%v,%v,%v)", s.Read, s.Write, s.Fail, s.Role, s.Command)
}

func (s Msg) Eq(t Msg) bool {
	if s.Read != ANY && t.Read != ANY && s.Read != t.Read {
		return false
	}

	if s.Write != ANY && t.Write != ANY && s.Write != t.Write {
		return false
	}

	if s.Fail != ANY && t.Fail != ANY && s.Fail != t.Fail {
		return false
	}

	if s.Role != ANY && t.Role != ANY && s.Role != t.Role {
		return false
	}

	if s.Command != ANY && t.Command != ANY && s.Command != t.Command {
		return false
	}

	return true
}

const (
	StateRunning           = "RUNNING"
	StateWaitFailoverBegin = "WAIT_FAILOVER_BEGIN"
	StateWaitFailoverEnd   = "WAIT_FAILOVER_END"
	StateOffline           = "OFFLINE"
)

type State struct {
	Name    string
	OnEnter func()
	OnLeave func()
}

type Transition struct {
	From       string
	To         string
	Input      Msg
	Priority   int
	Constraint func() bool
	Apply      func()
}
