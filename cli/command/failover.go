package command

import (
	"fmt"
	"sort"
	"time"

	"github.com/codegangsta/cli"

	"github.com/ksarch-saas/cc/cli/context"
	"github.com/ksarch-saas/cc/frontend/api"
	"github.com/ksarch-saas/cc/utils"
)

var FailoverCommand = cli.Command{
	Name:   "failover",
	Usage:  "failover <id>",
	Action: failoverAction,
}

var ListFailoverRecordCommand = cli.Command{
	Name:   "listfailover",
	Usage:  "listfailover",
	Action: listfailoverAction,
}

var GetFailoverRecordCommand = cli.Command{
	Name:   "getfailover",
	Usage:  "getfailover",
	Action: failoverrecordAction,
}

type stringSlice []string

func (a stringSlice) Len() int {
	return len(a)
}
func (a stringSlice) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a stringSlice) Less(i, j int) bool {
	idx1 := a[i][len(a[i])-10 : len(a[i])]
	idx2 := a[j][len(a[j])-10 : len(a[j])]
	return string(idx1) < string(idx2)
}

func listfailoverAction(c *cli.Context) {
	if len(c.Args()) != 0 {
		fmt.Println(ErrInvalidParameter)
		return
	}
	rs, err := context.ListFailoverRecord()
	if err != nil {
		fmt.Println(err)
		return
	}
	sort.Sort(stringSlice(rs))
	for i, r := range rs {
		fmt.Printf("%d:  %s\n", i+1, r)
	}
	fmt.Printf("Total: %d failover record(s)\n", len(rs))
}

func failoverrecordAction(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println(ErrInvalidParameter)
		return
	}
	rn := c.Args()[0]
	rs, _, err := context.GetFailoverRecord(rn)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(rs)
}

func failoverAction(c *cli.Context) {
	if len(c.Args()) != 1 {
		fmt.Println(ErrInvalidParameter)
		return
	}
	addr := context.GetLeaderAddr()
	extraHeader := &utils.ExtraHeader{
		User:  context.Config.User,
		Role:  context.Config.Role,
		Token: context.Config.Token,
	}

	url := "http://" + addr + api.NodeSetAsMasterPath
	nodeid, err := context.GetId(c.Args()[0])
	if err != nil {
		fmt.Println(err)
		return
	}

	req := api.FailoverTakeoverParams{
		NodeId: nodeid,
	}
	resp, err := utils.HttpPostExtra(url, req, 5*time.Second, extraHeader)
	if err != nil {
		fmt.Println(err)
		return
	}
	ShowResponse(resp)
}
