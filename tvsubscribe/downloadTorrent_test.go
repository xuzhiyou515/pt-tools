package tvsubscribe

import (
	"testing"
)

// TestBuildDownloadURL 测试 buildDownloadURL 函数
func TestBuildDownloadURL(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		passkey  string
		expected string
	}{
		{
			name:     "正常的ID和passkey",
			id:       "577692",
			passkey:  "123456",
			expected: "https://springsunday.net/download.php?id=577692&passkey=123456&https=1",
		},
		{
			name:     "空ID",
			id:       "",
			passkey:  "123456",
			expected: "",
		},
		{
			name:     "空passkey",
			id:       "577692",
			passkey:  "",
			expected: "",
		},
		{
			name:     "都为空",
			id:       "",
			passkey:  "",
			expected: "",
		},
		{
			name:     "长ID和passkey",
			id:       "1234567890",
			passkey:  "abcdefghijk123456789",
			expected: "https://springsunday.net/download.php?id=1234567890&passkey=abcdefghijk123456789&https=1",
		},
		{
			name:     "包含特殊字符的passkey",
			id:       "577692",
			passkey:  "abc-123_def/456",
			expected: "https://springsunday.net/download.php?id=577692&passkey=abc-123_def/456&https=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDownloadURL(tt.id, tt.passkey)
			if result != tt.expected {
				t.Errorf("buildDownloadURL(%q, %q) = %q, expected %q", tt.id, tt.passkey, result, tt.expected)
			}
		})
	}
}

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

// ExampleBuildDownloadURL buildDownloadURL 函数的示例测试
func ExampleBuildDownloadURL() {
	// 示例：构建种子下载链接
	url := buildDownloadURL("577692", "123456")

	// 输出: https://springsunday.net/download.php?id=577692&passkey=123456&https=1
	_ = url
}
