package main

import (
	"fmt"
	"log"
	"loglog/varlog"
)

func main() {
	// 调用解析函数并获取解析后的消息
	//jsonData, err := varlog.ParseSyslogMessages()
	//jsonData, err := varlog.ParseAlternativesLogMessages()
	//jsonData, err := varlog.ParseAuthLogMessages()
	//jsonData, err := varlog.ParseDpkgLogMessages()
	//jsonData, err := varlog.ParseFontconfigLogMessages()
	//jsonData, err := varlog.ParseKernelLogMessages()
	//jsonData, err := varlog.ParseXorgLogMessages()
	//jsonData, err := varlog.ParseDmesgLogMessages()
	//jsonData, err := varlog.ParseFaillogEntries()
	//jsonData, err := varlog.ParseLastlogLogMessages()
	jsonData, err := varlog.ParseWtmpLogMessages()

	if err != nil {
		log.Fatalf("解析日志消息失败: %v", err)
	}

	// 打印解析后的消息
	for _, msg := range jsonData {
		fmt.Println(msg)
	}
}
