package command

import (
	"fmt"
	"os"
	"time"

	"encoding/json"
	"github.com/codegangsta/cli"
	"github.com/ksarch-saas/cc/cli/context"
	"github.com/ksarch-saas/cc/meta"
	"strings"
)

var AppAddCommand = cli.Command{
	Name:   "appadd",
	Usage:  "appadd",
	Action: appAddAction,
	Flags: []cli.Flag{
		cli.StringFlag{"n,appname", "", "appname"},
		cli.BoolFlag{"s,autoenableslaveread", "AutoEnableSlaveRead"},
		cli.BoolFlag{"m,autoenablemasterwrite", "AutoEnableMasterWrite"},
		cli.BoolFlag{"f,failover", "AutoFailover"},
		cli.IntFlag{"i,interval", 1000, "AutoFailoverInterval"},
		cli.StringFlag{"r,masterregion", "bj", "MasterRegion"},
		cli.StringFlag{"R,regions", "bj,nj", "Regions"},
		cli.IntFlag{"k,migratekey", 100, "MigrateKeysEachTime"},
		cli.IntFlag{"t,migratetimeout", 2000, "MigrateTimeout"},
	},
	Description: `
    add app configuration to zookeeper
    `,
}

func appAddAction(c *cli.Context) {
	appname := c.String("n")
	s := c.Bool("s")
	m := c.Bool("m")
	f := c.Bool("f")
	i := c.Int("i")
	r := c.String("r")
	R := c.String("R")
	k := c.Int("k")
	t := c.Int("t")

	if appname == "" {
		fmt.Println("-n,appname must be assigned")
		os.Exit(-1)
	}
	appConfig := meta.AppConfig{
		AppName:               appname,
		AutoEnableSlaveRead:   s,
		AutoEnableMasterWrite: m,
		AutoFailover:          f,
		AutoFailoverInterval:  time.Duration(i),
		MasterRegion:          r,
		Regions:               strings.Split(R, ","),
		MigrateKeysEachTime:   k,
		MigrateTimeout:        t,
	}
	out, err := json.Marshal(appConfig)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	err = context.AddApp(appname, out)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	fmt.Printf("Add %s success\n%s\n", appname, string(out))
}
