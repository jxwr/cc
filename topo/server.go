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

type Server struct {
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

func NewServerFromString(addr string) *Server {
	xs := strings.Split(addr, ":")
	if len(xs) != 2 {
		return nil
	}
	port, err := strconv.Atoi(xs[1])
	if err != nil {
		return nil
	}
	return NewServer(xs[0], port)
}

func NewServer(ip string, port int) *Server {
	matched, _ := regexp.MatchString("\\d+\\.\\d+\\.\\d+\\.\\d+", ip)
	if !matched {
		// 'ip' is a hostname
		ips, err := net.LookupIP(ip)
		if err != nil {
			panic("can not resolve address of " + ip)
		}
		ip = ips[0].String()
	}

	server := Server{ip: ip, port: port, ranges: []Range{}}
	return &server
}

func (s *Server) Addr() string {
	return fmt.Sprintf("%s:%d", s.ip, s.port)
}

func (s *Server) Ip() string {
	return s.ip
}

func (s *Server) Port() int {
	return s.port
}

func (s *Server) Id() string {
	return s.id
}

func (s *Server) SetId(id string) {
	s.id = id
}

func (s *Server) ParentId() string {
	return s.parentid
}

func (s *Server) SetParentId(pid string) {
	s.parentid = pid
}

func (s *Server) Migrating() bool {
	return s.migrating
}

func (s *Server) SetMigrating(val bool) {
	s.migrating = val
}

func (s *Server) Readable() bool {
	return s.readable
}

func (s *Server) SetReadable(val bool) {
	s.readable = val
}

func (s *Server) Writable() bool {
	return s.writable
}

func (s *Server) SetWritable(val bool) {
	s.writable = val
}

func (s *Server) PFail() bool {
	return s.pfail
}

func (s *Server) SetPFail(val bool) {
	s.pfail = val
}

func (s *Server) Fail() bool {
	return s.fail
}

func (s *Server) SetFail(val bool) {
	s.fail = val
}

func (s *Server) PFailCount() int {
	return s.failcount
}

func (s *Server) IncrPFailCount() {
	s.failcount++
}

func (s *Server) IsMaster() bool {
	return s.role == "master"
}

func (s *Server) SetRole(val string) {
	s.role = val
}

func (s *Server) Tag() string {
	return s.tag
}

func (s *Server) SetTag(val string) {
	s.tag = val
}

func (s *Server) Region() string {
	return s.region
}

func (s *Server) SetRegion(val string) {
	s.region = val
}

func (s *Server) Zone() string {
	return s.zone
}

func (s *Server) SetZone(val string) {
	s.zone = val
}

func (s *Server) Room() string {
	return s.room
}

func (s *Server) SetRoom(val string) {
	s.room = val
}

func (s *Server) Ranges() []Range {
	return s.ranges
}

func (s *Server) AddRange(r Range) {
	s.ranges = append(s.ranges, r)
}

func (s *Server) Compare(t *Server) bool {
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

func (s *Server) Hostname() string {
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

func (s *Server) MachineRoom() string {
	xs := strings.Split(s.Hostname(), ".")
	if len(xs) != 2 {
		panic("invalid host name: " + s.Hostname())
	}
	return xs[1]
}
