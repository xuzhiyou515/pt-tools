package tvsubscribe

import (
	"context"
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
func addTorrentToTransmission(torrentPath, endpoint string) error {
	// Transmission RPC 配置
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("解析 Transmission 端点失败: %v", err)
	}

	// 创建 Transmission 客户端
	client, err := transmissionrpc.New(endpointURL, nil)
	if err != nil {
		return fmt.Errorf("创建 Transmission 客户端失败: %v", err)
	}
	torrent, err := client.TorrentAddFile(context.TODO(), torrentPath)
	if err != nil {
		return fmt.Errorf("添加种子文件失败: %v", err)
	}
	fmt.Printf("添加%v种子文件成功\n",torrent.Name)
	return nil
}

// downloadATorrent 下载单个种子文件并添加到 Transmission
func downloadATorrent(id, passkey, path, endpoint string) error {
	// 下载种子文件
	err := downloadFile(buildDownloadURL(id, passkey), path)
	if err != nil {
		// 删除可能已创建的不完整文件
		os.Remove(path)
		return fmt.Errorf("下载种子文件失败: %v", err)
	}

	// 添加到 Transmission
	err = addTorrentToTransmission(path, endpoint)
	if err != nil {
		// 删除种子文件
		os.Remove(path)
		return fmt.Errorf("添加种子到 Transmission 失败: %v", err)
	}

	return nil
}

// DownloadTorrent 批量下载种子并添加到 Transmission
func DownloadTorrent(ids []string, passkey, endpoint string) error {
	var lastError error

	for i := range ids {
		path := fmt.Sprintf("torrents/%s.torrent", ids[i])

		// 检查文件是否已存在
		if _, err := os.Stat(path); err == nil {
			continue // 文件已存在，跳过
		}

		err := downloadATorrent(ids[i], passkey, path, endpoint)
		if err != nil {
			lastError = err
			// 记录错误但继续处理其他种子
			fmt.Printf("下载种子 %s 失败: %v\n", ids[i], err)
		}
	}

	return lastError
}
