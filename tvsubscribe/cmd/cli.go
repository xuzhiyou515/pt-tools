package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"tvsubscribe"
	"tvsubscribe/client"
)

// parseKeyValuePairs 解析key=value格式的参数
func parseKeyValuePairs(args []string) map[string]string {
	result := make(map[string]string)
	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

// handleConfigCommand 处理config命令
func handleConfigCommand(args []string) {
	var serverURL string
	var listFlag bool
	var setFlag bool

	configCmd := flag.NewFlagSet("config", flag.ExitOnError)
	configCmd.StringVar(&serverURL, "url", "127.0.0.1:8443", "服务器地址")
	configCmd.BoolVar(&listFlag, "list", false, "获取配置")
	configCmd.BoolVar(&setFlag, "set", false, "设置配置")

	configCmd.Parse(args)

	if !listFlag && !setFlag {
		fmt.Println("使用方法: tvsubscribe config [选项]")
		fmt.Println("选项:")
		fmt.Println("  --list              获取配置")
		fmt.Println("  --set key=value...  设置配置")
		fmt.Println("  --url string        服务器地址 (默认 \"127.0.0.1:8443\")")
		os.Exit(1)
	}

	client := client.NewClient("http://" + serverURL)

	if listFlag {
		configMap, err := client.GetConfig()
		if err != nil {
			log.Fatalf("获取配置失败: %v", err)
		}

		jsonData, err := json.MarshalIndent(configMap, "", "  ")
		if err != nil {
			log.Fatalf("序列化配置失败: %v", err)
		}

		fmt.Println(string(jsonData))
		return
	}

	if setFlag {
		setArgs := configCmd.Args()
		if len(setArgs) == 0 {
			log.Fatal("设置配置时需要提供 key=value 参数")
		}

		kvPairs := parseKeyValuePairs(setArgs)
		if len(kvPairs) == 0 {
			log.Fatal("无效的 key=value 参数格式")
		}

		// 构建更新配置
		updateConfig := make(map[string]interface{})
		updated := false

		for key, value := range kvPairs {
			switch key {
			case "endpoint":
				updateConfig["endpoint"] = value
				updated = true
			case "cookie":
				updateConfig["cookie"] = value
				updated = true
			case "interval_minutes":
				if interval, err := strconv.Atoi(value); err == nil && interval > 0 {
					updateConfig["interval_minutes"] = interval
					updated = true
				} else {
					log.Printf("警告: 无效的 interval_minutes 值: %s", value)
				}
			case "wechat_server":
				updateConfig["wechat_server"] = value
				updated = true
			case "wechat_token":
				updateConfig["wechat_token"] = value
				updated = true
			case "port":
				if port, err := strconv.Atoi(value); err == nil && port > 0 {
					updateConfig["port"] = port
					updated = true
				} else {
					log.Printf("警告: 无效的 port 值: %s", value)
				}
			default:
				log.Printf("警告: 未知的配置项: %s", key)
			}
		}

		if !updated {
			log.Fatal("没有有效的配置项被更新")
		}

		if err := client.SetConfig(updateConfig); err != nil {
			log.Fatalf("设置配置失败: %v", err)
		}

		fmt.Println("配置更新成功")
	}
}

// handleSubscribeCommand 处理subscribe命令
func handleSubscribeCommand(args []string) {
	var serverURL string
	var listFlag bool
	var addFlag bool
	var delFlag bool

	subscribeCmd := flag.NewFlagSet("subscribe", flag.ExitOnError)
	subscribeCmd.StringVar(&serverURL, "url", "127.0.0.1:8443", "服务器地址")
	subscribeCmd.BoolVar(&listFlag, "list", false, "获取订阅列表")
	subscribeCmd.BoolVar(&addFlag, "add", false, "添加订阅")
	subscribeCmd.BoolVar(&delFlag, "del", false, "删除订阅")

	subscribeCmd.Parse(args)

	if !listFlag && !addFlag && !delFlag {
		fmt.Println("使用方法: tvsubscribe subscribe [选项]")
		fmt.Println("选项:")
		fmt.Println("  --list                          获取订阅列表")
		fmt.Println("  --add douban_id=xxx...          添加订阅")
		fmt.Println("  --del douban_id=xxx...          删除订阅")
		fmt.Println("  --url string                    服务器地址 (默认 \"127.0.0.1:8443\")")
		fmt.Println()
		fmt.Println("添加/删除订阅的参数格式:")
		fmt.Println("  douban_id=豆瓣ID (必填)")
		fmt.Println("  resolution=分辨率 (可选，默认为1)")
		os.Exit(1)
	}

	client := client.NewClient("http://" + serverURL)

	if listFlag {
		subscribes, err := client.GetSubscribeList()
		if err != nil {
			log.Fatalf("获取订阅列表失败: %v", err)
		}

		if len(subscribes) == 0 {
			fmt.Println("暂无订阅")
			return
		}

		jsonData, err := json.MarshalIndent(subscribes, "", "  ")
		if err != nil {
			log.Fatalf("序列化订阅列表失败: %v", err)
		}

		fmt.Println(string(jsonData))
		return
	}

	if addFlag || delFlag {
		cmdArgs := subscribeCmd.Args()
		if len(cmdArgs) == 0 {
			log.Fatalf("%s订阅时需要提供参数", map[bool]string{true: "添加", false: "删除"}[addFlag])
		}

		kvPairs := parseKeyValuePairs(cmdArgs)
		if len(kvPairs) == 0 {
			log.Fatal("无效的参数格式")
		}

		// 构造TVInfo
		tvInfo := tvsubscribe.TVInfo{}
		if doubanID, ok := kvPairs["douban_id"]; ok {
			tvInfo.DouBanID = doubanID
		} else {
			log.Fatal("必须提供 douban_id 参数")
		}

		if resolution, ok := kvPairs["resolution"]; ok {
			if res, err := strconv.Atoi(resolution); err == nil && res > 0 {
				tvInfo.Resolution = res
			} else {
				log.Printf("警告: 无效的 resolution 值: %s，使用默认值 1", resolution)
				tvInfo.Resolution = 1
			}
		} else {
			tvInfo.Resolution = 1 // 默认分辨率
		}

		if addFlag {
			if err := client.AddSubscribe(tvInfo); err != nil {
				log.Fatalf("添加订阅失败: %v", err)
			}
			fmt.Printf("订阅添加成功: 豆瓣ID=%s, 分辨率=%d\n", tvInfo.DouBanID, tvInfo.Resolution)
		} else {
			if err := client.DelSubscribe(tvInfo); err != nil {
				log.Fatalf("删除订阅失败: %v", err)
			}
			fmt.Printf("订阅删除成功: 豆瓣ID=%s, 分辨率=%d\n", tvInfo.DouBanID, tvInfo.Resolution)
		}
	}
}

// RunCLI 运行命令行界面
func RunCLI() {
	if len(os.Args) < 2 {
		fmt.Println("使用方法: tvsubscribe <command> [options]")
		fmt.Println("命令:")
		fmt.Println("  config      配置管理")
		fmt.Println("  subscribe   订阅管理")
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "config":
		handleConfigCommand(args)
	case "subscribe":
		handleSubscribeCommand(args)
	default:
		log.Fatalf("未知命令: %s", command)
	}
}