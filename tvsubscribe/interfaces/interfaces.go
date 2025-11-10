package interfaces

import "tvsubscribe"

// ConfigManager 配置管理器接口
type ConfigManager interface {
	GetConfig() map[string]interface{}
	UpdateConfig(updates map[string]interface{}) error
}

// SubscribeManager 订阅管理器接口
type SubscribeManager interface {
	GetSubscribes() []tvsubscribe.TVInfo
	AddSubscribe(tvInfo tvsubscribe.TVInfo) error
	RemoveSubscribe(tvInfo tvsubscribe.TVInfo) error
}