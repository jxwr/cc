package command

import (
	"fmt"
	"strings"
	"time"

	"encoding/json"
	"github.com/codegangsta/cli"
	"github.com/ksarch-saas/cc/cli/context"
	"github.com/ksarch-saas/cc/meta"
)

var AppModCommand = cli.Command{
	Name:   "appmod",
	Usage:  "appmod",
	Action: appModAction,
	Flags: []cli.Flag{
		cli.StringFlag{"s,enableslaveread", "", "AutoEnableSlaveRead <true> or <false>"},
		cli.StringFlag{"m,enablemasterwrite", "", "AutoEnableMasterWrite <true> or <false>"},
		cli.StringFlag{"f,failover", "", "AutoFailover <true> or <false>"},
		cli.IntFlag{"i,interval", -1, "AutoFailoverInterval"},
		cli.StringFlag{"r,masterregion", "", "MasterRegion"},
		cli.StringFlag{"R,regions", "", "Regions"},
		cli.IntFlag{"k,migratekey", -1, "MigrateKeysEachTime"},
		cli.IntFlag{"t,migratetimeout", -1, "MigrateTimeout"},
	},
	Description: `
    update app configuraton in zookeeper
    `,
}

func appModAction(c *cli.Context) {
	appname := context.GetAppName()
	s := c.String("s")
	m := c.String("m")
	f := c.String("f")
	i := c.Int("i")
	r := c.String("r")
	R := c.String("R")
	k := c.Int("k")
	t := c.Int("t")

	appConfig := meta.AppConfig{}
	config, version, err := context.GetApp(appname)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = json.Unmarshal(config, &appConfig)
	if err != nil {
		fmt.Println(err)
		return
	}
	//update config if needed
	if s != "" {
		if s == "true" {
			appConfig.AutoEnableSlaveRead = true
		} else if s == "false" {
			appConfig.AutoEnableSlaveRead = false
		}
	}
	if m != "" {
		if m == "true" {
			appConfig.AutoEnableMasterWrite = true
		} else if m == "false" {
			appConfig.AutoEnableMasterWrite = false
		}
	}
	if f != "" {
		if f == "true" {
			appConfig.AutoFailover = true
		} else if f == "false" {
			appConfig.AutoFailover = false
		}
	}
	if i != -1 {
		appConfig.AutoFailoverInterval = time.Duration(i)
	}
	if r != "" {
		appConfig.MasterRegion = r
	}
	if R != "" {
		appConfig.Regions = strings.Split(R, ",")
	}
	if k != -1 {
		appConfig.MigrateKeysEachTime = k
	}
	if t != -1 {
		appConfig.MigrateTimeout = t
	}

	out, err := json.Marshal(appConfig)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = context.ModApp(appname, out, version)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Mod %s success\n%s\n", appname, string(out))
}
