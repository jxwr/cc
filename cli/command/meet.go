package command

import (
	"fmt"

	"github.com/codegangsta/cli"
)

var MeetCommand = cli.Command{
	Name:  "meet",
	Usage: "meet a node",
	Action: func(c *cli.Context) {
		fmt.Println("call meet")
	},
}
