package varlog

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/hpcloud/tail"
)

func readOnlyBuf(fileString string) { // 开缓冲按行读取文件内容！！！！！！！
	file, err := os.OpenFile(fileString, os.O_RDONLY, 0666)
	if err == nil {
		fmt.Println("文件打开成功")
	} else {
		fmt.Println("文件打开失败，err=", err)
		return
	}
	defer func() {
		err := file.Close()
		if err != nil {
			return
		}
		fmt.Println("文件已关闭")
	}()

	reader := bufio.NewReader(file)

	for {
		str, err := reader.ReadString('\n')
		if err == nil {
			fmt.Println((str))
		} else {
			if err == io.EOF {
				fmt.Println("已到文件末尾")
				break
			} else {
				fmt.Println("读取失败，err", err)
				return
			}
		}
	}
	fmt.Println("文件读取完毕!")
}

// 定义一个消息类型，用于在通道中传递

func tailFile(fileName string, msgChan chan<- string) { // 实现tail功能实时按行读取文件内容，并发送到管道！！！！！！！
	config := tail.Config{
		ReOpen:    true,                                 // 重新打开
		Follow:    true,                                 // 当有新内容时，会持续读取这些新内容
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2}, // Offset: 0 表示从文件的开始位置读取。Whence: 0-开头，1-当前位置，2-结尾
		MustExist: false,                                // 文件不存在不报错（例如，新启动的服务可能还没有开始写入日志）
		Poll:      true,                                 //轮询来检查文件变化。可能比依赖文件系统的通知更可靠，但它可能会消耗更多的CPU资源
	}
	tails, err := tail.TailFile(fileName, config)
	if err != nil {
		fmt.Println("tail file failed, err:", err)
		return
	}

	var (
		line *tail.Line
		ok   bool
	)
	for {
		line, ok = <-tails.Lines //遍历chan，读取日志内容
		if !ok {
			fmt.Printf("tail file close reopen, filename:%s\n", tails.Filename)
			time.Sleep(time.Second)
			continue
		}
		fmt.Println("line:", line.Text)
		msgChan <- line.Text // 发送字符串到通道
	}
}

func cmdFile(cmdString string, arg ...string) string { // 执行cmd命令，并返回输出！！！！！！！
	// 如果调用者没有指定目录，则使用默认目录
	dirString := "/var/log"
	if len(arg) > 0 && arg[len(arg)-1] == "" {
		// 如果arg的最后一个元素是空字符串，我们将其视为目录的占位符
		dirString = arg[len(arg)-2]
		//		fmt.Println(dirString)
		// 移除最后一个元素（目录占位符）
		arg = arg[:len(arg)-2]
		//		fmt.Println(arg)
	}

	path, err0 := exec.LookPath(cmdString) //查找命令位置
	if err0 != nil {
		log.Printf("'%s' not found", cmdString)
	} else {
		log.Printf("'%s' is in '%s'\n", cmdString, path)
	}

	cmd := exec.Command(cmdString, arg...) //执行命令

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} //父进程退出时kill子进程与孙进程
	cmd.Dir = dirString                                   //执行所在位置

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run() //运行命令

	if err != nil {
		log.Fatalf("failed to call cmd.Run(): %v", err)
	}
	log.Printf("\nout:\n%s\nerr:\n%s\n", stdout.String(), stderr.String()) //输出

	log.Printf("pid: %d", cmd.Process.Pid)                  //pid号
	log.Printf("exitcode: %d", cmd.ProcessState.ExitCode()) //退出码
	return stdout.String()
}

/*
func IdontKnow() { // 测试tailFile函数与cmdFile函数！！！！！！！
	// 创建一个带缓冲的通道
	msgChan := make(chan string, 10)

	// 启动死循环函数作为goroutine
	go tailFile("/var/log/auth.log", msgChan)

	// 定义一个等待组来等待所有goroutine完成（在这个例子中其实不需要，因为我们不会等待死循环结束）
	var wg sync.WaitGroup

	test := cmdFile("lastlog")
	fmt.Println(test)

	// 启动一个goroutine来接收通道中的消息，并打印它们
	wg.Add(1)

	message := ""
	go func() {
		defer wg.Done()
		for msg := range msgChan {
			fmt.Println("Collected:", msg)
			message += msg + "\n"
		}
	}()

	fmt.Printf("Press any key to exit...") //按任意键退出
	b := make([]byte, 1)
	os.Stdin.Read(b)

	fmt.Println(message)

	close(msgChan) // 关闭通道，这样接收方会知道没有更多的消息会发送过来
	// 等待接收goroutine完成（实际上，由于我们关闭了通道，它会在处理完所有消息后自然退出）
	wg.Wait()
	fmt.Println("Main function finished collecting messages.")
}
*/

