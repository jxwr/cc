package migrate

import (
	"fmt"
	"testing"
	"time"

	"github.com/jxwr/cc/topo"
)

func TestCreate(t *testing.T) {
	m := NewMigrateManager()

	fromId := "5f674075196119c0d94037583b8a4a9a0e902dd5"
	toId := "8e05f3ec5ab3b21da8337bb6519124847a93fc3f"

	fromNode := topo.NewNode("127.0.0.1", 7000).SetId(fromId)
	toNode := topo.NewNode("127.0.0.1", 7002).SetId(toId)

	m.Create(fromNode, toNode, []Range{Range{0, 20}})

	go m.RunTask(fromId)

	fmt.Println("=======")
	time.Sleep(3 * time.Second)
	fmt.Println("pause:", m.Pause(fromId))
	time.Sleep(3 * time.Second)
	fmt.Println("resume:", m.Resume(fromId))
	fmt.Println("cancel:", m.Cancel(fromId))
	fmt.Println("resume:", m.Resume(fromId))
}
