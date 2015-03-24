package meta

import (
	"fmt"
	"log"
	"net"
	"path"
	"strings"

	"launchpad.net/gozk"
)

const (
	PERM_DIRECTORY = zookeeper.PERM_ADMIN | zookeeper.PERM_CREATE | zookeeper.PERM_DELETE | zookeeper.PERM_READ | zookeeper.PERM_WRITE
	PERM_FILE      = zookeeper.PERM_ADMIN | zookeeper.PERM_READ | zookeeper.PERM_WRITE
)

func resolveIPv4Addr(addr string) (string, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", err
	}
	ipAddrs, err := net.LookupIP(host)
	for _, ipAddr := range ipAddrs {
		ipv4 := ipAddr.To4()
		if ipv4 != nil {
			return net.JoinHostPort(ipv4.String(), port), nil
		}
	}
	return "", fmt.Errorf("no IPv4addr for name %v", host)
}

func resolveZkAddr(zkAddr string) (string, error) {
	parts := strings.Split(zkAddr, ",")
	resolved := make([]string, 0, len(parts))
	for _, part := range parts {
		// The zookeeper client cannot handle IPv6 addresses before version 3.4.x.
		if r, err := resolveIPv4Addr(part); err != nil {
			log.Printf("cannot resolve %v, will not use it: %v", part, err)
		} else {
			resolved = append(resolved, r)
		}
	}
	if len(resolved) == 0 {
		return "", fmt.Errorf("no valid address found in %v", zkAddr)
	}
	return strings.Join(resolved, ","), nil
}

func DialZk(zkAddr string) (*zookeeper.Conn, <-chan zookeeper.Event, error) {
	resolvedZkAddr, err := resolveZkAddr(zkAddr)
	if err != nil {
		return nil, nil, err
	}

	zconn, session, err := zookeeper.Dial(resolvedZkAddr, 5e9)
	if err == nil {
		// Wait for connection, possibly forever
		event := <-session
		if event.State != zookeeper.STATE_CONNECTED {
			err = fmt.Errorf("zk connect failed: %v", event.State)
		}
		if err == nil {
			return zconn, session, nil
		} else {
			zconn.Close()
		}
	}
	return nil, nil, err
}

func CreateRecursive(zconn *zookeeper.Conn, zkPath, value string, flags int, aclv []zookeeper.ACL) (pathCreated string, err error) {
	pathCreated, err = zconn.Create(zkPath, value, flags, aclv)
	if zookeeper.IsError(err, zookeeper.ZNONODE) {
		dirAclv := make([]zookeeper.ACL, len(aclv))
		for i, acl := range aclv {
			dirAclv[i] = acl
			dirAclv[i].Perms = PERM_DIRECTORY
		}
		_, err = CreateRecursive(zconn, path.Dir(zkPath), "", flags, dirAclv)
		if err != nil && !zookeeper.IsError(err, zookeeper.ZNODEEXISTS) {
			return "", err
		}
		pathCreated, err = zconn.Create(zkPath, value, flags, aclv)
	}
	return
}