func sendAllFiles() { // 打印所有日志文件内容！！！！！！！
	// 文件名列表，这里你可以从命令行参数、配置文件或其他地方获取这个列表
	files := []string{"/var/log/alternatives.log", "/var/log/auth.log", "/var/log/dmesg", "/var/log/dpkg.log", "/var/log/fontconfig.log", "/var/log/kern.log", "/var/log/syslog", "Xorg.0.log"}

	// 遍历文件列表并调用 processFile 函数处理每个文件
	for _, file := range files {
		readOnlyBuf(file)
	}
	cmdFile("faillog", "-a")
	cmdFile("lastlog")
	cmdFile("last", "-f", "wtmp")
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////

func readLogMessages(logPath string) ([]string, error) { // 读取log文件所有内容，返回字符串切片！！！！！！！！
	file, err := os.Open(logPath)
	if err != nil {
		return nil, fmt.Errorf("无法打开log文件: %w", err)
	}
	defer file.Close()

	// 创建一个读取器
	reader := bufio.NewReader(file)

	// 用于存储消息的切片
	var messages []string

	// 逐行读取syslog消息
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// 文件结束，退出循环
				break
			}
			return nil, fmt.Errorf("读取log文件时出错: %w", err)
		}

		// 去除消息末尾的换行符，并添加到切片中
		message := strings.TrimSuffix(line, "\n")
		messages = append(messages, message)
	}

	return messages, nil
}

func cmdLogFile(cmdString string, arg ...string) ([]string, error) { // cmd命令读取文件内容，返回字符串切片！！！！！！！！
	// 如果调用者没有指定目录，则使用默认目录
	dirString := "/var/log"

	// 检查最后一个参数是否为空字符串，如果是，则将其视为目录占位符
	if len(arg) > 0 && arg[len(arg)-1] == "" {
		dirString = arg[len(arg)-2]
		arg = arg[:len(arg)-2] // 移除目录占位符和空字符串
	}

	cmd := exec.Command(cmdString, arg...)                // 执行命令
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // 父进程退出时kill子进程与孙进程
	cmd.Dir = dirString                                   // 执行所在位置

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run() // 运行命令
	if err != nil {
		return nil, fmt.Errorf("failed to call cmd.Run(): %w, stderr: %s", err, stderr.String())
	}

	// 将标准输出分割为行，并返回字符串切片
	lines := strings.Split(stdout.String(), "\n")
	return lines, nil
}

func allLogFile() { // 返回所有日志文件切片！！！！！！！
	// 文件名列表，这里你可以从命令行参数、配置文件或其他地方获取这个列表
	files := []string{"/var/log/alternatives.log", "/var/log/auth.log", "/var/log/dmesg", "/var/log/dpkg.log", "/var/log/fontconfig.log", "/var/log/kern.log", "/var/log/syslog", "Xorg.0.log"}

	// 遍历文件列表并调用 processFile 函数处理每个文件
	for _, file := range files {
		readLogMessages(file)
	}
	cmdLogFile("dmesg")
	cmdLogFile("faillog", "-a")
	cmdLogFile("lastlog")
	cmdLogFile("last", "-f", "wtmp")
}

/*
func test() { // 切片内容取出测试！！！！！！！

	// 示例使用
	lines, err := cmdLogFile("faillog", "-a") // 假设最后一个参数是空字符串，用于指定目录
	if err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
	for _, line := range lines {
		fmt.Println(line) // 处理或打印每一行输出
	}

	// 定义syslog文件的路径
	syslogPath := "/var/log/syslog"

	// 调用函数并获取所有消息
	messages, err := readLogMessages(syslogPath)
	if err != nil {
		fmt.Printf("读取syslog消息失败: %v\n", err)
		return
	}

	// 打印所有消息或进行其他处理
	for _, msg := range messages {
		fmt.Println(msg)
	}
}
*/

///////////////////////////////////////////////////////////////////////////////////////////////////////////////

