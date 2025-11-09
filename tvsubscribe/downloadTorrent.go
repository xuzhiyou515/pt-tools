package tvsubscribe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/hekmon/transmissionrpc/v3"
)

// 种子下载链接
// 种子id 577692 passkey 123456 https://springsunday.net/download.php?id=577692&passkey=123456&https=1

// WeChatMessageRequest 微信消息发送请求
type WeChatMessageRequest struct {
	Token   string `json:"token"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// WeChatMessageResponse 微信消息发送响应
type WeChatMessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// sendWeChatMessage 发送微信消息
func sendWeChatMessage(serverURL, token, title, content string) error {
	if serverURL == "" || token == "" {
		return fmt.Errorf("微信服务器配置不完整，跳过消息发送")
	}

	request := WeChatMessageRequest{
		Token:   token,
		Title:   title,
		Content: content,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("序列化微信消息失败: %v", err)
	}

	resp, err := http.Post(serverURL+"/send-message", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("发送微信消息失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("微信消息发送失败，状态码: %d", resp.StatusCode)
	}

	var response WeChatMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("解析微信消息响应失败: %v", err)
	}

	if !response.Success {
		return fmt.Errorf("微信消息发送失败: %s", response.Error)
	}

	return nil
}

// buildDownloadURL 构建种子下载链接
func buildDownloadURL(id, passkey string) string {
	return fmt.Sprintf("https://springsunday.net/download.php?id=%s&passkey=%s&https=1", id, passkey)
}
func downloadFile(url string, path string) error {
	// 创建HTTP请求
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 创建目录
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 创建文件
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	// 复制数据
	_, err = io.Copy(out, resp.Body)
	return err
}

// addTorrentToTransmission 通过 transmissionrpc 库添加种子到 Transmission
func addTorrentToTransmission(torrentPath, endpoint string) (*transmissionrpc.Torrent, error) {
	// Transmission RPC 配置
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("解析 Transmission 端点失败: %v", err)
	}

	// 创建 Transmission 客户端
	client, err := transmissionrpc.New(endpointURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建 Transmission 客户端失败: %v", err)
	}
	torrent, err := client.TorrentAddFile(context.TODO(), torrentPath)
	if err != nil {
		return nil, fmt.Errorf("添加种子文件失败: %v", err)
	}
	return &torrent, nil
}

// downloadATorrent 下载单个种子文件并添加到 Transmission
func downloadATorrent(id, passkey, path, endpoint, wechatServer, wechatToken string) error {
	// 下载种子文件
	err := downloadFile(buildDownloadURL(id, passkey), path)
	if err != nil {
		// 删除可能已创建的不完整文件
		os.Remove(path)
		// 发送下载失败通知
		if sendErr := sendWeChatMessage(wechatServer, wechatToken,
			"种子下载失败", fmt.Sprintf("种子ID: %s\n错误信息: %v", id, err)); sendErr != nil {
			fmt.Printf("发送下载失败通知失败: %v\n", sendErr)
		}
		return fmt.Errorf("下载种子文件失败: %v", err)
	}

	// 添加到 Transmission
	torrent, err := addTorrentToTransmission(path, endpoint)
	if err != nil {
		// 删除种子文件
		os.Remove(path)
		// 发送添加失败通知
		if sendErr := sendWeChatMessage(wechatServer, wechatToken,
			"添加种子失败", fmt.Sprintf("种子ID: %s\n错误信息: %v", id, err)); sendErr != nil {
			fmt.Printf("发送添加失败通知失败: %v\n", sendErr)
		}
		return fmt.Errorf("添加种子到 Transmission 失败: %v", err)
	}

	// 发送成功通知
	if err := sendWeChatMessage(wechatServer, wechatToken,
		"种子下载成功", fmt.Sprintf("种子名称: %s\n种子ID: %s\n已成功添加到 Transmission", *torrent.Name, id)); err != nil {
		fmt.Printf("发送成功通知失败: %v\n", err)
	}

	return nil
}

// DownloadTorrent 批量下载种子并添加到 Transmission
func DownloadTorrent(ids []string, passkey, endpoint, wechatServer, wechatToken string) error {
	var lastError error

	for i := range ids {
		path := fmt.Sprintf("torrents/%s.torrent", ids[i])

		// 检查文件是否已存在
		if _, err := os.Stat(path); err == nil {
			continue // 文件已存在，跳过
		}

		err := downloadATorrent(ids[i], passkey, path, endpoint, wechatServer, wechatToken)
		if err != nil {
			lastError = err
			// 记录错误但继续处理其他种子
			fmt.Printf("下载种子 %s 失败: %v\n", ids[i], err)
		}
	}

	return lastError
}
