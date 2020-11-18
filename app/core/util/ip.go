package util

import (
	"fmt"
	"github.com/go-ping/ping"
	"net"
	"strings"
	"time"
)

func GetRecommendIp() []string {
	return GetIP([]string{"docker0"}, []string{"eth0"})
}

//获取本机IP
// IgnoredInterfaces 忽略的网卡
// PreferredNetworks 倾向使用的地址
func GetIP(IgnoredInterfaces, PreferredNetworks []string) []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return []string{"0.0.0.0"}
	}

	addresies := []string{}

FACE_LOOP:
	for _, face := range ifaces {
		//忽略的网络
		for _, ignored := range IgnoredInterfaces {
			if ignored == face.Name {
				continue FACE_LOOP
			}
		}
		//优先使用的网络
		preferred := false
		for _, v := range PreferredNetworks {
			if v == face.Name {
				preferred = true
				break
			}
		}

		if addresss, err := face.Addrs(); err == nil {
			for _, address := range addresss {
				if ipNet, ok := address.(*net.IPNet); ok &&
					!ipNet.IP.IsLoopback() &&
					!ipNet.IP.IsInterfaceLocalMulticast() &&
					!ipNet.IP.IsMulticast() &&
					!ipNet.IP.IsUnspecified() {
					if ipNet.IP.To4() != nil {
						if preferred {
							return []string{ipNet.IP.String()}
						} else {
							addresies = append(addresies, ipNet.IP.String())
						}
					}
				}
			}
		}
	}
	return addresies
}

func SockTo(host string, port uint32, timeout time.Duration) bool {
	c, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), timeout)
	if err == nil {
		_ = c.Close()
	}
	return err == nil
}

func Ping(host string, count int, interval, timeout time.Duration) bool {
	p, err := ping.NewPinger(host)
	if err != nil {
		return false
	}
	p.Timeout = timeout
	p.Interval = interval
	p.Count = count
	if err = p.Run(); err != nil {
		return false
	} // blocks until finished
	stats := p.Statistics() // get send/receive/rtt stats
	return stats.PacketsSent == stats.PacketsRecv
}

//是否和本地网络在同一网段
func IsSegment(a string) bool {
	localIps := GetRecommendIp()
	aseg := a[:strings.LastIndex(a, ".")]
	for _, b := range localIps {
		bseg := b[:strings.LastIndex(b, ".")]
		if aseg == bseg {
			return true
		}
	}
	return false
}