// // 在这个系统里用不了(syslog文件本身就没有PRI字段)
// func ParseSyslogMessage0(syslogMessages []string) {
// 	for _, msg := range syslogMessages {
// 		// 将消息转换为字节切片
// 		buff := []byte(msg)

// 		// 尝试检测 RFC 格式
// 		rfc, err := syslogParser.DetectRFC(buff)
// 		if err != nil {
// 			panic(err)
// 		}

// 		// 根据 RFC 格式选择解析器
// 		switch rfc {
// 		case syslogParser.RFC_3164:
// 			p := rfc3164.NewParser(buff)
// 			err := p.Parse()
// 			if err != nil {
// 				panic(err)
// 			}
// 			fmt.Println("RFC 3164 Format:")
// 			for k, v := range p.Dump() {
// 				fmt.Println(k, ":", v)
// 			}
// 		case syslogParser.RFC_5424:
// 			p := rfc5424.NewParser(buff)
// 			err := p.Parse()
// 			if err != nil {
// 				panic(err)
// 			}
// 			fmt.Println("RFC 5424 Format:")
// 			for k, v := range p.Dump() {
// 				fmt.Println(k, ":", v)
// 			}
// 		default:
// 			fmt.Println("Unknown RFC format")
// 		}
// 		fmt.Println("-----")
// 	}
// }

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type SyslogMessage struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Hostname  string `json:"hostname"`
	Appname   string `json:"appname"`
	Message   string `json:"message"`
}

func ParseSyslogMessages() ([]string, error) {
	var ParsedMsgs []*SyslogMessage

	// 定义syslog文件的路径
	syslogPath := "/var/log/syslog"

	// 调用函数并获取所有消息
	syslogMessages, err := readLogMessages(syslogPath)
	if err != nil {
		fmt.Printf("读取syslog消息失败: %v\n", err)
		return nil, fmt.Errorf("Error")
	}

	//// 示例 syslog 消息字符串切片
	//syslogMessages := []string{
	//	"Apr 27 18:50:38 k-k1234-pc org.kde.KScreen[2111]:  #011Rotation: 1",
	//	"Apr 27 18:50:38 k-k1234-pc org.kde.KScreen[2111]: kscreen.xrandr: #011Result:  0",
	//	// 添加更多 Syslog 消息...
	//}

	// 定义用于匹配 syslog 消息的正则表达式
	re := regexp.MustCompile(`^(\w+\s+\d+\s\d+:\d+:\d+)\s(\S+)\s(\S+):(.+)$`)

	// 解析 syslog 消息
	for _, msg := range syslogMessages {
		// 使用正则表达式查找匹配的组
		match := re.FindStringSubmatch(msg)
		if len(match) != 5 {
			//return nil, fmt.Errorf("Invalid syslog message format")
			continue
		}

		// 提取匹配的组并创建 SyslogMessage 对象
		syslogMsg := &SyslogMessage{
			Type:      "syslog_log",
			Timestamp: match[1],
			Hostname:  match[2],
			Appname:   match[3],
			Message:   strings.TrimSpace(match[4]),
		}

		ParsedMsgs = append(ParsedMsgs, syslogMsg)
	}

	// 将解析后的消息列表转换为 JSON 格式的字符串切片
	var jsonData []string
	for _, msg := range ParsedMsgs {
		data, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}
		jsonData = append(jsonData, string(data))
	}

	return jsonData, nil
}

// AlternativesLogMessage 结构体用于表示 alternatives.log 中的日志消息
type AlternativesLogMessage struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

// ParseAlternativesLogMessages 函数用于解析 alternatives.log 中的日志消息并返回 JSON 格式的字符串切片
func ParseAlternativesLogMessages() ([]string, error) {
	var ParsedMsgs []*AlternativesLogMessage

	// 定义 alternatives.log 文件的路径
	alternativesLogPath := "/var/log/alternatives.log"

	// 调用函数并获取所有消息
	alternativesLogMessages, err := readLogMessages(alternativesLogPath)
	if err != nil {
		fmt.Printf("读取 alternatives.log 消息失败: %v\n", err)
		return nil, err
	}

	// 定义用于匹配 alternatives.log 消息的正则表达式
	re := regexp.MustCompile(`^update-alternatives\s+(\d{4}-\d{2}-\d{2}\s\d{2}:\d{2}:\d{2}):\s(.+)$`)

	// 解析 alternatives.log 消息
	for _, msg := range alternativesLogMessages {
		// 使用正则表达式查找匹配的组
		match := re.FindStringSubmatch(msg)
		if len(match) != 3 {
			return nil, fmt.Errorf("Invalid message format")
			//continue // 不是有效的 alternatives.log 消息，跳过
		}

		// 提取匹配的组并创建 AlternativesLogMessage 对象
		logMessage := &AlternativesLogMessage{
			Type:      "alternatives_log",
			Timestamp: match[1],
			Message:   match[2],
		}

		ParsedMsgs = append(ParsedMsgs, logMessage)
	}

	// 将解析后的消息列表转换为 JSON 格式的字符串切片
	var jsonData []string
	for _, msg := range ParsedMsgs {
		data, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}
		jsonData = append(jsonData, string(data))
	}

	return jsonData, nil
}

