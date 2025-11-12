package subscribe

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"tvsubscribe"
)

// SubscribeManager 订阅管理器
type SubscribeManager struct {
	subscribes    []tvsubscribe.TVInfo
	subscribePath string
	mu            sync.RWMutex
}

// loadSubscribes 从订阅文件加载订阅列表
func loadSubscribes(subscribePath string) ([]tvsubscribe.TVInfo, error) {
	// 如果文件不存在，返回空列表
	if _, err := os.Stat(subscribePath); os.IsNotExist(err) {
		return []tvsubscribe.TVInfo{}, nil
	}

	data, err := os.ReadFile(subscribePath)
	if err != nil {
		return nil, fmt.Errorf("读取订阅文件失败: %v", err)
	}

	// 如果文件为空，返回空列表
	if len(data) == 0 {
		return []tvsubscribe.TVInfo{}, nil
	}

	var subscribes []tvsubscribe.TVInfo
	if err := json.Unmarshal(data, &subscribes); err != nil {
		return nil, fmt.Errorf("解析订阅文件失败: %v", err)
	}

	return subscribes, nil
}

// saveSubscribes 保存订阅列表到文件
func saveSubscribes(subscribePath string, subscribes []tvsubscribe.TVInfo) error {
	data, err := json.MarshalIndent(subscribes, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化订阅数据失败: %v", err)
	}

	if err := os.WriteFile(subscribePath, data, 0644); err != nil {
		return fmt.Errorf("写入订阅文件失败: %v", err)
	}

	return nil
}

// NewSubscribeManager 创建新的订阅管理器
func NewSubscribeManager(subscribePath string) (*SubscribeManager, error) {
	subscribes, err := loadSubscribes(subscribePath)
	if err != nil {
		return nil, err
	}

	// 获取订阅文件的绝对路径
	absPath, err := filepath.Abs(subscribePath)
	if err != nil {
		return nil, fmt.Errorf("获取订阅文件绝对路径失败: %v", err)
	}

	manager := &SubscribeManager{
		subscribes:    subscribes,
		subscribePath: absPath,
	}

	return manager, nil
}

// GetSubscribes 获取当前订阅列表（线程安全）
func (m *SubscribeManager) GetSubscribes() []tvsubscribe.TVInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回副本，避免外部修改
	result := make([]tvsubscribe.TVInfo, len(m.subscribes))
	copy(result, m.subscribes)
	return result
}

// AddSubscribe 添加订阅
func (m *SubscribeManager) AddSubscribe(tvInfo tvsubscribe.TVInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查是否已存在相同的订阅
	for _, existing := range m.subscribes {
		if existing.DouBanID == tvInfo.DouBanID && existing.Resolution == tvInfo.Resolution {
			return fmt.Errorf("订阅已存在: 豆瓣ID=%s, 分辨率=%d", tvInfo.DouBanID, tvInfo.Resolution)
		}
	}

	// 如果名称为空，尝试从豆瓣获取
	if tvInfo.Name == "" {
		name, err := tvsubscribe.GetTVNameByDouBanID(tvInfo.DouBanID)
		if err != nil {
			// 获取名称失败，但不阻止添加订阅，使用默认名称
			tvInfo.Name = fmt.Sprintf("豆瓣ID: %s", tvInfo.DouBanID)
		} else {
			tvInfo.Name = name
		}
	}

	// 添加新订阅
	m.subscribes = append(m.subscribes, tvInfo)

	// 保存到文件
	if err := saveSubscribes(m.subscribePath, m.subscribes); err != nil {
		// 回滚内存中的修改
		m.subscribes = m.subscribes[:len(m.subscribes)-1]
		return err
	}

	return nil
}

// RemoveSubscribe 删除订阅
func (m *SubscribeManager) RemoveSubscribe(tvInfo tvsubscribe.TVInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 查找并删除订阅
	found := false
	newSubscribes := make([]tvsubscribe.TVInfo, 0, len(m.subscribes))
	for _, existing := range m.subscribes {
		if !(existing.DouBanID == tvInfo.DouBanID && existing.Resolution == tvInfo.Resolution) {
			newSubscribes = append(newSubscribes, existing)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("订阅不存在: 豆瓣ID=%s, 分辨率=%d", tvInfo.DouBanID, tvInfo.Resolution)
	}

	// 保存到文件
	if err := saveSubscribes(m.subscribePath, newSubscribes); err != nil {
		return err
	}

	m.subscribes = newSubscribes
	return nil
}