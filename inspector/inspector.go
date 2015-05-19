package inspector

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/glog"
	"github.com/jxwr/cc/meta"
	"github.com/jxwr/cc/redis"
	"github.com/jxwr/cc/topo"
)

var (
	ErrNoSeed           = errors.New("inspector: no seed node found")
	ErrInvalidTag       = errors.New("inspector: invalid tag")
	ErrEmptyTag         = errors.New("inspector: empty tag")
	ErrSeedIsFreeNode   = errors.New("inspector: seed is free node")
	ErrNodeNotExist     = errors.New("inspector: node not exist")
	ErrNodesInfoNotSame = errors.New("inspector: cluster nodes info of seeds are different")
)

type Inspector struct {
	mutex       *sync.RWMutex
	LocalRegion string
	Seeds       []*topo.Node
	SeedIndex   int
	ClusterTopo *topo.Cluster
}

func NewInspector(seeds []*topo.Node) *Inspector {
	sp := &Inspector{
		mutex:       &sync.RWMutex{},
		Seeds:       seeds,
		LocalRegion: meta.LocalRegion(),
	}
	return sp
}

func (self *Inspector) buildNode(line string) (*topo.Node, bool, error) {
	xs := strings.Split(line, " ")
	mod, tag, id, addr, flags, parent := xs[0], xs[1], xs[2], xs[3], xs[4], xs[5]
	node := topo.NewNodeFromString(addr)
	ranges := []string{}
	for _, word := range xs[10:] {
		if strings.HasPrefix(word, "[") {
			word = word[1 : len(word)-1]
			xs := strings.Split(word, "->-")
			if len(xs) == 2 {
				slot, _ := strconv.Atoi(xs[0])
				node.AddMigrating(xs[1], slot)
			}
			xs = strings.Split(word, "-<-")
			if len(xs) == 2 {
				slot, _ := strconv.Atoi(xs[0])
				node.AddImporting(xs[1], slot)
			}
			continue
		}
		ranges = append(ranges, word)
	}

	for _, r := range ranges {
		xs = strings.Split(r, "-")
		if len(xs) == 2 {
			left, _ := strconv.Atoi(xs[0])
			right, _ := strconv.Atoi(xs[1])
			node.AddRange(topo.Range{left, right})
		} else {
			left, _ := strconv.Atoi(r)
			node.AddRange(topo.Range{left, left})
		}
	}

	// basic info
	node.SetId(id)
	node.SetParentId(parent)
	node.SetTag(tag)
	node.SetReadable(mod[0] == 'r')
	node.SetWritable(mod[1] == 'w')
	myself := false
	if strings.Contains(flags, "myself") {
		myself = true
	}
	if strings.Contains(flags, "master") {
		node.SetRole("master")
	} else {
		node.SetRole("slave")
	}
	if strings.Contains(flags, "fail?") {
		node.SetPFail(true)
		node.IncrPFailCount()
	}
	xs = strings.Split(tag, ":")
	if len(xs) == 3 {
		node.SetRegion(xs[0])
		node.SetZone(xs[1])
		node.SetRoom(xs[2])
	} else if node.Tag != "-" {
		return nil, myself, ErrInvalidTag
	}

	return node, myself, nil
}

func (self *Inspector) MeetNode(node *topo.Node) {
	for _, seed := range self.Seeds {
		if seed.Ip == node.Ip && seed.Port == node.Port {
			continue
		}
		_, err := redis.ClusterMeet(seed.Addr(), node.Ip, node.Port)
		if err == nil {
			break
		}
	}
}

func (self *Inspector) initClusterTopo(seed *topo.Node) (*topo.Cluster, error) {
	resp, err := redis.ClusterNodes(seed.Addr())
	if err != nil {
		return nil, err
	}

	cluster := topo.NewCluster(self.LocalRegion)

	var summary topo.SummaryInfo
	lines := strings.Split(resp, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			summary.ReadLine(line)
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		node, myself, err := self.buildNode(line)
		if err != nil {
			return nil, err
		}
		if node.Ip == "127.0.0.1" {
			node.Ip = seed.Ip
		}
		// 遇到myself，读取该节点的ClusterInfo
		if myself {
			info, err := redis.FetchClusterInfo(node.Addr())
			if err != nil {
				return nil, err
			}
			node.ClusterInfo = info
			node.SummaryInfo = summary
		}
		cluster.AddNode(node)
	}

	return cluster, nil
}

func (self *Inspector) isFreeNode(seed *topo.Node) (bool, *topo.Node) {
	resp, err := redis.ClusterNodes(seed.Addr())
	if err != nil {
		return false, nil
	}
	numNode := 0
	lines := strings.Split(resp, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "# ") {
			continue
		}
		numNode++
	}
	if numNode != 1 {
		return false, nil
	}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "# ") {
			continue
		}
		node, myself, err := self.buildNode(line)
		// 只看到自己，是主，且没有slots，才认为是FreeNode
		if !myself {
			return false, nil
		}
		if err != nil || len(node.Ranges) > 0 || !node.IsMaster() {
			return false, nil
		} else {
			return true, node
		}
	}
	return false, nil
}

