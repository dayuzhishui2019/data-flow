package e_1400server

import (
	"dyzs/data-flow/util/times"
	"errors"
	"net"
	"time"
)

type SystemTime struct {
	VIIDServerID string  `json:"VIIDServerID"` //设备ID，该服务器标识符，中文统称设备ID

	TimeMode string  `json:"TimeMode"` //校时模式  1 -- 网络  2 -- 手动

	LocalTime string  `json:"LocalTime"` //日期时间 dateTime	dateTime	DE00554	格式：YYYYMMDDhhmmss

	TimeZone string `json:"TimeZone"`  //时区 北京属于东八区，UTC+8
}

func BuildSystemTime() *SystemTime{

	var ip string
	ipNet ,e := externalIP()
	if e == nil && ipNet != nil{
		ip = ipNet.String()
	}
	return &SystemTime{
		VIIDServerID: ip,
		TimeMode: "1",
		LocalTime:times.Time2StrF(time.Now(),"20060102150405"),
		TimeZone: "UTC+8",
	}
}

func externalIP() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			ip := getIpFromAddr(addr)
			if ip == nil {
				continue
			}
			return ip, nil
		}
	}
	return nil, errors.New("connected to the network?")
}

func getIpFromAddr(addr net.Addr) net.IP {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil || ip.IsLoopback() {
		return nil
	}
	ip = ip.To4()
	if ip == nil {
		return nil // not an ipv4 address
	}

	return ip
}