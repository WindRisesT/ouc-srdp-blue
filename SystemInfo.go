//记得先安装go get github.com/shirou/gopsutil/v3
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// 获取CPU硬件信息
func GetCPUInfo() {
	data, _ := os.ReadFile("/proc/cpuinfo")
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.Contains(line, "model name") {
			parts := strings.Split(line, ":")
			fmt.Printf("CPU型号: \t%s\n", strings.TrimSpace(parts[1]))
			break
		}
	}

	if physicalCnt, err := cpu.Counts(false); err == nil {
		fmt.Printf("物理核心数: \t%d\n", physicalCnt)
	}

	if logicalCnt, err := cpu.Counts(true); err == nil {
		fmt.Printf("逻辑核心数: \t%d\n", logicalCnt)
	}
}

// 获取CPU使用率
func GetCPUUsage() {
	percent, _ := cpu.Percent(1*time.Second, false)
	fmt.Printf("CPU使用率: \t%.2f%%\n", percent[0])
}

// 获取内存信息
func GetMemoryInfo() {
	if v, err := mem.VirtualMemory(); err == nil {
		fmt.Printf("总内存: \t%.2f GB\n", float64(v.Total)/1024/1024/1024)
		fmt.Printf("可用内存: \t%.2f GB\n", float64(v.Available)/1024/1024/1024)
		fmt.Printf("内存使用率: \t%.2f%%\n", v.UsedPercent)
	}
}

// 获取磁盘信息
func GetDiskInfo() {
	if parts, err := disk.Partitions(true); err == nil {
		for _, part := range parts {
			fmt.Printf("分区: \t\t%s\n", part.Device)
			fmt.Printf("挂载点: \t%s\n", part.Mountpoint)
			if usage, err := disk.Usage(part.Mountpoint); err == nil {
				fmt.Printf("总空间: \t%.2f GB\n", float64(usage.Total)/1024/1024/1024)
				fmt.Printf("使用率: \t%.2f%%\n", usage.UsedPercent)
			}
			fmt.Println("----------------------------------------")
		}
	}
}

// 获取网络信息
func GetNetworkInfo() {
	if interfaces, err := net.Interfaces(); err == nil {
		for _, iface := range interfaces {
			fmt.Printf("网卡名称: \t%s\n", iface.Name)
			fmt.Printf("MAC地址: \t%s\n", iface.HardwareAddr)
			for _, addr := range iface.Addrs {
				fmt.Printf("IP地址: \t%s\n", addr.Addr)
			}
			fmt.Println("----------------------------------------")
		}
	}
}

// 获取系统信息
func GetSystemInfo() {
	if hostInfo, err := host.Info(); err == nil {
		fmt.Printf("主机名: \t%s\n", hostInfo.Hostname)
		fmt.Printf("操作系统: \t%s\n", hostInfo.Platform)
		fmt.Printf("内核版本: \t%s\n", hostInfo.KernelVersion)
		fmt.Printf("系统启动时间: \t%s\n", time.Unix(int64(hostInfo.BootTime), 0))
	}

	if osRelease, err := ioutil.ReadFile("/etc/os-release"); err == nil {
		lines := strings.Split(string(osRelease), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				fmt.Println("系统版本: \t" + strings.Trim(line[12:], "\""))
				break
			}
		}
	}
}

func main() {
	fmt.Println("============== 系统信息 ==============")
	GetSystemInfo()
	
	fmt.Println("\n============== CPU信息 ==============")
	GetCPUInfo()
	GetCPUUsage()
	
	fmt.Println("\n============== 内存信息 ==============")
	GetMemoryInfo()
	
	fmt.Println("\n============== 磁盘信息 ==============")
	GetDiskInfo()
	
	fmt.Println("\n============== 网络信息 ==============")
	GetNetworkInfo()
}