// AuthLogMessage 结构体用于表示 auth.log 中的日志消息
type AuthLogMessage struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Host      string `json:"host"`
	Service   string `json:"service"`
	Message   string `json:"message"`
}

// ParseAuthLogMessages 函数用于解析 auth.log 中的日志消息并返回 JSON 格式的字符串切片
func ParseAuthLogMessages() ([]string, error) {
	var ParsedMsgs []*AuthLogMessage

	// 定义 auth.log 文件的路径
	authLogPath := "/var/log/auth.log"

	// 调用函数并获取所有消息
	authLogMessages, err := readLogMessages(authLogPath)
	if err != nil {
		fmt.Printf("读取 auth.log 消息失败: %v\n", err)
		return nil, err
	}

	// 定义用于匹配 auth.log 消息的正则表达式
	re := regexp.MustCompile(`^(\w{3}\s+\d{1,2}\s\d{2}:\d{2}:\d{2})\s(\S+)\s(\S+):\s(.+)$`)

	// 解析 auth.log 消息
	for _, msg := range authLogMessages {
		// 使用正则表达式查找匹配的组
		match := re.FindStringSubmatch(msg)
		if len(match) != 5 {
			return nil, fmt.Errorf("Invalid message format")
			//continue // 不是有效的 auth.log 消息，跳过
		}

		// 提取匹配的组并创建 AuthLogMessage 对象
		logMessage := &AuthLogMessage{
			Type:      "auth_log",
			Timestamp: match[1],
			Host:      match[2],
			Service:   match[3],
			Message:   match[4],
		}

		ParsedMsgs = append(ParsedMsgs, logMessage)
	}

	// 将解析后的消息列表转换为 JSON 格式的字符串切片
	var jsonData []string
	for _, msg := range ParsedMsgs {
		data, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}
		jsonData = append(jsonData, string(data))
	}

	return jsonData, nil
}

// DpkgLogMessage 结构体用于表示 dpkg.log 中的日志消息
type DpkgLogMessage struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	ActionPkg string `json:"action_package"`
}

// ParseDpkgLogMessages 函数用于解析 dpkg.log 中的日志消息并返回 JSON 格式的字符串切片
func ParseDpkgLogMessages() ([]string, error) {
	var jsonData []string

	// 定义 dpkg.log 文件的路径
	dpkgLogPath := "/var/log/dpkg.log"

	// 调用函数并获取所有消息
	dpkgLogMessages, err := readLogMessages(dpkgLogPath)
	if err != nil {
		fmt.Printf("读取 dpkg.log 消息失败: %v\n", err)
		return nil, err
	}

	// 解析 dpkg.log 消息
	for _, msg := range dpkgLogMessages {
		// 分割消息字符串
		parts := strings.Fields(msg)
		if len(parts) < 5 {
			continue
		}

		// 创建 DpkgLogMessage 对象
		logMessage := &DpkgLogMessage{
			Type:      "dpkg_log",
			Timestamp: parts[0] + " " + parts[1],
			ActionPkg: parts[2] + " " + strings.Join(parts[3:], " "),
		}

		// 转换为 JSON 格式的字符串
		data, err := json.Marshal(logMessage)
		if err != nil {
			return nil, err
		}
		jsonData = append(jsonData, string(data))
	}

	return jsonData, nil
}

