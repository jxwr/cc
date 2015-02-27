package redis

import (
	"fmt"
	"testing"
)

func TestIsAlvie(t *testing.T) {
	fmt.Println(IsAlive("127.0.0.1:7000"))
}

func TestClusterNodes(t *testing.T) {
	fmt.Println(ClusterNodes("127.0.0.1:7000"))
}
