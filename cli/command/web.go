package command

import (
	"github.com/codegangsta/cli"

	"github.com/jxwr/cc/cli/context"
)

var WebCommand = cli.Command{
	Name:   "web",
	Usage:  "web, show web console url",
	Action: webAction,
}

func webAction(c *cli.Context) {
	addr := context.GetLeaderAddr()
	Put("http://" + addr + "/ui/cluster.html")
}
