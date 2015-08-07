package initialize

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/fatih/color"
	"os"
	"strings"
)

var (
	flags = []cli.Flag{
		cli.StringFlag{"s,service", "", "the service to initialize"},
		cli.StringFlag{"l,logic", "", "logic machine rooms list"},
		cli.StringFlag{"m,master", "", "master machine rooms list"},
		cli.IntFlag{"r,replicas", 0, "replicaset of each master node"},
		cli.BoolFlag{"force", "reset nodes before init cluster force"},
	}

	Command = cli.Command{
		Name:   "init",
		Usage:  "initialize a new empty cluster",
		Action: action,
		Flags:  flags,
	}
)

func action(c *cli.Context) {
	fmt.Println(c.Args())
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	var cmd string

	s := c.String("s")
	if s == "" {
		fmt.Println(red("-s service must be assigned"))
		os.Exit(-1)
	}
	l := c.String("l")
	if l == "" {
		fmt.Println(red("-l logic machine room must be assigned"))
		os.Exit(-1)
	}
	m := c.String("m")
	if m == "" {
		fmt.Println(red("-m master machine rooms must be assigned"))
		os.Exit(-1)
	}
	force := c.Bool("force")
	replicas := c.Int("r")

	allnodes := []*Node{}
	masterRooms := strings.Split(m, ",")
	rooms := strings.Split(l, ",")
	for _, room := range rooms {
		service_name := fmt.Sprintf("%s.osp.%s", s, room)
		nodes, err := getNodes(service_name)
		if err == nil {
			for _, n := range nodes {
				/* set logic mr */
				n.LogicMR = room
			}
		}
		allnodes = append(allnodes, nodes...)
	}
	/* done get allnodes */
	if replicas != 0 && len(allnodes)%(replicas+1) != 0 {
		fmt.Printf("%s. Not enough nodes\n", red("ERROR"))
		os.Exit(-1)
	}

	/* reset all nodes */
	if force {
		fmt.Printf("Type %s to continue: ", green("yes"))
		fmt.Printf("%s\n", red("(--force will reset the cluster)"))

		fmt.Scanf("%s\n", &cmd)
		if cmd != "yes" {
			os.Exit(0)
		}
		resp, err := resetNodes(allnodes)
		if err != nil {
			fmt.Println(resp, err)
		}
	}

	/* check nodes state */
	for _, node := range allnodes {
		node.Alive = isAlive(node)
		fmt.Printf("connecting to %s\t%s\t", node.Ip, node.Port)
		if node.Alive {
			fmt.Printf("%s\n", green("OK"))
		} else {
			fmt.Printf("%s\n", red("FAILED"))
		}
	}

	/* check and set state */
	for _, node := range allnodes {
		checkAndSetState(node)
	}

	/* validate the state and continue */
	if validateProcess(allnodes) == false {
		fmt.Printf("Not all nodes have the right status  %s\n", red("Error"))
		os.Exit(-1)
	}

	/* build cluster */
	masterNodes, err := buildCluster(allnodes, replicas, masterRooms, rooms)
	if err != nil {
		fmt.Printf("%v buildCluster failed\n", red(err))
		os.Exit(-1)
	}

	err = assignSlots(masterNodes)
	if err != nil {
		fmt.Printf("%s assignSlot failed\n", red("Error"))
		os.Exit(-1)
	}

	/* assignment summary */
	for _, node := range masterNodes {
		fmt.Printf("%s %s\t%s\t%s\t%s\n", yellow("M:"), node.Id, node.Ip, node.Port, yellow(node.SlotsRange))
		slaves := getSlaves(allnodes, node)
		for _, slave := range slaves {
			fmt.Printf("%s %s\t%s\t%s\t%s\n", cyan("S:"), slave.Id, slave.Ip, slave.Port, slave.MasterId)
		}
	}
	fmt.Printf("Type %s to continue: ", green("yes"))
	fmt.Scanf("%s\n", &cmd)
	if cmd != "yes" {
		os.Exit(0)
	}

	/* send cmd to cluster */
	meetEach(allnodes)

	for _, node := range masterNodes {
		fmt.Printf("Node:%s\n", node.Id)
		fmt.Printf("%-40s", "setting slots...")
		resp, err := addSlotRange(node)
		if err != nil {
			fmt.Println(red(resp))
			break
		} else {
			fmt.Println(green(resp))
		}
		resp, err = rwMasterState(node)
		if err != nil {
			fmt.Printf("%s\n", red("FAILED to chmod, please check"))
		}
		slaves := getSlaves(allnodes, node)
		fmt.Printf("%-40s", "setting replicas...")
		err = rwReplicasState(slaves)
		if err != nil {
			fmt.Printf("%s\n", red("FAILED to chmod, please check"))
		}
		resp, err = setReplicas(slaves)
		if err != nil {
			fmt.Printf("%s\n", red(resp))
			break
		} else {
			fmt.Printf("%s\n", green(resp))
		}
	}

	/* cluster info check */
	if checkClusterInfo(allnodes) {
		fmt.Printf("%s. All node aggree the configure\n", green("OK"))
	} else {
		fmt.Printf("%s. Node configure inconsistent or slots incomplete\n", red("Error"))
	}
}
