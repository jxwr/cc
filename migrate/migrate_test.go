package migrate

import (
	"fmt"
	"testing"

	"github.com/ksarch-saas/cc/topo"
)

func TestCreate(t *testing.T) {
	/*
		m := NewMigrateManager()

		fromId := "5f674075196119c0d94037583b8a4a9a0e902dd5"
		toId := "8e05f3ec5ab3b21da8337bb6519124847a93fc3f"

		fromNode := topo.NewNode("127.0.0.1", 7000).SetId(fromId)
		toNode := topo.NewNode("127.0.0.1", 7002).SetId(toId)

		task, err := m.CreateTask(fromNode, toNode, []topo.Range{topo.Range{60, 80}})
		if err != nil {
			fmt.Println(err)
		}
		task.Run()
		m.RemoveTask(task)
	*/
}

func TestCutTailRebalancer__0(t *testing.T) {
	ss := []*topo.Node{
		&topo.Node{Id: "s0", Ranges: []topo.Range{
			topo.Range{0, 100},
			topo.Range{200, 300},
			topo.Range{400, 500},
		}},
	}
	ts := []*topo.Node{
		&topo.Node{Id: "t0", Ranges: []topo.Range{}},
		&topo.Node{Id: "t1", Ranges: []topo.Range{}},
		&topo.Node{Id: "t2", Ranges: []topo.Range{}},
	}
	plans := CutTailRebalancer(ss, ts)
	fmt.Println(plans)
}

func TestCutTailRebalancer__1(t *testing.T) {
	ss := []*topo.Node{
		&topo.Node{Id: "s0", Ranges: []topo.Range{
			topo.Range{0, 100},
			topo.Range{200, 300},
		}},
		&topo.Node{Id: "s1", Ranges: []topo.Range{
			topo.Range{400, 500},
		}},
	}
	ts := []*topo.Node{
		&topo.Node{Id: "t0", Ranges: []topo.Range{}},
		&topo.Node{Id: "t1", Ranges: []topo.Range{}},
		&topo.Node{Id: "t2", Ranges: []topo.Range{}},
	}
	plans := CutTailRebalancer(ss, ts)
	fmt.Println(plans)
}

func TestCutTailRebalancer__2(t *testing.T) {
	ss := []*topo.Node{
		&topo.Node{Id: "s0", Ranges: []topo.Range{
			topo.Range{200, 300},
		}},
		&topo.Node{Id: "s1", Ranges: []topo.Range{
			topo.Range{400, 500},
		}},
		&topo.Node{Id: "s2", Ranges: []topo.Range{
			topo.Range{600, 700},
		}},
		&topo.Node{Id: "s3", Ranges: []topo.Range{
			topo.Range{800, 900},
		}},
		&topo.Node{Id: "s4", Ranges: []topo.Range{
			topo.Range{1000, 1100},
		}},
	}
	ts := []*topo.Node{
		&topo.Node{Id: "t0", Ranges: []topo.Range{}},
		&topo.Node{Id: "t1", Ranges: []topo.Range{}},
		&topo.Node{Id: "t2", Ranges: []topo.Range{}},
	}
	plans := CutTailRebalancer(ss, ts)
	fmt.Println(plans)
}
