package command

import (
	"fmt"
	"os"
	"os/user"

	"github.com/codegangsta/cli"
	"github.com/ksarch-saas/cc/cli/context"
)

var ConfigCommand = cli.Command{
	Name:   "config",
	Usage:  "config",
	Action: ConfigAction,
	Flags: []cli.Flag{
		cli.StringFlag{"k,key", "", "key"},
		cli.StringFlag{"v,value", "", "value"},
	},
	Description: `
    config the cli tool
    avaliable keys:
        zkhosts:ip1:<port1,ip2:port2>
        historyfile:<dir>
        display:<simple|full>
    `,
}

func ConfigAction(c *cli.Context) {
	key := c.String("k")
	value := c.String("v")
	if key != "zkhosts" && key != "historyfile" && key != "display" {
		fmt.Printf("key %s not exists\n", key)
		os.Exit(-1)
	}
	u, err := user.Current()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	conf, err := context.LoadConfig(u.HomeDir + context.DEFAULT_CONFIG_FILE)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	switch key {
	case "zkhosts":
		conf.Zkhosts = value
		break
	case "historyfile":
		conf.HistoryFile = value
		break
	case "display":
		conf.Display = value
		break
	default:
		break
	}
	err = context.SaveConfig(u.HomeDir+context.DEFAULT_CONFIG_FILE, conf)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	fmt.Println("Update config done")

}
