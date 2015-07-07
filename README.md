
Alpha.

### Features

Conbined with `https://github.com/ksarch-saas/redis` and `https://github.com/ksarch-saas/r3proxy`:

* Multi-IDC Support, every node has a tag, never elect node in a backup region as master.
* Replication via siblings, do not resync from new master if it was our sibling.
* Builtin web console and command line tools for realtime monitor and control.
* Configurable read preferences, you can read from primary, primary_preferred or neareat region.
* Global failover constraint, if many many clusters deploy on the same machine pool, two clusters will never exec failover jobs at the same time.
* Slot rebalance.
