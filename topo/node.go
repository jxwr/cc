package topo

import (
	"fmt"
	"math"
	"net"
	"regexp"
	"strconv"
	"strings"
)

type Range struct {
	Left  int
	Right int
}

func (r Range) NumSlots() int {
	return (r.Right - r.Left + 1)
}

type Node struct {
	Ip        string
	Port      int
	Id        string
	ParentId  string
	Migrating map[string][]int
	Importing map[string][]int
	Readable  bool
	Writable  bool
	PFail     bool
	Fail      bool
	Free      bool // 是否是游离于集群之外的节点（ClusterNodes信息里只有一个节点，主，无slots）
	Role      string
	Tag       string
	Region    string
	Zone      string
	Room      string
	Ranges    []Range
	FailCount int
	hostname  string
	ClusterInfo
}

func NewNodeFromString(addr string) *Node {
	xs := strings.Split(addr, ":")
	if len(xs) != 2 {
		return nil
	}
	port, err := strconv.Atoi(xs[1])
	if err != nil {
		return nil
	}
	if xs[0] == "" {
		xs[0] = "127.0.0.1"
	}
	return NewNode(xs[0], port)
}

func NewNode(ip string, port int) *Node {
	matched, _ := regexp.MatchString("\\d+\\.\\d+\\.\\d+\\.\\d+", ip)
	if !matched {
		// 'ip' is a hostname
		ips, err := net.LookupIP(ip)
		if err != nil {
			panic("can not resolve address of " + ip)
		}
		ip = ips[0].String()
	}

	node := Node{
		Ip:        ip,
		Port:      port,
		Ranges:    []Range{},
		Migrating: map[string][]int{},
		Importing: map[string][]int{},
	}
	return &node
}

func (s *Node) Addr() string {
	return fmt.Sprintf("%s:%d", s.Ip, s.Port)
}

func (s *Node) SetId(id string) *Node {
	s.Id = id
	return s
}

func (s *Node) SetParentId(pid string) *Node {
	s.ParentId = pid
	return s
}

func (s *Node) AddMigrating(nodeId string, slot int) *Node {
	s.Migrating[nodeId] = append(s.Migrating[nodeId], slot)
	return s
}

func (s *Node) AddImporting(nodeId string, slot int) *Node {
	s.Importing[nodeId] = append(s.Importing[nodeId], slot)
	return s
}

func (s *Node) SetReadable(val bool) *Node {
	s.Readable = val
	return s
}

func (s *Node) SetWritable(val bool) *Node {
	s.Writable = val
	return s
}

func (s *Node) SetPFail(val bool) *Node {
	s.PFail = val
	return s
}

func (s *Node) SetFail(val bool) *Node {
	s.Fail = val
	return s
}

func (s *Node) PFailCount() int {
	return s.FailCount
}

func (s *Node) IncrPFailCount() {
	s.FailCount++
}

func (s *Node) IsMaster() bool {
	return s.Role == "master"
}

func (s *Node) IsStandbyMaster() bool {
	return (s.Role == "master" && s.Fail && len(s.Ranges) == 0)
}

func (s *Node) SetRole(val string) *Node {
	s.Role = val
	return s
}

func (s *Node) SetTag(val string) *Node {
	s.Tag = val
	return s
}

func (s *Node) SetRegion(val string) *Node {
	s.Region = val
	return s
}

func (s *Node) SetZone(val string) *Node {
	s.Zone = val
	return s
}

func (s *Node) SetRoom(val string) *Node {
	s.Room = val
	return s
}

func (s *Node) AddRange(r Range) {
	s.Ranges = append(s.Ranges, r)
}

func (s *Node) Empty() bool {
	return len(s.Ranges) == 0
}

func (s *Node) NumSlots() int {
	total := 0
	for _, r := range s.Ranges {
		total += r.NumSlots()
	}
	return total
}

func (s *Node) RangesSplitN(n int) [][]Range {
	total := s.NumSlots()
	numSlotsPerPart := int(math.Ceil(float64(total) / float64(n)))

	parts := [][]Range{}
	ranges := []Range{}
	need := numSlotsPerPart
	for i := len(s.Ranges) - 1; i >= 0; i-- {
		rang := s.Ranges[i]
		num := rang.NumSlots()
		if need > num {
			need -= num
			ranges = append(ranges, rang)
		} else if need == num {
			ranges = append(ranges, rang)
			parts = append(parts, ranges)
			ranges = []Range{}
			need = numSlotsPerPart
		} else {
			ranges = append(ranges, Range{
				Left:  rang.Right - need + 1,
				Right: rang.Right,
			})
			parts = append(parts, ranges)
			remain := Range{rang.Left, rang.Right - need}
			ranges = []Range{remain}
			need = numSlotsPerPart - remain.NumSlots()
		}
	}
	if len(ranges) > 0 {
		parts = append(parts, ranges)
	}
	return parts
}

func (s *Node) Compare(t *Node) bool {
	b := true
	b = b && (s.Port == t.Port)
	b = b && (s.ParentId == t.ParentId)
	b = b && (s.Readable == t.Readable)
	b = b && (s.Writable == t.Writable)
	b = b && (s.Role == t.Role)
	b = b && (s.Tag == t.Tag)

	if b == false {
		return false
	}
	/*
		b = b && (len(s.Ranges) == len(t.Ranges))
		for i, r := range s.Ranges {
			if r != t.Ranges[i] {
				return false
			}
		}
	*/
	return true
}

func (s *Node) Hostname() string {
	if s.hostname == "" {
		addr, err := net.LookupAddr(s.Ip)
		if len(addr) == 0 || err != nil {
			panic("unknown host for " + s.Ip)
		}
		s.hostname = strings.TrimSuffix(addr[0], ".baidu.com")
		s.hostname = strings.TrimSuffix(addr[0], ".baidu.com.")
	}
	return s.hostname
}

func (s *Node) MachineRoom() string {
	xs := strings.Split(s.Hostname(), ".")
	if len(xs) != 2 {
		panic("invalid host name: " + s.Hostname())
	}
	return xs[1]
}
