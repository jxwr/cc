package redis

import (
	"errors"
	"strings"

	"github.com/garyburd/redigo/redis"
)

var (
	ErrNotCutter   = errors.New("redis: the server is not a cutter")
	ErrConnFailed  = errors.New("redis: connection error")
	ErrPingFailed  = errors.New("redis: ping error")
	ErrServer      = errors.New("redis: server error")
	ErrInvalidAddr = errors.New("redis: invalid address string")
)

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

type RedisInfo map[string]string

func FetchInfo(addr string) (*RedisInfo, error) {
	conn, err := redis.Dial("tcp", addr)
	if err != nil {
		return nil, ErrConnFailed
	}
	defer conn.Close()

	resp, err := redis.String(conn.Do("info"))
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
