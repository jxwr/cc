package topo

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

type Range struct {
	Left  int
	Right int
}

type Node struct {
	ip        string
	port      int
	id        string
	parentid  string
	migrating bool
	readable  bool
	writable  bool
	pfail     bool
	fail      bool
	role      string
	tag       string
	region    string
	zone      string
	room      string
	ranges    []Range
	failcount int
	hostname  string
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

	node := Node{ip: ip, port: port, ranges: []Range{}}
	return &node
}

func (s *Node) Addr() string {
	return fmt.Sprintf("%s:%d", s.ip, s.port)
}

func (s *Node) Ip() string {
	return s.ip
}

func (s *Node) Port() int {
	return s.port
}

func (s *Node) Id() string {
	return s.id
}

func (s *Node) SetId(id string) *Node {
	s.id = id
	return s
}

func (s *Node) ParentId() string {
	return s.parentid
}

func (s *Node) SetParentId(pid string) *Node {
	s.parentid = pid
	return s
}

func (s *Node) Migrating() bool {
	return s.migrating
}

func (s *Node) SetMigrating(val bool) *Node {
	s.migrating = val
	return s
}

func (s *Node) Readable() bool {
	return s.readable
}

func (s *Node) SetReadable(val bool) *Node {
	s.readable = val
	return s
}

func (s *Node) Writable() bool {
	return s.writable
}

func (s *Node) SetWritable(val bool) *Node {
	s.writable = val
	return s
}

func (s *Node) PFail() bool {
	return s.pfail
}

func (s *Node) SetPFail(val bool) *Node {
	s.pfail = val
	return s
}

func (s *Node) Fail() bool {
	return s.fail
}

func (s *Node) SetFail(val bool) *Node {
	s.fail = val
	return s
}

func (s *Node) PFailCount() int {
	return s.failcount
}

func (s *Node) IncrPFailCount() {
	s.failcount++
}

func (s *Node) IsMaster() bool {
	return s.role == "master"
}

func (s *Node) SetRole(val string) *Node {
	s.role = val
	return s
}

func (s *Node) Role() string {
	return s.role
}

func (s *Node) Tag() string {
	return s.tag
}

func (s *Node) SetTag(val string) *Node {
	s.tag = val
	return s
}

func (s *Node) Region() string {
	return s.region
}

func (s *Node) SetRegion(val string) *Node {
	s.region = val
	return s
}

func (s *Node) Zone() string {
	return s.zone
}

func (s *Node) SetZone(val string) *Node {
	s.zone = val
	return s
}

func (s *Node) Room() string {
	return s.room
}

func (s *Node) SetRoom(val string) *Node {
	s.room = val
	return s
}

func (s *Node) Ranges() []Range {
	return s.ranges
}

func (s *Node) AddRange(r Range) {
	s.ranges = append(s.ranges, r)
}

func (s *Node) Compare(t *Node) bool {
	b := true
	b = b && (s.port == t.port)
	b = b && (s.parentid == t.parentid)
	b = b && (s.readable == t.readable)
	b = b && (s.writable == t.writable)
	b = b && (s.role == t.role)
	b = b && (s.tag == t.tag)
	b = b && (len(s.ranges) == len(t.ranges))

	if b == false {
		return false
	}

	for i, r := range s.ranges {
		if r != t.ranges[i] {
			fmt.Println("EE")
			return false
		}
	}

	return true
}

func (s *Node) Hostname() string {
	if s.hostname == "" {
		addr, err := net.LookupAddr(s.ip)
		if len(addr) == 0 || err != nil {
			panic("unknown host for " + s.ip)
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
