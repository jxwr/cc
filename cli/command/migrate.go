package command

import (
	"fmt"
	"time"

	"github.com/codegangsta/cli"

	"github.com/jxwr/cc/cli/context"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/utils"
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
	snodeid := context.GetId(c.Args()[0])
	tnodeid := context.GetId(c.Args()[1])
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