func (self *Inspector) checkClusterTopo(seed *topo.Node, cluster *topo.Cluster) error {
	resp, err := redis.ClusterNodes(seed.Addr())
	if err != nil {
		return err
	}

	var summary topo.SummaryInfo
	lines := strings.Split(resp, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			summary.ReadLine(line)
			continue
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		s, myself, err := self.buildNode(line)
		if err != nil {
			return err
		}
		if s.Ip == "127.0.0.1" {
			s.Ip = seed.Ip
		}
		node := cluster.FindNode(s.Id)
		if node == nil {
			if s.PFail {
				glog.Warningf("forget dead node %s(%s)", s.Id, s.Addr())
				redis.ClusterForget(seed.Addr(), s.Id)
			}
			return fmt.Errorf("node not exist %s(%s)", s.Id, s.Addr())
		}

		// 对比节点数据是否相同
		if !node.Compare(s) {
			glog.Infof("%#v vs %#v different", s, node)
			if s.Tag == "-" && node.Tag != "-" {
				// 可能存在处于不被Cluster接受的节点，节点可以看见Cluster，但Cluster看不到它。
				// 一种复现情况情况：某个节点已经死了，系统将其Forget，但是OP并未被摘除该节点，
				// 而是恢复了该节点。
				glog.Warningf("remeet node %s", seed.Addr())
				self.MeetNode(seed)
			}
			return ErrNodesInfoNotSame
		}
		if len(node.Ranges) == 0 && len(s.Ranges) > 0 {
			glog.Warningf("Ranges not equal, use nonempty ranges.")
			node.Ranges = s.Ranges
		}

		if myself {
			info, err := redis.FetchClusterInfo(node.Addr())
			if err != nil {
				return err
			}
			node.ClusterInfo = info
			node.SummaryInfo = summary
		}

		if len(s.Migrating) != 0 {
			node.Migrating = s.Migrating
		}
		if len(s.Importing) != 0 {
			node.Importing = s.Importing
		}

		if s.PFail {
			node.IncrPFailCount()
		}
	}
	return nil
}

func (self *Inspector) HasSeed(seed *topo.Node) bool {
	for _, s := range self.Seeds {
		if s.Addr() == seed.Addr() {
			if s.Id == "" {
				*s = *seed
			}
			return true
		}
	}
	return false
}

func (self *Inspector) MergeSeeds(seeds []*topo.Node) {
	for _, seed := range seeds {
		if !self.HasSeed(seed) {
			self.Seeds = append(self.Seeds, seed)
		}
	}
}

// 生成ClusterSnapshot
func (self *Inspector) BuildClusterTopo() (*topo.Cluster, []*topo.Node, error) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if len(self.Seeds) == 0 {
		return nil, nil, ErrNoSeed
	}

	// 过滤掉连接不上的节点
	seeds := []*topo.Node{}
	for _, s := range self.Seeds {
		if redis.IsAlive(s.Addr()) {
			seeds = append(seeds, s)
		}
	}

	if len(seeds) == 0 {
		return nil, seeds, ErrNoSeed
	}

	// 顺序选一个节点，获取nodes数据作为基准，再用其他节点的数据与基准做对比
	if self.SeedIndex >= len(seeds) {
		self.SeedIndex = len(seeds) - 1
	}
	var seed *topo.Node
	for i := 0; i < len(seeds); i++ {
		seed = seeds[self.SeedIndex]
		self.SeedIndex++
		self.SeedIndex %= len(seeds)
		if seed.Free {
			glog.Info("Seed node is free, ", seed.Addr())
		} else {
			break
		}
	}
	cluster, err := self.initClusterTopo(seed)
	if err != nil {
		return nil, seeds, err
	}

	// 检查所有节点返回的信息是不是相同，如果不同说明正在变化中，直接返回等待重试
	if len(seeds) > 1 {
		for _, s := range seeds {
			if s == seed {
				continue
			}
			err := self.checkClusterTopo(s, cluster)
			if err != nil {
				free, node := self.isFreeNode(s)
				if free {
					node.Free = true
					glog.Infof("Found free node %s", node.Addr())
					cluster.AddNode(node)
				} else {
					return cluster, seeds, err
				}
			} else {
				s.Free = false
			}
		}
	}

	// 构造LocalRegion视图
	for _, s := range cluster.LocalRegionNodes() {
		if s.PFailCount() > cluster.NumLocalRegionNode()/2 {
			glog.Infof("Found %d/%d PFAIL state on %s, set FAIL",
				s.PFailCount(), cluster.NumLocalRegionNode(), s.Addr())
			s.SetFail(true)
		}
	}

	cluster.BuildReplicaSets()

	self.MergeSeeds(cluster.LocalRegionNodes())
	self.ClusterTopo = cluster
	return cluster, seeds, nil
}
