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

func (r Range) String() string {
	if r.Left == r.Right {
		return fmt.Sprint(r.Left)
	}
	return fmt.Sprintf("%d-%d", r.Left, r.Right)
}

func (r Range) NumSlots() int {
	return (r.Right - r.Left + 1)
}

type Ranges []Range

func (rs Ranges) String() string {
	slots := ""
	for _, r := range rs {
		slots += r.String() + ","
	}
	if len(slots) > 0 {
		return slots[:len(slots)-1]
	}
	return slots
}

func (rs Ranges) NumSlots() int {
	sum := 0
	for _, r := range rs {
		sum += r.NumSlots()
	}
	return sum
}

type SummaryInfo struct {
	UsedMemory              int64
	Keys                    int64
	Expires                 int64
	MasterLinkStatus        string
	MasterSyncLeftBytes     int64
	ReplOffset              int64
	Loading                 bool
	RdbBgsaveInProgress     bool
	InstantaneousOpsPerSec  int
	InstantaneousInputKbps  float64
	InstantaneousOutputKbps float64
}

func (s *SummaryInfo) ReadLine(line string) {
	xs := strings.Split(strings.TrimSpace(line), " ")
	xs = strings.Split(xs[1], ":")
	s.SetField(xs[0], xs[1])
}

func (s *SummaryInfo) SetField(key, val string) {
	switch key {
	case "used_memory":
		v, _ := strconv.ParseInt(val, 10, 64)
		s.UsedMemory = v
	case "db0_keys":
		v, _ := strconv.ParseInt(val, 10, 64)
		s.Keys = v
	case "db0_expires":
		v, _ := strconv.ParseInt(val, 10, 64)
		s.Expires = v
	case "master_link_status":
		s.MasterLinkStatus = val
	case "master_sync_left_bytes":
		v, _ := strconv.ParseInt(val, 10, 64)
		s.MasterSyncLeftBytes = v
	case "repl_offset":
		v, _ := strconv.ParseInt(val, 10, 64)
		s.ReplOffset = v
	case "loading":
		s.Loading = (val == "1")
	case "rdb_bgsave_in_progress":
		s.RdbBgsaveInProgress = (val == "1")
	case "instantaneous_ops_per_sec":
		v, _ := strconv.Atoi(val)
		s.InstantaneousOpsPerSec = v
	case "instantaneous_input_kbps":
		v, _ := strconv.ParseFloat(val, 64)
		s.InstantaneousInputKbps = v
	case "instantaneous_output_kbps":
		v, _ := strconv.ParseFloat(val, 64)
		s.InstantaneousOutputKbps = v
	}
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
	SummaryInfo
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

func (s *Node) String() string {
	return fmt.Sprintf("%s(%s)", s.Addr(), s.Id)
}
