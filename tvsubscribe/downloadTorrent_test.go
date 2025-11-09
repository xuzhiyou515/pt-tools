package tvsubscribe

import (
	"testing"
)

// TestAddTorrentToTransmission 测试 addTorrentToTransmission 函数
// 注意：这个测试需要实际的 Transmission 服务器和种子文件
func TestAddTorrentToTransmission(t *testing.T) {
	// 由于函数中硬编码了 Transmission 服务器地址和认证信息
	// 并且需要实际的种子文件，这个测试需要真实的运行环境
	//t.Skip("跳过实际测试，需要真实的 Transmission 服务器和种子文件")

	// 如果要运行测试，请确保：
	// 1. Transmission 服务器运行在 http://192.168.2.5:9091
	// 2. 认证信息正确
	// 3. 提供一个有效的种子文件路径
	_, err := addTorrentToTransmission("1.torrent", "")
	if err != nil {
		t.Fatal(err)
	}
}