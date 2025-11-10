package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"tvsubscribe"
	"tvsubscribe/config"
	"tvsubscribe/server"
	"tvsubscribe/subscribe"
)

// 辅助函数
func getString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func getInt(v interface{}) int {
	if i, ok := v.(int); ok {
		return i
	}
	if f, ok := v.(float64); ok {
		return int(f)
	}
	return 0
}


// ConfigManager 配置管理器
type ConfigManager struct {
	config      *config.Config
	configPath  string
	mu          sync.RWMutex
}

// loadConfig 从配置文件加载配置
func loadConfig(configPath string) (*config.Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var cfg config.Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 验证必填字段
	if cfg.Cookie == "" {
		return nil, fmt.Errorf("配置文件中 cookie 不能为空")
	}
	if cfg.IntervalMinutes <= 0 {
		cfg.IntervalMinutes = 60 // 默认60分钟
	}
	if cfg.Port <= 0 {
		cfg.Port = 8443 // 默认8443端口
	}

	return &cfg, nil
}

// NewConfigManager 创建新的配置管理器
func NewConfigManager(configPath string) (*ConfigManager, error) {
	config, err := loadConfig(configPath)
	if err != nil {
		return nil, err
	}

	// 获取配置文件的绝对路径
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("获取配置文件绝对路径失败: %v", err)
	}

	manager := &ConfigManager{
		config:     config,
		configPath: absPath,
	}

	return manager, nil
}

// GetConfig 获取当前配置（线程安全）
func (m *ConfigManager) GetConfig() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 将config.Config转换为map
	result := map[string]interface{}{
		"endpoint":         m.config.Endpoint,
		"cookie":           m.config.Cookie,
		"interval_minutes": m.config.IntervalMinutes,
		"wechat_server":    m.config.WeChatServer,
		"wechat_token":     m.config.WeChatToken,
		"port":             m.config.Port,
	}
	return result
}

// UpdateConfig 更新配置
func (m *ConfigManager) UpdateConfig(updates map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	updated := false
	if endpoint, ok := updates["endpoint"].(string); ok && endpoint != "" {
		m.config.Endpoint = endpoint
		updated = true
	}
	if cookie, ok := updates["cookie"].(string); ok && cookie != "" {
		m.config.Cookie = cookie
		updated = true
	}
	if interval, ok := updates["interval_minutes"].(float64); ok && interval > 0 {
		m.config.IntervalMinutes = int(interval)
		updated = true
	}
	if wechatServer, ok := updates["wechat_server"].(string); ok && wechatServer != "" {
		m.config.WeChatServer = wechatServer
		updated = true
	}
	if wechatToken, ok := updates["wechat_token"].(string); ok && wechatToken != "" {
		m.config.WeChatToken = wechatToken
		updated = true
	}
	if port, ok := updates["port"].(float64); ok && port > 0 {
		m.config.Port = int(port)
		updated = true
	}

	if !updated {
		return fmt.Errorf("没有有效的配置字段被更新")
	}

	// 保存配置到文件
	return m.saveConfig()
}

// saveConfig 保存配置到文件
func (m *ConfigManager) saveConfig() error {
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}

	return nil
}

// processSingleTV 处理单个电视剧订阅
func processSingleTV(configMgr *ConfigManager, tvInfo tvsubscribe.TVInfo) {
	// 获取实际的config对象
	configMap := configMgr.GetConfig()

	// 模拟config对象的行为
	config := struct {
		Endpoint        string
		Cookie          string
		IntervalMinutes int
		WeChatServer    string
		WeChatToken     string
		Port            int
	}{
		Endpoint:        getString(configMap["endpoint"]),
		Cookie:          getString(configMap["cookie"]),
		IntervalMinutes: getInt(configMap["interval_minutes"]),
		WeChatServer:    getString(configMap["wechat_server"]),
		WeChatToken:     getString(configMap["wechat_token"]),
		Port:            getInt(configMap["port"]),
	}

	log.Printf("处理豆瓣ID: %s, 分辨率: %d", tvInfo.DouBanID, tvInfo.Resolution)

	// 查询种子列表
	torrentInfos, err := tvsubscribe.QueryTorrentList(config.Cookie, &tvInfo)
	if err != nil {
		log.Printf("查询种子列表失败 (豆瓣ID: %s): %v", tvInfo.DouBanID, err)
		return
	}

	if len(torrentInfos) == 0 {
		log.Printf("未找到可下载的种子 (豆瓣ID: %s)", tvInfo.DouBanID)
		return
	}

	log.Printf("找到 %d 个种子 (豆瓣ID: %s)", len(torrentInfos), tvInfo.DouBanID)

	// 下载种子
	if err := tvsubscribe.DownloadTorrent(torrentInfos, config.Endpoint, config.WeChatServer, config.WeChatToken); err != nil {
		log.Printf("下载种子失败 (豆瓣ID: %s): %v", tvInfo.DouBanID, err)
	} else {
		log.Printf("成功处理 %d 个种子 (豆瓣ID: %s)", len(torrentInfos), tvInfo.DouBanID)
	}
}

