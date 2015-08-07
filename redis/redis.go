package redis

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/ksarch-saas/cc/log"
	"github.com/ksarch-saas/cc/topo"
)

var (
	ErrNotCutter   = errors.New("redis: the server is not a cutter")
	ErrConnFailed  = errors.New("redis: connection error")
	ErrPingFailed  = errors.New("redis: ping error")
	ErrServer      = errors.New("redis: server error")
	ErrInvalidAddr = errors.New("redis: invalid address string")
	poolMap        map[string]*redis.Pool //redis connection pool for each server
)

const (
	SLOT_MIGRATING = "MIGRATING"
	SLOT_IMPORTING = "IMPORTING"
	SLOT_STABLE    = "STABLE"
	SLOT_NODE      = "NODE"

	NUM_RETRY     = 3
	CONN_TIMEOUT  = 1 * time.Second
	READ_TIMEOUT  = 2 * time.Second
	WRITE_TIMEOUT = 2 * time.Second
)

func dial(addr string) (redis.Conn, error) {
	if poolMap == nil {
		poolMap = make(map[string]*redis.Pool)
	}

	inner := func(addr string) (redis.Conn, error) {
		if _, ok := poolMap[addr]; !ok {
			//not exist in map
			poolMap[addr] = &redis.Pool{
				MaxIdle:     3,
				IdleTimeout: 240 * time.Second,
				Dial: func() (redis.Conn, error) {
					c, err := redis.DialTimeout("tcp", addr, CONN_TIMEOUT, READ_TIMEOUT, WRITE_TIMEOUT)
					if err != nil {
						return nil, err
					}
					return c, nil
				},
				TestOnBorrow: func(c redis.Conn, t time.Time) error {
					_, err := c.Do("PING")
					return err
				},
			}
		}
		pool, ok := poolMap[addr]
		if ok {
			return pool.Get(), nil
		} else {
			return nil, ErrConnFailed
		}
	}
	resp, err := inner(addr)
	if err == nil {
		return resp, nil
	}
	return nil, err
}

/// Misc

