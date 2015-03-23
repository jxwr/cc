package main

import (
	"flag"
	"log"
	"strings"

	"github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/frontend"
	"github.com/jxwr/cc/inspector"
	"github.com/jxwr/cc/meta"
	"github.com/jxwr/cc/streams"
	"github.com/jxwr/cc/topo"
)

var (
	appName     string
	localRegion string
	seeds       string
	zkHosts     string
	httpPort    int
	wsPort      int
)

func init() {
	flag.StringVar(&appName, "appname", "", "app name")
	flag.StringVar(&localRegion, "local-region", "", "local region")
	flag.StringVar(&seeds, "seeds", "", "redis cluster seeds, seperate by comma")
	flag.StringVar(&zkHosts, "zkhosts", "", "zk hosts, seperate by comma")
	flag.IntVar(&httpPort, "http-port", 0, "http port")
	flag.IntVar(&wsPort, "ws-port", 0, "ws port")
}

func main() {
	flag.Parse()

	seedNodes := []*topo.Node{}
	for _, addr := range strings.Split(seeds, ",") {
		n := topo.NewNodeFromString(addr)
		if n == nil {
			log.Fatalf("invalid seeds %s", addr)
		}
		seedNodes = append(seedNodes, n)
	}
	if httpPort == 0 {
		log.Fatalf("invalid http port")
		flag.PrintDefaults()
	}
	if wsPort == 0 {
		log.Fatalf("invalid websocket port")
		flag.PrintDefaults()
	}

	err := meta.Init(appName, localRegion, httpPort, wsPort, zkHosts)
	if err != nil {
		log.Fatal(err)
	}

	streams.StartAllStreams()

	sp := inspector.NewInspector(seedNodes)
	go sp.Run()

	c := controller.NewController()
	fe := frontend.NewFrontEnd(c, httpPort, wsPort)
	fe.Run()
}