// processTVSubscribes 处理所有订阅的电视剧
func processTVSubscribes(configMgr *ConfigManager, subscribes []tvsubscribe.TVInfo) {
	// 获取实际的config对象
	configMap := configMgr.GetConfig()

	// 模拟config对象的行为
	config := struct {
		Endpoint        string
		Cookie          string
		IntervalMinutes int
		WeChatServer    string
		WeChatToken     string
		Port            int
	}{
		Endpoint:        getString(configMap["endpoint"]),
		Cookie:          getString(configMap["cookie"]),
		IntervalMinutes: getInt(configMap["interval_minutes"]),
		WeChatServer:    getString(configMap["wechat_server"]),
		WeChatToken:     getString(configMap["wechat_token"]),
		Port:            getInt(configMap["port"]),
	}
	log.Println("开始处理电视剧订阅...")

	if len(subscribes) == 0 {
		log.Println("没有订阅的电视剧")
		return
	}

	for _, tv := range subscribes {
		log.Printf("处理豆瓣ID: %s, 分辨率: %d", tv.DouBanID, tv.Resolution)

		// 查询种子列表
		torrentInfos, err := tvsubscribe.QueryTorrentList(config.Cookie, &tv)
		if err != nil {
			log.Printf("查询种子列表失败 (豆瓣ID: %s): %v", tv.DouBanID, err)
			continue
		}

		if len(torrentInfos) == 0 {
			log.Printf("未找到可下载的种子 (豆瓣ID: %s)", tv.DouBanID)
			continue
		}

		log.Printf("找到 %d 个种子 (豆瓣ID: %s)", len(torrentInfos), tv.DouBanID)

		// 下载种子
		if err := tvsubscribe.DownloadTorrent(torrentInfos, config.Endpoint, config.WeChatServer, config.WeChatToken); err != nil {
			log.Printf("下载种子失败 (豆瓣ID: %s): %v", tv.DouBanID, err)
		} else {
			log.Printf("成功处理 %d 个种子 (豆瓣ID: %s)", len(torrentInfos), tv.DouBanID)
		}
	}

	log.Println("电视剧订阅处理完成")
}

// startScheduler 启动定时任务
func startScheduler(configManager *ConfigManager, subscribeManager *subscribe.SubscribeManager) {
	// 立即执行一次
	processTVSubscribes(configManager, subscribeManager.GetSubscribes())

	// 定时执行
	go func() {
		for {
			configMap := configManager.GetConfig()
			interval := time.Duration(getInt(configMap["interval_minutes"])) * time.Minute

			log.Printf("定时任务等待 %v 后执行", interval)
			time.Sleep(interval)

			processTVSubscribes(configManager, subscribeManager.GetSubscribes())
		}
	}()
}

func main() {
	// 检查是否为CLI模式
	if len(os.Args) > 1 {
		RunCLI()
		return
	}

	// 服务器模式
	// 创建配置管理器
	configManager, err := NewConfigManager("./config.json")
	if err != nil {
		log.Fatalf("配置管理器创建失败: %v", err)
	}

	// 创建订阅管理器
	subscribeManager, err := subscribe.NewSubscribeManager("./subscribes.json")
	if err != nil {
		log.Fatalf("订阅管理器创建失败: %v", err)
	}

	// 获取初始配置
	configMap := configManager.GetConfig()
	log.Printf("配置加载成功，监听端口: %d, 检查间隔: %d 分钟", getInt(configMap["port"]), getInt(configMap["interval_minutes"]))

	// 设置信号监听，优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 创建处理函数
	processTVFunc := func() {
		subscribes := subscribeManager.GetSubscribes()
		processTVSubscribes(configManager, subscribes)
	}

	processSingleFunc := func(tvInfo tvsubscribe.TVInfo) {
		processSingleTV(configManager, tvInfo)
	}

	// 创建HTTP服务器
	httpServer := server.NewServer(configManager, subscribeManager, processTVFunc, processSingleFunc)

	// 启动定时任务
	startScheduler(configManager, subscribeManager)

	// 在单独的goroutine中启动HTTP服务器
	go func() {
		if err := httpServer.Start(); err != nil {
			log.Printf("HTTP服务器启动失败: %v", err)
			// HTTP服务器启动失败，退出程序
			sigChan <- syscall.SIGTERM
		}
	}()

	log.Println("程序已启动，按 Ctrl+C 退出")
	log.Printf("HTTP API服务器已启动，访问地址: http://127.0.0.1:%d", getInt(configMap["port"]))

	// 等待退出信号
	sig := <-sigChan
	log.Printf("接收到信号: %v，正在退出...", sig)
}