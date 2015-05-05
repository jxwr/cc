package meta

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jxwr/cc/topo"
	"launchpad.net/gozk"
)

const (
	DEFAULT_AUTOFAILOVER_INTERVAL  time.Duration = 5 * time.Minute // 5min
	DEFAULT_MIGRATE_KEYS_EACH_TIME               = 100
	DEFAULT_MIGRATE_TIMEOUT                      = 2000
)

type AppConfig struct {
	AppName               string
	AutoEnableSlaveRead   bool
	AutoEnableMasterWrite bool
	AutoFailover          bool
	AutoFailoverInterval  time.Duration
	MasterRegion          string
	Regions               []string
	MigrateKeysEachTime   int
	MigrateTimeout        int
}

type ControllerConfig struct {
	Ip       string
	HttpPort int
	WsPort   int
	Region   string
}

type FailoverRecord struct {
	AppName   string
	NodeId    string
	NodeAddr  string
	Timestamp time.Time
	Region    string
	Tag       string
	Role      string
	Ranges    []topo.Range
}

func (m *Meta) handleAppConfigChanged(watch <-chan zookeeper.Event) {
	for {
		event := <-watch
		if event.Type == zookeeper.EVENT_CHANGED {
			a, w, err := m.FetchAppConfig()
			if err == nil {
				if a.MigrateKeysEachTime == 0 {
					a.MigrateKeysEachTime = DEFAULT_MIGRATE_KEYS_EACH_TIME
				}
				if a.MigrateTimeout == 0 {
					a.MigrateTimeout = DEFAULT_MIGRATE_TIMEOUT
				}
				if a.AutoFailoverInterval == 0 {
					a.AutoFailoverInterval = DEFAULT_AUTOFAILOVER_INTERVAL
				}
				mutex.Lock()
				m.appConfig = a
				mutex.Unlock()
				log.Println("meta: app config changed.", a)
			} else {
				log.Printf("meta: fetch app config failed, %v", err)
			}
			watch = w
		} else {
			log.Printf("meta: unexpected event coming, %v", event)
			break
		}
	}
}

func (m *Meta) FetchAppConfig() (*AppConfig, <-chan zookeeper.Event, error) {
	zconn := m.zconn
	appName := m.appName
	data, _, watch, err := zconn.GetW("/r3/app/" + appName)
	if err != nil {
		return nil, watch, err
	}
	var c AppConfig
	err = json.Unmarshal([]byte(data), &c)
	if err != nil {
		return nil, watch, fmt.Errorf("meta: parse app config error, %v", err)
	}
	if c.AppName != appName {
		return nil, watch, fmt.Errorf("meta: local appname is different from zk, %s <-> %s", appName, c.AppName)
	}
	if c.MasterRegion == "" {
		return nil, watch, fmt.Errorf("meta: master region not set")
	}
	if len(c.Regions) == 0 {
		return nil, watch, fmt.Errorf("meta: regions empty")
	}
	if c.MigrateKeysEachTime == 0 {
		c.MigrateKeysEachTime = DEFAULT_MIGRATE_KEYS_EACH_TIME
	}
	if c.MigrateTimeout == 0 {
		c.MigrateTimeout = DEFAULT_MIGRATE_TIMEOUT
	}
	if c.AutoFailoverInterval == 0 {
		c.AutoFailoverInterval = DEFAULT_AUTOFAILOVER_INTERVAL
	}
	return &c, watch, nil
}

func (m *Meta) RegisterLocalController() error {
	zconn := m.zconn
	zkPath := fmt.Sprintf(m.ccDirPath + "/cc_" + m.localRegion + "_")
	conf := &ControllerConfig{
		Ip:       m.localIp,
		HttpPort: m.httpPort,
		Region:   m.localRegion,
		WsPort:   m.wsPort,
	}
	data, err := json.Marshal(conf)
	if err != nil {
		return err
	}
	path, err := zconn.Create(zkPath, string(data), zookeeper.SEQUENCE|zookeeper.EPHEMERAL, zookeeper.WorldACL(PERM_FILE))
	if err == nil {
		xs := strings.Split(path, "/")
		m.selfZNodeName = xs[len(xs)-1]
	}
	return err
}

func (m *Meta) FetchControllerConfig(zkNode string) (*ControllerConfig, <-chan zookeeper.Event, error) {
	data, _, watch, err := m.zconn.GetW(m.ccDirPath + "/" + zkNode)
	if err != nil {
		return nil, watch, err
	}
	var c ControllerConfig
	err = json.Unmarshal([]byte(data), &c)
	if err != nil {
		return nil, watch, err
	}
	return &c, watch, nil
}

func (m *Meta) IsDoingFailover() (bool, error) {
	stat, err := m.zconn.Exists("/r3/failover/doing")
	if err == nil {
		if stat != nil {
			return true, nil
		} else {
			return false, nil
		}
	} else {
		return true, err
	}
}

func (m *Meta) MarkFailoverDoing(record *FailoverRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	path, err := m.zconn.Create("/r3/failover/doing", string(data),
		zookeeper.EPHEMERAL, zookeeper.WorldACL(PERM_FILE))
	if err != nil {
		return err
	}
	log.Printf("meta: mark doing failover at %s", path)
	return nil
}

func (m *Meta) UnmarkFailoverDoing() error {
	err := m.zconn.Delete("/r3/failover/doing", -1)
	if err != nil {
		return err
	}
	log.Printf("meta: unmark doing failover")
	return nil
}

func (m *Meta) LastFailoverRecord() (*FailoverRecord, error) {
	children, stat, err := m.zconn.Children("/r3/failover/history")
	if err != nil {
		return nil, err
	}
	if stat.NumChildren() == 0 {
		return nil, nil
	}

	last := children[len(children)-1]
	data, _, err := m.zconn.Get("/r3/failover/history/" + last)
	if err != nil {
		return nil, err
	}

	var record FailoverRecord
	err = json.Unmarshal([]byte(data), &record)
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (m *Meta) AddFailoverRecord(record *FailoverRecord) error {
	zkPath := fmt.Sprintf("/r3/failover/history/record_%s_%s", record.AppName, record.Region)
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	path, err := m.zconn.Create(zkPath, string(data), zookeeper.SEQUENCE, zookeeper.WorldACL(PERM_FILE))
	log.Printf("meta: failover record created at %s", path)
	return nil
}
