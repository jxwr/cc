package command

import (
	cc "github.com/jxwr/cc/controller"
	"github.com/jxwr/cc/meta"
	"github.com/jxwr/cc/topo"
)

/// Merge seeds in the same region, reported by controllers in different zones.
type MergeSeedsCommand struct {
	Region string
	Seeds  []*topo.Node
}

func (self *MergeSeedsCommand) Execute(c *cc.Controller) (cc.Result, error) {
	if meta.LocalRegion() == self.Region {
		meta.MergeSeeds(self.Seeds)
	}
	return nil, nil
}
