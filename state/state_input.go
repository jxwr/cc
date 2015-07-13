package state

import (
	"fmt"

	"github.com/ksarch-saas/cc/fsm"
)

/// InputField

type InputField int

// 状态转换输入是一个(Read,Write,Fail?,Role,Command)五元组
// 常量名字起成这样是为了定义状态转换时看起来直观一些
// 如：Input{T,ANY,FAIL,M,CMD_NONE}
const (
	T    InputField = iota + 1 // True
	F                          // False
	FAIL                       // Fail?
	FINE                       // Not Fail
	S                          // Slave
	M                          // Master
	ANY                        // *

	CMD_NONE                  // 状态控制命令，空命令
	CMD_FAILOVER_BEGIN_SIGNAL // 开始进行Failover
	CMD_FAILOVER_END_SIGNAL   // Failover结束信号
)

var CmdNames = map[InputField]string{
	CMD_NONE:                  "NONE",
	CMD_FAILOVER_BEGIN_SIGNAL: "FAILOVER_BEGIN_SIGNAL",
	CMD_FAILOVER_END_SIGNAL:   "FAILOVER_END_SIGNAL",
}

func (f InputField) String() string {
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

/// Input

type Input struct {
	Read    InputField
	Write   InputField
	Fail    InputField
	Role    InputField
	Command InputField
}

// 判断Input是否等价，ANY为通配
func (s Input) Eq(it fsm.Input) bool {
	t := it.(Input)
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

func (s Input) String() string {
	str := fmt.Sprintf(
		"(%v,%v,%v,%v,%v)",
		s.Read, s.Write, s.Fail, s.Role, s.Command,
	)
	return str
}
