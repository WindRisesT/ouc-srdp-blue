package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
    "os/exec"
    "strings"
)

// 定义结构体来存储软件包信息
type PackageInfo struct {
    Name        string `json:"name"`
    Version     string `json:"version"`
    Description string `json:"description"`
    Status      string `json:"status"`
}

// 获取软件包信息
func getPackageInfos() ([]PackageInfo, error) {
    var packageInfos []PackageInfo
    // 执行 dpkg --list 命令获取软件列表
    cmd := exec.Command("dpkg", "--list")
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("执行命令出错: %w", err)
    }

    // 逐行解析输出
    scanner := bufio.NewScanner(strings.NewReader(string(output)))
    for scanner.Scan() {
        line := scanner.Text()
        // 跳过标题行和空行
        if strings.HasPrefix(line, "Desired=Unknown/Install/Remove/Purge/Hold") ||
            strings.HasPrefix(line, "| Status=Not/Inst/Conf-files/Unpacked/halF-conf/Half-inst/trig-aWait/Trig-pend") ||
            strings.HasPrefix(line, "|/ Err?=(none)/Reinst-required (Status,Err: uppercase=bad)") ||
            strings.HasPrefix(line, "||/ Name") ||
            strings.TrimSpace(line) == "" {
            continue
        }

        fields := strings.Fields(line)
        if len(fields) >= 4 {
            packageName := fields[1]
            packageVersion := fields[2]
            packageStatus := fields[0]
            // 尝试提取描述信息，描述信息可能跨多个字段
            packageDescription := ""
            if len(fields) > 4 {
                packageDescription = strings.Join(fields[4:], " ")
            }
            packageInfos = append(packageInfos, PackageInfo{
                Name:        packageName,
                Version:     packageVersion,
                Description: packageDescription,
                Status:      packageStatus,
            })
        }
    }

    if err := scanner.Err(); err != nil {
        return nil, fmt.Errorf("读取输出出错: %w", err)
    }
    return packageInfos, nil
}

// 将数据保存到 JSON 文件
func saveToJSONFile(data interface{}, filePath string) error {
    // 将软件包信息转换为 JSON 格式
    jsonData, err := json.MarshalIndent(data, "", "  ")
    if err != nil {
        return fmt.Errorf("转换为 JSON 出错: %w", err)
    }

    // 打开文件用于写入
    file, err := os.Create(filePath)
    if err != nil {
        return fmt.Errorf("创建文件出错: %w", err)
    }
    defer file.Close()

    // 将 JSON 数据写入文件
    _, err = file.Write(jsonData)
    if err != nil {
        return fmt.Errorf("写入文件出错: %w", err)
    }
    return nil
}

func main() {
    packageInfos, err := getPackageInfos()
    if err != nil {
        fmt.Println(err)
        return
    }

    filePath := "software_list.json"
    err = saveToJSONFile(packageInfos, filePath)
    if err != nil {
        fmt.Println(err)
        return
    }

    fmt.Printf("软件列表数据已成功保存为 %s\n", filePath)
}
