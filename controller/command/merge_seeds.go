package command

import (
	cc "github.com/ksarch-saas/cc/controller"
	"github.com/ksarch-saas/cc/meta"
	"github.com/ksarch-saas/cc/topo"
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
