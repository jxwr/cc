package command

import (
	"io"
	"strings"

	"github.com/codegangsta/cli"
	"golang.org/x/net/websocket"

	"github.com/ksarch-saas/cc/cli/context"
	"github.com/ksarch-saas/cc/streams"
)

var LogCommand = cli.Command{
	Name:   "log",
	Usage:  "log",
	Action: logAction,
}

func LevelGE(level, base string) bool {
	const levels = "VERBOSE,INFO,EVENT,WARNING,ERROR,FATAL"
	n := strings.Index(levels, level)
	m := strings.Index(levels, strings.ToUpper(base))
	return n >= m
}

func logAction(c *cli.Context) {
	addr := context.GetLeaderWebSocketAddr()
	url := "ws://" + addr + "/log"

	conn, err := websocket.Dial(url, "", url)
	if err != nil {
		Put(err)
		return
	}

	args := c.Args()
	level := "VERBOSE"
	if len(args) > 0 {
		level = args[0]
	}

	var msg streams.LogStreamData
	for {
		err := websocket.JSON.Receive(conn, &msg)
		if err != nil {
			if err == io.EOF {
				break
			}
			Put("Couldn't receive msg " + err.Error())
			break
		}
		if LevelGE(msg.Level, level) {
			Putf("%s %s: [%s] - %s\n", msg.Level, msg.Time.Format("2006/01/02 15:04:05"), msg.Target, msg.Message)
		}
	}
}
