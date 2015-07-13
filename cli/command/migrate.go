package command

import (
	"fmt"
	"time"

	"github.com/codegangsta/cli"

	"github.com/ksarch-saas/cc/cli/context"
	"github.com/ksarch-saas/cc/frontend/api"
	"github.com/ksarch-saas/cc/utils"
)

var MigrateCommand = cli.Command{
	Name:   "migrate",
	Usage:  "migrate <sid> <tid> range",
	Action: migrateAction,
}

func migrateAction(c *cli.Context) {
	fmt.Println(c.Args())
	if len(c.Args()) < 3 {
		fmt.Println(ErrInvalidParameter)
		return
	}
	addr := context.GetLeaderAddr()

	url := "http://" + addr + api.MigrateCreatePath
	snodeid, err := context.GetId(c.Args()[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	tnodeid, err := context.GetId(c.Args()[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	ranges := c.Args()[2:]

	req := api.MigrateParams{
		SourceId: snodeid,
		TargetId: tnodeid,
		Ranges:   ranges,
	}
	resp, err := utils.HttpPost(url, req, 5*time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}
	ShowResponse(resp)
}
