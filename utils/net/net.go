package net

import (
	"errors"
	"net"
	"os"
	"strings"
)

func LocalIP() (string, error) {
	ifs, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, v := range ifs {
		addrs, _ := v.Addrs()
		if len(addrs) == 0 {
			continue
		}
		for _, addr := range addrs {
			v4 := addr.(*net.IPNet).IP.To4()
			if len(v4) != 4 {
				continue
			}
			if v4[0] != 127 {
				return addr.(*net.IPNet).IP.String(), nil
			}
		}
		return "127.0.0.1", nil
	}
	return "", errors.New("unknown-ip")
}

func Hostname() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return LocalIP()
	}

	hostname = strings.TrimSuffix(hostname, ".baidu.com")
	hostname = strings.TrimSuffix(hostname, ".baidu.com.")

	return hostname, nil
}
