package redis

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	ErrNotCutter   = errors.New("redis: the server is not a cutter")
	ErrConnFailed  = errors.New("redis: connection error")
	ErrPingFailed  = errors.New("redis: ping error")
	ErrServer      = errors.New("redis: server error")
	ErrInvalidAddr = errors.New("redis: invalid address string")
)

const (
	SLOT_MIGRATING = "MIGRATING"
	SLOT_IMPORTING = "IMPORTING"
	SLOT_STABLE    = "STABLE"
	SLOT_NODE      = "NODE"
)

/// Misc

func IsAlive(addr string) bool {
	conn, err := redis.Dial("tcp", addr)
	if err != nil {
		return false
	}
	defer conn.Close()
	resp, err := redis.String(conn.Do("PING"))
	if err != nil || resp != "PONG" {
		return false
	}
	return true
}

/// Cluster

func SetAsMasterWaitSyncDone(addr string, waitSyncDone bool) error {
	conn, err := redis.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = redis.String(conn.Do("cluster", "failover", "force"))
	if err != nil {
		return err
	}

	if !waitSyncDone {
		return nil
	}

	for {
		info, err := FetchInfo(addr, "replication")
		if err == nil {
			time.Sleep(5 * time.Second)
			n, err := info.GetInt64("connected_slaves")
			if err != nil {
				continue
			}
			done := true
			for i := int64(0); i < n; i++ {
				repl := info.Get(fmt.Sprintf("slave%d", i))
				if !strings.Contains(repl, "online") {
					done = false
				}
			}
			if done {
				return nil
			}
		}
	}
	return nil
}

func ClusterNodes(addr string) (string, error) {
	conn, err := redis.Dial("tcp", addr)
	if err != nil {
		return "", ErrConnFailed
	}
	defer conn.Close()

	resp, err := redis.String(conn.Do("cluster", "nodes", "extra"))
	if err != nil {
		return "", err
	}

	return resp, nil
}

func DisableRead(addr, id string) (string, error) {
	conn, err := redis.Dial("tcp", addr)
	if err != nil {
		return "", ErrConnFailed
	}
	defer conn.Close()

	resp, err := redis.String(conn.Do("cluster", "chmod", "-r", id))
	if err != nil {
		return "", err
	}

	return resp, nil
}

func EnableRead(addr, id string) (string, error) {
	conn, err := redis.Dial("tcp", addr)
	if err != nil {
		return "", ErrConnFailed
	}
	defer conn.Close()

	resp, err := redis.String(conn.Do("cluster", "chmod", "+r", id))
	if err != nil {
		return "", err
	}

	return resp, nil
}

func DisableWrite(addr, id string) (string, error) {
	conn, err := redis.Dial("tcp", addr)
	if err != nil {
		return "", ErrConnFailed
	}
	defer conn.Close()

	resp, err := redis.String(conn.Do("cluster", "chmod", "-w", id))
	if err != nil {
		return "", err
	}

	return resp, nil
}

func EnableWrite(addr, id string) (string, error) {
	conn, err := redis.Dial("tcp", addr)
	if err != nil {
		return "", ErrConnFailed
	}
	defer conn.Close()

	resp, err := redis.String(conn.Do("cluster", "chmod", "+w", id))
	if err != nil {
		return "", err
	}

	return resp, nil
}

func ClusterFailover(addr string) (string, error) {
	conn, err := redis.Dial("tcp", addr)
	if err != nil {
		return "", ErrConnFailed
	}
	defer conn.Close()

	resp, err := redis.String(conn.Do("cluster", "failover", "force"))
	if err != nil {
		return "", err
	}

	return resp, nil
}

/// Info

type RedisInfo map[string]string

func FetchInfo(addr, section string) (*RedisInfo, error) {
	conn, err := redis.Dial("tcp", addr)
	if err != nil {
		return nil, ErrConnFailed
	}
	defer conn.Close()

	resp, err := redis.String(conn.Do("info", section))
	if err != nil {
		return nil, err
	}

	infomap := map[string]string{}
	lines := strings.Split(resp, "\r\n")
	for _, line := range lines {
		xs := strings.Split(line, ":")
		if len(xs) != 2 {
			continue
		}
		key := xs[0]
		value := xs[1]
		infomap[key] = value
	}

	redisInfo := RedisInfo(infomap)
	return &redisInfo, nil
}

func (info *RedisInfo) Get(key string) string {
	return (*info)[key]
}

func (info *RedisInfo) GetInt64(key string) (int64, error) {
	return strconv.ParseInt((*info)[key], 10, 64)
}

/// Migrate

func SetSlot(addr string, slot int, action, toId string) error {
	conn, err := redis.Dial("tcp", addr)
	if err != nil {
		return ErrConnFailed
	}
	defer conn.Close()

	if action == SLOT_STABLE {
		_, err = redis.String(conn.Do("cluster", "setslot", slot, action))
	} else {
		_, err = redis.String(conn.Do("cluster", "setslot", slot, action, toId))
	}

	if err != nil {
		return err
	}
	return nil
}

func GetKeysInSlot(addr string, slot, num int) ([]string, error) {
	conn, err := redis.Dial("tcp", addr)
	if err != nil {
		return nil, ErrConnFailed
	}
	defer conn.Close()

	resp, err := redis.Strings(conn.Do("cluster", "getkeysinslot", slot, num))
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func Migrate(addr, toIp string, toPort int, key string, timeout int) (string, error) {
	conn, err := redis.Dial("tcp", addr)
	if err != nil {
		return "", ErrConnFailed
	}
	defer conn.Close()

	resp, err := redis.String(conn.Do("migrate", toIp, toPort, key, 0, timeout))
	if err != nil {
		return "", err
	}
	return resp, nil
}
