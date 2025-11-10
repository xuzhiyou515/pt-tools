package config

// Config 应用配置
type Config struct {
	Endpoint        string `json:"endpoint"`
	Cookie          string `json:"cookie"`
	IntervalMinutes int    `json:"interval_minutes"`
	WeChatServer    string `json:"wechat_server"`
	WeChatToken     string `json:"wechat_token"`
	Port            int    `json:"port"`
}