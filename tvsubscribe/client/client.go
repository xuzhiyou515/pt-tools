package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"tvsubscribe"
)

// Client HTTP客户端
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient 创建新的HTTP客户端
func NewClient(serverURL string) *Client {
	return &Client{
		baseURL: serverURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetConfig 获取配置
func (c *Client) GetConfig() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/getConfig", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("服务器返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var response struct {
		Success bool                   `json:"success"`
		Message string                 `json:"message"`
		Data    map[string]interface{} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("操作失败: %s", response.Message)
	}

	return response.Data, nil
}

// SetConfig 设置配置
func (c *Client) SetConfig(config map[string]interface{}) error {
	url := fmt.Sprintf("%s/setConfig", c.baseURL)

	jsonData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("服务器返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	var response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    interface{} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if !response.Success {
		return fmt.Errorf("操作失败: %s", response.Message)
	}

	return nil
}

// GetSubscribeList 获取订阅列表
func (c *Client) GetSubscribeList() ([]tvsubscribe.TVInfo, error) {
	url := fmt.Sprintf("%s/getSubscribeList", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("服务器返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var response struct {
		Success bool              `json:"success"`
		Message string            `json:"message"`
		Data    []tvsubscribe.TVInfo `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("操作失败: %s", response.Message)
	}

	return response.Data, nil
}

// AddSubscribe 添加订阅
func (c *Client) AddSubscribe(tvInfo tvsubscribe.TVInfo) error {
	url := fmt.Sprintf("%s/addSubscribe", c.baseURL)

	jsonData, err := json.Marshal(tvInfo)
	if err != nil {
		return fmt.Errorf("序列化订阅信息失败: %v", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("服务器返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	var response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    interface{} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if !response.Success {
		return fmt.Errorf("操作失败: %s", response.Message)
	}

	return nil
}

// DelSubscribe 删除订阅
func (c *Client) DelSubscribe(tvInfo tvsubscribe.TVInfo) error {
	url := fmt.Sprintf("%s/delSubscribe", c.baseURL)

	jsonData, err := json.Marshal(tvInfo)
	if err != nil {
		return fmt.Errorf("序列化订阅信息失败: %v", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("服务器返回错误状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	var response struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    interface{} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	if !response.Success {
		return fmt.Errorf("操作失败: %s", response.Message)
	}

	return nil
}