func IsAlive(addr string) bool {
	conn, err := dial(addr)
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
	conn, err := dial(addr)
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
		time.Sleep(5 * time.Second)
		if err == nil {
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
	inner := func(addr string) (string, error) {
		conn, err := dial(addr)
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
	retry := NUM_RETRY
	var err error
	var resp string
	for retry > 0 {
		resp, err = inner(addr)
		if err == nil {
			return resp, nil
		}
		retry--
	}
	return "", err
}

func FetchClusterInfo(addr string) (topo.ClusterInfo, error) {
	clusterInfo := topo.ClusterInfo{}
	conn, err := dial(addr)
	if err != nil {
		return clusterInfo, ErrConnFailed
	}
	defer conn.Close()

	resp, err := redis.String(conn.Do("cluster", "info"))
	if err != nil {
		return clusterInfo, err
	}

	lines := strings.Split(resp, "\n")
	for _, line := range lines {
		xs := strings.Split(strings.TrimSpace(line), ":")
		if len(xs) != 2 {
			continue
		}
		switch xs[0] {
		case "cluster_state":
			clusterInfo.ClusterState = xs[1]
		case "cluster_slots_assigned":
			n, _ := strconv.Atoi(xs[1])
			clusterInfo.ClusterSlotsAssigned = n
		case "cluster_slots_ok":
			n, _ := strconv.Atoi(xs[1])
			clusterInfo.ClusterSlotsOk = n
		case "cluster_slots_pfail":
			n, _ := strconv.Atoi(xs[1])
			clusterInfo.ClusterSlotsPfail = n
		case "cluster_slots_fail":
			n, _ := strconv.Atoi(xs[1])
			clusterInfo.ClusterSlotsFail = n
		case "cluster_known_nodes":
			n, _ := strconv.Atoi(xs[1])
			clusterInfo.ClusterKnownNodes = n
		case "cluster_size":
			n, _ := strconv.Atoi(xs[1])
			clusterInfo.ClusterSize = n
		case "cluster_current_epoch":
			n, _ := strconv.Atoi(xs[1])
			clusterInfo.ClusterCurrentEpoch = n
		case "cluster_my_epoch":
			n, _ := strconv.Atoi(xs[1])
			clusterInfo.ClusterMyEpoch = n
		case "cluster_stats_messages_sent":
			n, _ := strconv.Atoi(xs[1])
			clusterInfo.ClusterStatsMessagesSent = n
		case "cluster_stats_messages_received":
			n, _ := strconv.Atoi(xs[1])
			clusterInfo.ClusterStatsMessagesReceived = n
		}
	}
	return clusterInfo, nil
}

func ClusterChmod(addr, id, op string) (string, error) {
	inner := func(addr, id, op string) (string, error) {
		conn, err := dial(addr)
		if err != nil {
			return "", ErrConnFailed
		}
		defer conn.Close()

		resp, err := redis.String(conn.Do("cluster", "chmod", op, id))
		if err != nil {
			return "", err
		}
		return resp, nil
	}
	retry := NUM_RETRY
	var err error
	var resp string
	for retry > 0 {
		resp, err = inner(addr, id, op)
		if err == nil {
			return resp, nil
		}
		retry--
	}
	return "", err
}

func DisableRead(addr, id string) (string, error) {
	return ClusterChmod(addr, id, "-r")
}

func EnableRead(addr, id string) (string, error) {
	return ClusterChmod(addr, id, "+r")
}

func DisableWrite(addr, id string) (string, error) {
	return ClusterChmod(addr, id, "-w")
}

func EnableWrite(addr, id string) (string, error) {
	return ClusterChmod(addr, id, "+w")
}

func ClusterFailover(addr string) (string, error) {
	conn, err := dial(addr)
	if err != nil {
		return "", ErrConnFailed
	}
	defer conn.Close()

	// 先正常Failover试试，如果主挂了再试试Force
	resp, err := redis.String(conn.Do("cluster", "failover"))
	if err != nil {
		if strings.HasPrefix(err.Error(), "ERR Master is down or failed") {
			resp, err = redis.String(conn.Do("cluster", "failover", "force"))
		}
		if err != nil {
			return "", err
		}
	}
	// 30s
	for i := 0; i < 30; i++ {
		info, err := FetchInfo(addr, "Replication")
		if err != nil {
			return resp, err
		}
		if info.Get("role") == "slave" {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	return resp, nil
}

func ClusterTakeover(addr string) (string, error) {
	conn, err := dial(addr)
	if err != nil {
		return "", ErrConnFailed
	}
	defer conn.Close()

	resp, err := redis.String(conn.Do("cluster", "failover", "takeover"))
	if err != nil {
		return "", err
	}

	// 30s
	for i := 0; i < 30; i++ {
		info, err := FetchInfo(addr, "Replication")
		if err != nil {
			return resp, err
		}
		if info.Get("role") == "slave" {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	return resp, nil
}

func ClusterReplicate(addr, targetId string) (string, error) {
	conn, err := dial(addr)
	if err != nil {
		return "", ErrConnFailed
	}
	defer conn.Close()

	resp, err := redis.String(conn.Do("cluster", "replicate", targetId))
	if err != nil {
		return "", err
	}

	return resp, nil
}

func ClusterMeet(seedAddr, newIp string, newPort int) (string, error) {
	conn, err := dial(seedAddr)
	if err != nil {
		return "", ErrConnFailed
	}
	defer conn.Close()

	resp, err := redis.String(conn.Do("cluster", "meet", newIp, newPort))
	if err != nil {
		return "", err
	}

	return resp, nil
}

func ClusterForget(seedAddr, nodeId string) (string, error) {
	conn, err := dial(seedAddr)
	if err != nil {
		return "", ErrConnFailed
	}
	defer conn.Close()

	resp, err := redis.String(conn.Do("cluster", "forget", nodeId))
	if err != nil {
		return "", err
	}

	return resp, nil
}

func ClusterReset(addr string, hard bool) (string, error) {
	conn, err := dial(addr)
	if err != nil {
		return "", ErrConnFailed
	}
	defer conn.Close()

	flag := "soft"
	if hard {
		flag = "hard"
	}

	resp, err := redis.String(conn.Do("cluster", "reset", flag))
	if err != nil {
		return "", err
	}

	return resp, nil
}

/// Info

type RedisInfo map[string]string

func FetchInfo(addr, section string) (*RedisInfo, error) {
	inner := func(addr, section string) (*RedisInfo, error) {
		conn, err := dial(addr)
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
	retry := NUM_RETRY
	var err error
	var redisInfo *RedisInfo
	for retry > 0 {
		redisInfo, err = inner(addr, section)
		if err == nil {
			return redisInfo, nil
		}
		retry--
	}
	return nil, err
}

func (info *RedisInfo) Get(key string) string {
	return (*info)[key]
}

func (info *RedisInfo) GetInt64(key string) (int64, error) {
	return strconv.ParseInt((*info)[key], 10, 64)
}

/// Migrate

func SetSlot(addr string, slot int, action, toId string) error {
	conn, err := dial(addr)
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

func CountKeysInSlot(addr string, slot int) (int, error) {
	inner := func(addr string, slot int) (int, error) {
		conn, err := dial(addr)
		if err != nil {
			return 0, ErrConnFailed
		}
		defer conn.Close()

		resp, err := redis.Int(conn.Do("cluster", "countkeysinslot", slot))
		if err != nil {
			return 0, err
		}
		return resp, nil
	}
	retry := NUM_RETRY
	var err error
	var resp int
	for retry > 0 {
		resp, err = inner(addr, slot)
		if err == nil {
			return resp, nil
		}
		retry--
	}
	return 0, err
}

func GetKeysInSlot(addr string, slot, num int) ([]string, error) {
	inner := func(addr string, slot, num int) ([]string, error) {
		conn, err := dial(addr)
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
	retry := NUM_RETRY
	var err error
	var resp []string
	for retry > 0 {
		resp, err = inner(addr, slot, num)
		if err == nil {
			return resp, nil
		}
		retry--
	}
	return nil, err
}

func Migrate(addr, toIp string, toPort int, key string, timeout int) (string, error) {
	inner := func(addr, toIp string, toPort int, key string, timeout int) (string, error) {
		conn, err := dial(addr)
		if err != nil {
			return "", ErrConnFailed
		}
		defer conn.Close()

		resp, err := redis.String(conn.Do("migrate", toIp, toPort, key, 0, timeout))
		if err != nil && strings.Contains(err.Error(), "BUSYKEY") {
			log.Warningf("Migrate", "Found BUSYKEY '%s', will overwrite it.", key)
			resp, err = redis.String(conn.Do("migrate", toIp, toPort, key, 0, timeout, "replace"))
		}
		if err != nil {
			return "", err
		}
		return resp, nil
	}
	retry := NUM_RETRY
	var err error
	var resp string
	for retry > 0 {
		resp, err = inner(addr, toIp, toPort, key, timeout)
		if err == nil {
			return resp, nil
		}
		retry--
	}
	return "", err
}

// used by cli
func ClusterNodesWithoutExtra(addr string) (string, error) {
	inner := func(addr string) (string, error) {
		conn, err := dial(addr)
		if err != nil {
			return "", ErrConnFailed
		}
		defer conn.Close()

		resp, err := redis.String(conn.Do("cluster", "nodes"))
		if err != nil {
			return "", err
		}
		return resp, nil
	}
	retry := NUM_RETRY
	var err error
	var resp string
	for retry > 0 {
		resp, err = inner(addr)
		if err == nil {
			return resp, nil
		}
		retry--
	}
	return "", err
}

func AddSlotRange(addr string, start, end int) (string, error) {
	conn, err := dial(addr)
	if err != nil {
		return "connect failed", ErrConnFailed
	}
	defer conn.Close()
	var resp string
	for i := start; i <= end; i++ {
		resp, err = redis.String(conn.Do("cluster", "addslots", i))
		if err != nil {
			return resp, err
		}
	}
	return resp, nil
}

func FlushAll(addr string) (string, error) {
	conn, err := dial(addr)
	if err != nil {
		return "connect failed", ErrConnFailed
	}
	defer conn.Close()
	resp, err := redis.String(conn.Do("flushall"))
	if err != nil {
		return resp, err
	}
	return resp, nil
}
