package command

import (
	"time"

	"github.com/codegangsta/cli"
	"github.com/garyburd/redigo/redis"

	"github.com/ksarch-saas/cc/cli/context"
)

var RedisCliCommand = cli.Command{
	Name:   "do",
	Usage:  "do <id> <redis command>, eval redis command in cli, eg. 'do 982146 info memory'",
	Action: redisCliAction,
}

var (
	CONN_TIMEOUT  = 1 * time.Second
	READ_TIMEOUT  = 2 * time.Second
	WRITE_TIMEOUT = 2 * time.Second
)

func redisCliAction(c *cli.Context) {
	if len(c.Args()) <= 1 {
		Put(ErrInvalidParameter)
		return
	}

	addr, err := context.GetNodeAddr(c.Args()[0])
	if err != nil {
		Put(err)
		return
	}

	conn, err := redis.DialTimeout("tcp", addr, CONN_TIMEOUT, READ_TIMEOUT, WRITE_TIMEOUT)
	if err != nil {
		Put(err)
		return
	}

	cmd := c.Args()[1]
	var args []interface{}
	for _, arg := range c.Args()[2:] {
		args = append(args, arg)
	}

	reply, err := conn.Do(cmd, args...)
	if err != nil {
		Put(err)
		return
	}

	switch reply.(type) {
	case []interface{}:
		replys, _ := redis.Strings(reply, nil)
		for _, reply := range replys {
			Put(reply)
		}
	default:
		reply, _ = redis.String(reply, nil)
		Put(reply)
	}
}
