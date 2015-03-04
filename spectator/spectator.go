package spectator

import (
	"errors"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jxwr/cc/redis"
	"github.com/jxwr/cc/topo"
)

var (
	ErrNoSeed           = errors.New("spectator: no seed server found")
	ErrInvalidTag       = errors.New("spectator: invalid tag")
	ErrServerNotExist   = errors.New("spectator: server not exist")
	ErrNodesInfoNotSame = errors.New("spectator: 'cluster nodes' info returned by seeds are different")
)

type Spectator struct {
	seeds []*topo.Server
}

func NewSpectator(seeds []*topo.Server) *Spectator {
	sp := &Spectator{
		seeds: seeds,
	}
	return sp
}

func (self *Spectator) Run() {
	tickChan := time.NewTicker(time.Second * 1).C
	for {
		select {
		case <-tickChan:
			self.BuildClusterTopo()
		}
	}
}

func (self *Spectator) buildServer(line string) (*topo.Server, error) {
	xs := strings.Split(line, " ")
	mod, tag, id, addr, flags, parent := xs[0], xs[1], xs[2], xs[3], xs[4], xs[5]
	server := topo.NewServerFromString(addr)
	ranges := []string{}
	for _, word := range xs[10:] {
		if strings.HasPrefix(word, "[") {
			server.SetMigrating(true)
			continue
		}
		ranges = append(ranges, word)
	}
	sort.Strings(ranges)

	for _, r := range ranges {
		xs = strings.Split(r, "-")
		if len(xs) == 2 {
			left, _ := strconv.Atoi(xs[0])
			right, _ := strconv.Atoi(xs[1])
			server.AddRange(topo.Range{left, right})
		}
	}

	// basic info
	server.SetId(id)
	server.SetParentId(parent)
	server.SetTag(tag)
	server.SetReadable(mod[0] == 'r')
	server.SetWritable(mod[1] == 'w')
	if strings.Contains(flags, "master") {
		server.SetRole("master")
	} else {
		server.SetRole("slave")
	}
	if strings.Contains(flags, "fail?") {
		server.SetPFail(true)
		server.IncrPFailCount()
	}
	xs = strings.Split(tag, ":")
	if len(xs) != 3 {
		return nil, ErrInvalidTag
	}
	server.SetRegion(xs[0])
	server.SetZone(xs[1])
	server.SetRoom(xs[2])

	return server, nil
}

func (self *Spectator) initClusterTopo(seed *topo.Server) (*topo.Cluster, error) {
	resp, err := redis.ClusterNodes(seed.Addr())
	if err != nil {
		return nil, err
	}

	cluster := topo.NewCluster("bj")

	lines := strings.Split(resp, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		server, err := self.buildServer(line)
		if err != nil {
			return nil, err
		}
		cluster.AddServer(server)
	}

	return cluster, nil
}

func (self *Spectator) checkClusterTopo(seed *topo.Server, cluster *topo.Cluster) error {
	resp, err := redis.ClusterNodes(seed.Addr())
	if err != nil {
		return err
	}

	lines := strings.Split(resp, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		s, err := self.buildServer(line)
		if err != nil {
			return err
		}

		server := cluster.FindServer(s.Id())
		if server == nil {
			return ErrServerNotExist
		}

		if !server.Compare(s) {
			return ErrNodesInfoNotSame
		}

		if s.PFail() {
			server.IncrPFailCount()
		}
	}

	return nil
}

func (self *Spectator) BuildClusterTopo() (*topo.Cluster, error) {
	if len(self.seeds) == 0 {
		return nil, ErrNoSeed
	}

	seeds := []*topo.Server{}
	for _, s := range self.seeds {
		if redis.IsAlive(s.Addr()) {
			seeds = append(seeds, s)
		}
	}

	if len(seeds) == 0 {
		return nil, ErrNoSeed
	}

	seed := seeds[0]
	cluster, err := self.initClusterTopo(seed)
	if err != nil {
		return nil, err
	}

	if len(seeds) > 1 {
		for _, seed := range seeds[1:] {
			err := self.checkClusterTopo(seed, cluster)
			if err != nil {
				return nil, err
			}
		}
	}

	for _, s := range cluster.RegionServers() {
		if s.PFailCount() > cluster.NumRegionServer()/2 {
			log.Printf("found %d/%d PFAIL state on %s, turning into FAIL state.",
				s.PFailCount(), cluster.NumRegionServer(), s.Addr())
			s.SetFail(true)
		}
	}

	cluster.BuildReplicaSets()

	self.seeds = cluster.RegionServers()

	return cluster, nil
}
