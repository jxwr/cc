package streams

import (
	"github.com/jxwr/cc/topo"
)

type NodeStateStreamData struct {
	*topo.Node
	State   string
	Version int64
}

type MigrateStateStreamData struct {
	SourceId       string
	TargetId       string
	State          string
	Ranges         []topo.Range
	CurrRangeIndex int
	CurrSlot       int
}

var (
	NodeStateStream      = NewStream("NodeStateStream", 4096)
	MigrateStateStream   = NewStream("MigrateStateStream", 4096)
	RebalanceStateStream = NewStream("RebalanceStateStream", 4096)
)

func StartAllStreams() {
	go NodeStateStream.Run()
	go MigrateStateStream.Run()
	go RebalanceStateStream.Run()
}
