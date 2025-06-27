package util

import (
	"fmt"
	"net"
)

func GetLocalIP() (string, error) {
	// 先尝试获取外网出口IP
	if ip, err := GetOutboundIP(); err == nil {
		return ip, nil
	}

	// 尝试常见网卡名称
	commonInterfaces := []string{"eth0", "en0", "enp0s3", "wlan0"}
	for _, iface := range commonInterfaces {
		if ip, err := GetInterfaceIP(iface); err == nil {
			return ip, nil
		}
	}

	// 最后尝试所有网卡
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipNet.IP
			if ip.To4() != nil && !ip.IsLoopback() && ip.IsGlobalUnicast() {
				return ip.String(), nil
			}
		}
	}

	return "", fmt.Errorf("没有找到有效的IP地址")
}
func GetOutboundIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func GetInterfaceIP(interfaceName string) (string, error) {
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return "", fmt.Errorf("找不到网卡 %s: %v", interfaceName, err)
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", fmt.Errorf("获取网卡 %s 地址失败: %v", interfaceName, err)
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}

		ip := ipNet.IP
		// 返回第一个 IPv4 地址
		if ip.To4() != nil && !ip.IsLoopback() {
			return ip.String(), nil
		}
	}

	return "", fmt.Errorf("网卡 %s 没有有效的 IPv4 地址", interfaceName)
}
