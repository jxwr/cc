package command

import (
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"golang.org/x/net/websocket"

	"github.com/jxwr/cc/cli/context"
	"github.com/jxwr/cc/frontend/api"
	"github.com/jxwr/cc/streams"
	"github.com/jxwr/cc/utils"
)

var LogCommand = cli.Command{
	Name:   "log",
	Usage:  "log [<num>], same as 'tail -n' or 'tail -f'",
	Action: logAction,
}

func LevelGE(level, base string) bool {
	const levels = "VERBOSE,INFO,EVENT,WARNING,ERROR,FATAL"
	n := strings.Index(levels, level)
	m := strings.Index(levels, strings.ToUpper(base))
	return n >= m
}

func logAction(c *cli.Context) {
	// tail -n
	if len(c.Args()) == 1 {
		n, err := strconv.Atoi(c.Args()[0])
		if err != nil {
			Put(err)
			return
		}
		addr := context.GetLeaderAddr()
		url := "http://" + addr + api.LogSlicePath
		req := api.LogSliceParams{
			Pos:   0,
			Count: n,
		}
		resp, err := utils.HttpPost(url, req, 5*time.Second)
		if err != nil {
			Put(err)
			return
		}

		var lines []string
		err = utils.InterfaceToStruct(resp.Body, &lines)
		if err != nil {
			Put(err)
			return
		}
		for _, line := range lines {
			Putf(line)
		}
		return
	}

	// blocking tail
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
