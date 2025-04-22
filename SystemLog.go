package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	// 检查是否以root权限运行
	if os.Geteuid() != 0 {
		fmt.Println("警告：部分日志需要root权限访问，建议使用sudo运行本程序")
	}

	// 定义常见日志源
	logSources := []struct {
		Name string
		Path string
	}{
		{"System", "/var/log/syslog"},              // Ubuntu系系统日志
		{"System", "/var/log/messages"},            // RedHat系系统日志
		{"Authentication", "/var/log/auth.log"},    // 认证日志
		{"APT", "/var/log/apt/history.log"},        // APT包管理器日志
		{"YUM", "/var/log/yum.log"},                // YUM包管理器日志
		{"Nginx", "/var/log/nginx/error.log"},     // Nginx错误日志
		{"MySQL", "/var/log/mysql/error.log"},     // MySQL错误日志
		{"Kernel", "/var/log/kern.log"},            // 内核日志
		{"Docker", "/var/log/docker.log"},          // Docker日志
		{"Custom", "/var/log/myapp.log"},           // 示例自定义应用日志
	}

	for _, source := range logSources {
		printSectionHeader(source.Name, source.Path)
		content, err := readLogFile(source.Path)
		if err != nil {
			handleReadError(err, source.Path)
			continue
		}
		fmt.Println(content)
	}
}

func printSectionHeader(name, path string) {
	header := fmt.Sprintf("===== %s (%s) =====", name, path)
	fmt.Println("\n" + header)
	fmt.Println(strings.Repeat("-", len(header)))
}

func readLogFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func handleReadError(err error, path string) {
	switch {
	case os.IsNotExist(err):
		fmt.Printf("文件不存在: %s\n", path)
	case os.IsPermission(err):
		fmt.Printf("权限不足: %s (请使用sudo运行)\n", path)
	default:
		fmt.Printf("读取错误[%s]: %v\n", path, err)
	}
}