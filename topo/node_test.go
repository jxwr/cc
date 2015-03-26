package topo

import (
	"fmt"
	"testing"
)

func TestRangesSplitN(t *testing.T) {
	n0 := Node{Ranges: []Range{
		Range{0, 100},
		Range{200, 300},
		Range{400, 500},
	}}
	fmt.Println(n0.RangesSplitN(3))

	n1 := Node{Ranges: []Range{
		Range{0, 100},
		Range{200, 300},
	}}
	fmt.Println(n1.RangesSplitN(3))

	n11 := Node{Ranges: []Range{
		Range{0, 100},
		Range{200, 300},
		Range{400, 500},
	}}
	fmt.Println(n11.RangesSplitN(2))

	n2 := Node{Ranges: []Range{
		Range{0, 100},
		Range{200, 300},
		Range{400, 500},
		Range{600, 700},
		Range{800, 800},
	}}
	fmt.Println(n2.RangesSplitN(3))
}
