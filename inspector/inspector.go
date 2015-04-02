package inspector

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/jxwr/cc/meta"
	"github.com/jxwr/cc/redis"
	"github.com/jxwr/cc/topo"
)

var (
	ErrNoSeed           = errors.New("inspector: no seed node found")
	ErrInvalidTag       = errors.New("inspector: invalid tag")
	ErrNodeNotExist     = errors.New("inspector: node not exist")
	ErrNodesInfoNotSame = errors.New("inspector: cluster nodes info of seeds are different")
)

type Inspector struct {
	mutex       *sync.RWMutex
	LocalRegion string
	Seeds       []*topo.Node
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

func (self *Inspector) buildNode(line string) (*topo.Node, error) {
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
	if len(xs) != 3 {
		return nil, ErrInvalidTag
	}
	node.SetRegion(xs[0])
	node.SetZone(xs[1])
	node.SetRoom(xs[2])

	return node, nil
}

func (self *Inspector) initClusterTopo(seed *topo.Node) (*topo.Cluster, error) {
	resp, err := redis.ClusterNodes(seed.Addr())
	if err != nil {
		return nil, err
	}

	cluster := topo.NewCluster(self.LocalRegion)

	lines := strings.Split(resp, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		node, err := self.buildNode(line)
		if err != nil {
			return nil, err
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
		if line == "" {
			continue
		}
		numNode++
	}
	if numNode != 1 {
		return false, nil
	}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		node, err := self.buildNode(line)
		// 只看到自己，是主，且没有slots，才认为是FreeNode
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

	lines := strings.Split(resp, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		s, err := self.buildNode(line)
		if err != nil {
			return err
		}

		node := cluster.FindNode(s.Id)
		if node == nil {
			return ErrNodeNotExist
		}

		// 对比节点数据是否相同
		if !node.Compare(s) {
			fmt.Println(s)
			fmt.Println(node)
			return ErrNodesInfoNotSame
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
		if s.Id == seed.Id {
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
func (self *Inspector) BuildClusterTopo() (*topo.Cluster, error) {
	self.mutex.Lock()
	defer self.mutex.Unlock()

	if len(self.Seeds) == 0 {
		return nil, ErrNoSeed
	}

	// 过滤掉连接不上的节点
	seeds := []*topo.Node{}
	for _, s := range self.Seeds {
		if redis.IsAlive(s.Addr()) {
			seeds = append(seeds, s)
		}
	}

	if len(seeds) == 0 {
		return nil, ErrNoSeed
	}

	// 随机选一个节点，获取nodes数据作为基准，再用其他节点的数据与基准做对比
	seed := seeds[0]
	cluster, err := self.initClusterTopo(seed)
	if err != nil {
		return nil, err
	}

	// 检查所有节点返回的信息是不是相同，如果不同说明正在变化中，直接返回等待重试
	if len(seeds) > 1 {
		for _, seed := range seeds[1:] {
			err := self.checkClusterTopo(seed, cluster)
			if err != nil {
				free, node := self.isFreeNode(seed)
				if free {
					node.Free = true
					log.Println("Found free node", node.Addr())
					cluster.AddNode(node)
				} else {
					return nil, err
				}
			}
		}
	}

	// 构造LocalRegion视图
	for _, s := range cluster.LocalRegionNodes() {
		if s.PFailCount() > cluster.NumLocalRegionNode()/2 {
			log.Printf("Found %d/%d PFAIL state on %s, set FAIL",
				s.PFailCount(), cluster.NumLocalRegionNode(), s.Addr())
			s.SetFail(true)
		}
	}

	cluster.BuildReplicaSets()

	self.MergeSeeds(cluster.LocalRegionNodes())
	self.ClusterTopo = cluster

	for _, se := range self.Seeds {
		log.Println("Seed:", se.Addr())
	}
	return cluster, nil
}
