package context

import (
	"fmt"

	"github.com/ksarch-saas/cc/frontend/auth"
	"github.com/ksarch-saas/cc/meta"
	zookeeper "github.com/samuel/go-zookeeper/zk"
)

func AddUser(userName, role string) (string, error) {
	zconn, _, err := meta.DialZk(ZkAddr)
	defer func() {
		if zconn != nil {
			zconn.Close()
		}
	}()
	if err != nil {
		return "", fmt.Errorf("zk: can't connect: %v", err)
	}
	zkPath := "/r3/users/" + userName
	exists, _, err := zconn.Exists(zkPath)
	if err != nil {
		return "", fmt.Errorf("zk: call exist failed %v", err)
	}
	if exists {
		return "", fmt.Errorf("zk: %s node already exists", userName)
	} else {
		//add node
		token := auth.GenerateToken(userName)
		_, err := zconn.Create(zkPath, []byte(token), 0, zookeeper.WorldACL(zookeeper.PermAll))
		if err != nil {
			return "", fmt.Errorf("zk: create failed %v", err)
		}
		return token, nil
	}
}

func ModUser(userName, role string, config []byte, version int32) error {
	zconn, _, err := meta.DialZk(ZkAddr)
	defer func() {
		if zconn != nil {
			zconn.Close()
		}
	}()
	if err != nil {
		return fmt.Errorf("zk: can't connect: %v", err)
	}
	zkPath := "/r3/users/" + userName
	exists, _, err := zconn.Exists(zkPath)
	if err != nil {
		return fmt.Errorf("zk: call exist failed %v", err)
	}
	if !exists {
		return fmt.Errorf("zk: %s node not exists", userName)
	} else {
		//update node
		_, err := zconn.Set(zkPath, config, version)
		if err != nil {
			return fmt.Errorf("zk: set failed %v", err)
		}
		return nil
	}
}

func GetUser(userName string) (string, int32, error) {
	zconn, _, err := meta.DialZk(ZkAddr)
	defer func() {
		if zconn != nil {
			zconn.Close()
		}
	}()
	if err != nil {
		return "", 0, fmt.Errorf("zk: can't connect: %v", err)
	}
	zkPath := "/r3/users/" + userName
	token, stat, err := zconn.Get(zkPath)
	if err != nil {
		return "", 0, fmt.Errorf("zk: get: %v", err)
	}
	return string(token), stat.Version, nil
}

func DelUser(userName string, version int32) error {
	zconn, _, err := meta.DialZk(ZkAddr)
	defer func() {
		if zconn != nil {
			zconn.Close()
		}
	}()
	if err != nil {
		return fmt.Errorf("zk: can't connect: %v", err)
	}
	zkPath := "/r3/users/" + userName
	err = zconn.Delete(zkPath, version)
	if err != nil {
		return fmt.Errorf("zk: path delete %v", err)
	}
	return nil
}

func CheckSuperPerm(userName string) (bool, error) {
	zconn, _, err := meta.DialZk(ZkAddr)
	defer func() {
		if zconn != nil {
			zconn.Close()
		}
	}()
	if err != nil {
		return false, fmt.Errorf("zk: can't connect: %v", err)
	}
	zkPath := "/r3/users/" + userName + "/super"
	exists, _, err := zconn.Exists(zkPath)
	if err != nil {
		return false, err
	}
	return exists, nil
}
