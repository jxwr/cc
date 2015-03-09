package streams

import (
	"github.com/jxwr/cc/topo"
)

type NodeStateStreamData struct {
	*topo.Node
	State   string
	Version int64
}

var (
	NodeStateStream = NewStream("NodeStateStream", 4096)
)

func StartAllStreams() {
	go NodeStateStream.Run()
}