// FontconfigLogMessage 结构体用于表示 fontconfig.log 中的日志消息
type FontconfigLogMessage struct {
	Type    string `json:"type"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

// ParseFontconfigLogMessages 函数用于解析 fontconfig.log 中的日志消息并返回 JSON 格式的字符串切片
func ParseFontconfigLogMessages() ([]string, error) {
	var ParsedMsgs []*FontconfigLogMessage

	// 定义 fontconfig.log 文件的路径
	fontconfigLogPath := "/var/log/fontconfig.log"

	// 调用函数并获取所有消息
	fontconfigLogMessages, err := readLogMessages(fontconfigLogPath)
	if err != nil {
		fmt.Printf("读取 fontconfig.log 消息失败: %v\n", err)
		return nil, err
	}

	// 定义用于匹配 fontconfig.log 消息的正则表达式
	re := regexp.MustCompile(`^(.*?):\s(.*)$`)

	// 解析 fontconfig.log 消息
	for _, msg := range fontconfigLogMessages {
		// 使用正则表达式查找匹配的组
		match := re.FindStringSubmatch(msg)
		if len(match) != 3 {
			return nil, fmt.Errorf("Invalid message format")
		}

		// 提取匹配的组并创建 FontconfigLogMessage 对象
		logMessage := &FontconfigLogMessage{
			Type:    "fontconfig_log",
			Path:    match[1],
			Message: match[2],
		}

		ParsedMsgs = append(ParsedMsgs, logMessage)
	}

	// 将解析后的消息列表转换为 JSON 格式的字符串切片
	var jsonData []string
	for _, msg := range ParsedMsgs {
		data, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}
		jsonData = append(jsonData, string(data))
	}

	return jsonData, nil
}

// KernelLogMessage 结构体用于表示 kern.log 中的日志消息
type KernelLogMessage struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Hostname  string `json:"hostname"`
	Kernel    string `json:"kernel"`
	Message   string `json:"message"`
}

// ParseKernelLogMessages 函数用于解析 kern.log 中的日志消息并返回 JSON 格式的字符串切片
func ParseKernelLogMessages() ([]string, error) {
	var ParsedMsgs []*KernelLogMessage

	// 定义 kern.log 文件的路径
	kernLogPath := "/var/log/kern.log"

	// 调用函数并获取所有消息
	kernLogMessages, err := readLogMessages(kernLogPath)
	if err != nil {
		fmt.Printf("读取 kern.log 消息失败: %v\n", err)
		return nil, err
	}

	// 定义用于匹配 kern.log 消息的正则表达式
	re := regexp.MustCompile(`^(\w{3}\s+\d{1,2}\s\d{2}:\d{2}:\d{2})\s(\S+)\skernel:\s+(\[.*?\])\s+(.*?)$`)

	// 解析 kern.log 消息
	for _, msg := range kernLogMessages {
		// 使用正则表达式查找匹配的组
		match := re.FindStringSubmatch(msg)
		if len(match) != 5 {
			return nil, fmt.Errorf("Invalid message format")
		}

		// 提取匹配的组并创建 KernelLogMessage 对象
		logMessage := &KernelLogMessage{
			Type:      "kernel_log",
			Timestamp: match[1],
			Hostname:  match[2],
			Kernel:    match[3],
			Message:   match[4],
		}

		ParsedMsgs = append(ParsedMsgs, logMessage)
	}

	// 将解析后的消息列表转换为 JSON 格式的字符串切片
	var jsonData []string
	for _, msg := range ParsedMsgs {
		data, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}
		jsonData = append(jsonData, string(data))
	}

	return jsonData, nil
}

type XorgLogMessage struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

func ParseXorgLogMessages() ([]string, error) {
	var ParsedMsgs []*XorgLogMessage

	// 定义 Xorg.0.log 文件的路径
	xorgLogPath := "/var/log/Xorg.0.log"

	// 调用函数并获取所有消息
	xorgLogMessages, err := readLogMessages(xorgLogPath)
	if err != nil {
		fmt.Printf("读取 Xorg.0.log 消息失败: %v\n", err)
		return nil, err
	}

	// 定义用于匹配 Xorg.0.log 消息的正则表达式
	re := regexp.MustCompile(`^\[\s*(\d+\.\d+)\s*\]\s*(.*?)$`)

	// 解析 Xorg.0.log 消息
	for _, msg := range xorgLogMessages {
		// 使用正则表达式查找匹配的组
		match := re.FindStringSubmatch(msg)

		// 创建 XorgLogMessage 对象
		logMessage := &XorgLogMessage{
			Type: "xorg_log",
		}

		// 根据匹配结果设置时间戳和消息字段
		if len(match) == 3 {
			logMessage.Timestamp = match[1]
			logMessage.Message = match[2]
		} else {
			logMessage.Message = msg
		}

		// 将解析后的消息存储到列表中
		ParsedMsgs = append(ParsedMsgs, logMessage)
	}

	// 将解析后的消息列表转换为 JSON 格式的字符串切片
	var jsonData []string
	for _, msg := range ParsedMsgs {
		data, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}
		jsonData = append(jsonData, string(data))
	}

	return jsonData, nil
}

// DmesgLogMessage 结构体用于表示 dmesg 日志消息
type DmesgLogMessage struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Kernel    string `json:"kernel"`
	Message   string `json:"message"`
}

// ParseDmesgLogMessages 函数用于解析 dmesg 中的日志消息并返回 JSON 格式的字符串切片
func ParseDmesgLogMessages() ([]string, error) {
	var ParsedMsgs []*DmesgLogMessage

	// 调用函数并获取所有消息
	dmesgMessages, err := cmdLogFile("dmesg")
	if err != nil {
		fmt.Printf("读取 dmesg 消息失败: %v\n", err)
		return nil, err
	}

	// 定义用于匹配 dmesg 消息的正则表达式
	re := regexp.MustCompile(`^\[\s*(\d+\.\d+)\]\s*(.*?)\s*:\s*(.*)$`)

	// 解析 dmesg 消息
	for _, msg := range dmesgMessages {
		// 使用正则表达式查找匹配的组
		match := re.FindStringSubmatch(msg)
		if len(match) != 4 {
			continue
		}

		// 创建 DmesgLogMessage 对象
		logMessage := &DmesgLogMessage{
			Type:      "dmesg_log",
			Timestamp: match[1],
			Kernel:    match[2],
			Message:   match[3],
		}

		ParsedMsgs = append(ParsedMsgs, logMessage)
	}

	// 将解析后的消息列表转换为 JSON 格式的字符串切片
	var jsonData []string
	for _, msg := range ParsedMsgs {
		data, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}
		jsonData = append(jsonData, string(data))
	}

	return jsonData, nil
}

// FaillogEntry 结构体用于表示 faillog 文件中的登录失败记录条目
type FaillogEntry struct {
	Type         string `json:"type"`
	Username     string `json:"username"`
	Failures     string `json:"failures"`
	MaxFailures  string `json:"max_failures"`
	LastLogin    string `json:"last_login"`
	LastLoginUTC string `json:"last_login_utc"`
}

// ParseFaillogEntries 函数用于解析 faillog 文件中的登录失败记录并返回 JSON 格式的字符串切片
func ParseFaillogEntries() ([]string, error) {
	var ParsedEntries []*FaillogEntry

	// 调用函数并获取所有消息
	faillogEntries, err := cmdLogFile("faillog", "-a")
	if err != nil {
		fmt.Printf("读取 faillog 文件失败: %v\n", err)
		return nil, err
	}

	// 解析 faillog 文件中的登录失败记录
	for _, entry := range faillogEntries {
		// 使用空格分割每行记录
		parts := strings.Fields(entry)
		if len(parts) != 6 {
			continue
		}

		// 创建 FaillogEntry 对象
		faillogEntry := &FaillogEntry{
			Type:         "faillog_log",
			Username:     parts[0],
			Failures:     parts[1],
			MaxFailures:  parts[2],
			LastLogin:    parts[3] + " " + parts[4],
			LastLoginUTC: parts[5],
		}

		ParsedEntries = append(ParsedEntries, faillogEntry)
	}

	// 将解析后的消息列表转换为 JSON 格式的字符串切片
	var jsonData []string
	for _, entry := range ParsedEntries {
		data, err := json.Marshal(entry)
		if err != nil {
			return nil, err
		}
		jsonData = append(jsonData, string(data))
	}

	return jsonData, nil
}

// LastlogLogMessage 结构体用于表示 lastlog 日志消息
type LastlogLogMessage struct {
	Type      string `json:"type"`
	Username  string `json:"username"`
	Port      string `json:"port"`
	From      string `json:"from"`
	LastLogin string `json:"last_login"`
}

// ParseLastlogLogMessages 函数用于解析 lastlog 中的日志消息并返回 JSON 格式的字符串切片
func ParseLastlogLogMessages() ([]string, error) {
	var ParsedMsgs []*LastlogLogMessage

	// 调用 cmdLogFile 函数并获取所有消息
	lastlogMessages, err := cmdLogFile("lastlog")
	if err != nil {
		fmt.Printf("读取 lastlog 消息失败: %v\n", err)
		return nil, err
	}

	// 解析 lastlog 消息
	for _, line := range lastlogMessages {
		// 忽略标题行和空行
		if strings.Contains(line, "用户名") || line == "" {
			continue
		}

		// 使用正则表达式匹配字段
		re := regexp.MustCompile(`^\s*(\S+)\s+(\S*)\s+(\S*)\s+(.*)$`)
		match := re.FindStringSubmatch(line)
		if match == nil || len(match) < 5 {
			fmt.Printf("解析错误：字段数量不足，行内容：%s\n", line)
			continue
		}

		// 提取字段
		username := match[1]
		port := match[2]
		from := match[3]
		lastLogin := match[4]

		// 创建 LastlogLogMessage 对象
		logMessage := &LastlogLogMessage{
			Type:      "lastlog_log",
			Username:  username,
			Port:      port,
			From:      from,
			LastLogin: lastLogin,
		}

		ParsedMsgs = append(ParsedMsgs, logMessage)
	}

	// 将解析后的消息列表转换为 JSON 格式的字符串切片
	var jsonData []string
	for _, msg := range ParsedMsgs {
		data, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}
		jsonData = append(jsonData, string(data))
	}

	return jsonData, nil
}

// WtmpLogMessage 结构体用于表示 wtmp 日志消息
type WtmpLogMessage struct {
	Type       string `json:"type"`
	User       string `json:"user"`
	Terminal   string `json:"terminal"`
	SystemInfo string `json:"system_info"`
	StartTime  string `json:"start_time"`
	EndTime    string `json:"end_time"`
	Duration   string `json:"duration"`
}

// ParseWtmpLogMessages 函数用于解析 wtmp 日志消息并返回 WtmpLogMessage 结构体的切片
func ParseWtmpLogMessages() ([]string, error) {
	var ParsedEntries []*WtmpLogMessage

	// 调用函数并获取所有消息
	wtmpMessages, err := cmdLogFile("last", "--time-format", "iso")
	if err != nil {
		fmt.Printf("读取 wtmp 消息失败: %v\n", err)
		return nil, err
	}

	for _, log := range wtmpMessages {
		// 忽略结尾行和空行
		if strings.Contains(log, "wtmp begins") || log == "" {
			continue
		}

		// 提取 user 字段
		user := strings.TrimSpace(log[0:9])

		// 提取 terminal 字段
		terminalEnd := 21
		for log[terminalEnd] == ' ' {
			terminalEnd--
		}
		terminal := strings.TrimSpace(log[9 : terminalEnd+1])

		// 提取 system_info 字段
		systemInfoStart := terminalEnd + 1
		systemInfoEnd := 38
		for log[systemInfoEnd] == ' ' {
			systemInfoEnd--
		}
		systemInfo := strings.TrimSpace(log[systemInfoStart : systemInfoEnd+1])

		// 提取 start_time 字段
		startTimeEnd := 64
		for log[startTimeEnd] == ' ' {
			startTimeEnd--
		}
		startTime := strings.TrimSpace(log[38 : startTimeEnd+1])

		endTime := ""
		duration := ""
		if len(log) > 95 {
			// 提取 end_time 字段
			endTimeStart := 67
			for log[endTimeStart] != ' ' {
				endTimeStart++
			}
			endTime = strings.TrimSpace(log[67:endTimeStart])

			// 提取 duration 字段
			durationStart := 94
			duration = strings.TrimSpace(log[durationStart:])
		} else {
			endTime = strings.TrimSpace(log[67:])
			duration = endTime
		}

		entry := &WtmpLogMessage{
			Type:       "wtmp_log",
			User:       user,
			Terminal:   terminal,
			SystemInfo: systemInfo,
			StartTime:  startTime,
			EndTime:    endTime,
			Duration:   duration,
		}

		ParsedEntries = append(ParsedEntries, entry)
	}
	// 将解析后的日志条目列表转换为 JSON 格式的字符串切片
	var jsonData []string
	for _, entry := range ParsedEntries {
		data, err := json.Marshal(entry)
		if err != nil {
			return nil, err
		}
		jsonData = append(jsonData, string(data))
	}

	return jsonData, nil
}
