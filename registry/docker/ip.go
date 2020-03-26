package docker

import "net"

func GetRecommendIp() string {
	return GetIP([]string{"docker0"}, []string{"eth0"})
}

//获取本机IP
// IgnoredInterfaces 忽略的网卡
// PreferredNetworks 倾向使用的地址
func GetIP(IgnoredInterfaces, PreferredNetworks []string) string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "0.0.0.0"
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
							return ipNet.IP.String()
						} else {
							addresies = append(addresies, ipNet.IP.String())
						}
					}
				}
			}
		}
	}
	if len(addresies) != 0 {
		return addresies[0]
	}
	return "0.0.0.0"
}
