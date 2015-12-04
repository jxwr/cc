package command

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/ksarch-saas/cc/cli/context"
	"github.com/ksarch-saas/cc/controller/command"
	"github.com/ksarch-saas/cc/frontend/api"
	"github.com/ksarch-saas/cc/redis"
	"github.com/ksarch-saas/cc/topo"
	"github.com/ksarch-saas/cc/utils"
	"sort"
	"time"
)

var Slot2NodeCommand = cli.Command{
	Name:   "slot2node",
	Usage:  "slot2node",
	Action: slot2NodeAction,
	Description: `
    set merger slot2node map config, using command:slot2node
    `,
}

func slot2NodeAction(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println(ErrInvalidParameter)
		return
	}
	dest := c.Args()[0]
	addr := context.GetLeaderAddr()
	url := "http://" + addr + api.FetchReplicaSetsPath

	resp, err := utils.HttpGet(url, nil, 5*time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}

	var rss command.FetchReplicaSetsResult
	err = utils.InterfaceToStruct(resp.Body, &rss)
	if err != nil {
		fmt.Println(err)
		return
	}
	sort.Sort(topo.ByMasterId(rss.ReplicaSets))

	for _, rs := range rss.ReplicaSets {
		masteraddr := fmt.Sprintf("%s:%d", rs.Master.Ip, rs.Master.Port)
		for _, r := range rs.Master.Ranges {
			setConfigByRange(masteraddr, r, dest)
		}
	}
}

func setConfigByRange(masteraddr string, r topo.Range, dest string) {
	fmt.Printf("Setting range [%d-%d] map to %s\n", r.Left, r.Right, masteraddr)
	for i := r.Left; i <= r.Right; i++ {
		redis.Slot2Node(dest, i, masteraddr)
	}
}
