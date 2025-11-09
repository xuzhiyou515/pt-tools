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

	"github.com/fsnotify/fsnotify"
	"tvsubscribe"
)

type Config struct {
	Endpoint       string                `json:"endpoint"`
	Cookie         string                `json:"cookie"`
	Passkey        string                `json:"passkey"`
	IntervalMinutes int                   `json:"interval_minutes"`
	WeChatServer   string                `json:"wechat_server"`
	WeChatToken    string                `json:"wechat_token"`
	Subscribes     []tvsubscribe.TVInfo `json:"subscribes"`
}

// ConfigManager 配置管理器，支持热重载
type ConfigManager struct {
	config      *Config
	configPath  string
	mu          sync.RWMutex
	watcher     *fsnotify.Watcher
	onReload    func(*Config) // 配置重载后的回调函数
}

// loadConfig 从配置文件加载配置
func loadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	// 验证必填字段
	if config.Cookie == "" {
		return nil, fmt.Errorf("配置文件中 cookie 不能为空")
	}
	if config.Passkey == "" {
		return nil, fmt.Errorf("配置文件中 passkey 不能为空")
	}
	if config.IntervalMinutes <= 0 {
		config.IntervalMinutes = 60 // 默认60分钟
	}

	return &config, nil
}

// NewConfigManager 创建新的配置管理器
func NewConfigManager(configPath string, onReload func(*Config)) (*ConfigManager, error) {
	config, err := loadConfig(configPath)
	if err != nil {
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("创建文件监视器失败: %v", err)
	}

	// 获取配置文件的绝对路径
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		watcher.Close()
		return nil, fmt.Errorf("获取配置文件绝对路径失败: %v", err)
	}

	// 添加配置文件到监视器
	if err := watcher.Add(absPath); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("添加配置文件到监视器失败: %v", err)
	}

	manager := &ConfigManager{
		config:     config,
		configPath: absPath,
		watcher:    watcher,
		onReload:   onReload,
	}

	return manager, nil
}

// GetConfig 获取当前配置（线程安全）
func (m *ConfigManager) GetConfig() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// ReloadConfig 重新加载配置
func (m *ConfigManager) ReloadConfig() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	config, err := loadConfig(m.configPath)
	if err != nil {
		return err
	}

	m.config = config
	log.Printf("配置文件已重新加载，订阅数量: %d", len(config.Subscribes))

	// 调用重载回调函数
	if m.onReload != nil {
		m.onReload(config)
	}

	return nil
}

// WatchConfigChanges 监视配置文件变化
func (m *ConfigManager) WatchConfigChanges() {
	go func() {
		for {
			select {
			case event, ok := <-m.watcher.Events:
				if !ok {
					return
				}
				// 只处理写入和重命名事件
				if event.Op.Has(fsnotify.Write) || event.Op.Has(fsnotify.Rename) {
					// 小延迟，确保文件写入完成
					time.Sleep(100 * time.Millisecond)
					if err := m.ReloadConfig(); err != nil {
						log.Printf("重新加载配置文件失败: %v", err)
					}
				}
			case err, ok := <-m.watcher.Errors:
				if !ok {
					return
				}
				log.Printf("配置文件监视错误: %v", err)
			}
		}
	}()
}

// Close 关闭配置管理器
func (m *ConfigManager) Close() {
	if m.watcher != nil {
		m.watcher.Close()
	}
}

// processTVSubscribes 处理所有订阅的电视剧
func processTVSubscribes(config *Config) {
	log.Println("开始处理电视剧订阅...")

	for _, tv := range config.Subscribes {
		log.Printf("处理豆瓣ID: %s, 分辨率: %d", tv.DouBanID, tv.Resolution)

		// 查询种子列表
		torrentIDs, err := tvsubscribe.QueryTorrentList(config.Cookie, &tv)
		if err != nil {
			log.Printf("查询种子列表失败 (豆瓣ID: %s): %v", tv.DouBanID, err)
			continue
		}

		if len(torrentIDs) == 0 {
			log.Printf("未找到可下载的种子 (豆瓣ID: %s)", tv.DouBanID)
			continue
		}

		log.Printf("找到 %d 个种子 (豆瓣ID: %s)", len(torrentIDs), tv.DouBanID)

		// 下载种子
		if err := tvsubscribe.DownloadTorrent(torrentIDs, config.Passkey, config.Endpoint, config.WeChatServer, config.WeChatToken); err != nil {
			log.Printf("下载种子失败 (豆瓣ID: %s): %v", tv.DouBanID, err)
		} else {
			log.Printf("成功处理 %d 个种子 (豆瓣ID: %s)", len(torrentIDs), tv.DouBanID)
		}
	}

	log.Println("电视剧订阅处理完成")
}

// startScheduler 启动定时任务
func startScheduler(configManager *ConfigManager) {
	// 立即执行一次
	processTVSubscribes(configManager.GetConfig())

	// 定时执行
	go func() {
		for {
			config := configManager.GetConfig()
			interval := time.Duration(config.IntervalMinutes) * time.Minute

			log.Printf("定时任务等待 %v 后执行", interval)
			time.Sleep(interval)

			processTVSubscribes(config)
		}
	}()
}

func main() {
	// 配置重载后的回调函数
	onConfigReload := func(config *Config) {
		log.Println("配置已更新，立即重新处理所有订阅的电视剧")
		processTVSubscribes(config)
	}

	// 创建配置管理器
	configManager, err := NewConfigManager("./config.json", onConfigReload)
	if err != nil {
		log.Fatalf("配置管理器创建失败: %v", err)
	}
	defer configManager.Close()

	// 获取初始配置
	config := configManager.GetConfig()
	log.Printf("配置加载成功，订阅数量: %d, 检查间隔: %d 分钟",
		len(config.Subscribes), config.IntervalMinutes)

	// 启动配置文件监视
	configManager.WatchConfigChanges()
	log.Println("配置文件监视已启动，修改配置文件将自动重新加载并立即执行")

	// 启动定时任务
	startScheduler(configManager)

	// 设置信号监听，优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("程序已启动，按 Ctrl+C 退出")

	// 等待退出信号
	sig := <-sigChan
	log.Printf("接收到信号: %v，正在退出...", sig)
}