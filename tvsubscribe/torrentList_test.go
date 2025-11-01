package tvsubscribe

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBuildSearchURL 测试 buildSearchURL 函数
func TestBuildSearchURL(t *testing.T) {
	tests := []struct {
		name     string
		info     *TVInfo
		expected string
	}{
		{
			name: "2160P分辨率URL构建",
			info: &TVInfo{
				DouBanID:   "36391902",
				Resolution: RES_2160P,
			},
			expected: "https://springsunday.net/torrents.php?standard1=1&team9=1&incldead=0&spstate=0&pick=0&inclbookmarked=0&search=36391902&search_area=5&search_mode=0",
		},
		{
			name: "1080P分辨率URL构建",
			info: &TVInfo{
				DouBanID:   "36391902",
				Resolution: RES_1080P,
			},
			expected: "https://springsunday.net/torrents.php?standard2=1&team9=1&incldead=0&spstate=0&pick=0&inclbookmarked=0&search=36391902&search_area=5&search_mode=0",
		},
		{
			name: "默认分辨率URL构建（无效分辨率）",
			info: &TVInfo{
				DouBanID:   "36391902",
				Resolution: 999, // 无效分辨率
			},
			expected: "https://springsunday.net/torrents.php?standard2=1&team9=1&incldead=0&spstate=0&pick=0&inclbookmarked=0&search=36391902&search_area=5&search_mode=0",
		},
		{
			name: "特殊字符豆瓣ID",
			info: &TVInfo{
				DouBanID:   "12345678",
				Resolution: RES_2160P,
			},
			expected: "https://springsunday.net/torrents.php?standard1=1&team9=1&incldead=0&spstate=0&pick=0&inclbookmarked=0&search=12345678&search_area=5&search_mode=0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildSearchURL(tt.info)
			if result != tt.expected {
				t.Errorf("buildSearchURL() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// TestExtractTorrentIDs 测试 extractTorrentIDs 函数
func TestExtractTorrentIDs(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected []string
	}{
		{
			name:     "空HTML内容",
			html:     "",
			expected: []string{},
		},
		{
			name:     "只有空格的HTML内容",
			html:     "   \n\t  ",
			expected: []string{},
		},
		{
			name: "单个种子ID（在outer表格内）",
			html: `<html><body><div id="outer"><div><table><tr><td><a href="details.php?id=123456&hit=1">种子详情</a></td></tr></table></div></div></body></html>`,
			expected: []string{"123456"},
		},
		{
			name: "多个种子ID（在outer表格内）",
			html: `<html><body><div id="outer"><div><table><tr><td><a href="details.php?id=123456&hit=1">种子1</a></td><td><a href="details.php?id=789012&hit=1">种子2</a></td></tr></table></div></div></body></html>`,
			expected: []string{"123456", "789012"},
		},
		{
			name: "重复种子ID去重（在outer表格内）",
			html: `<html><body><div id="outer"><div><table><tr><td><a href="details.php?id=123456&hit=1">种子1</a></td><td><a href="details.php?id=123456&hit=2">种子2</a></td><td><a href="details.php?id=789012&hit=1">种子3</a></td></tr></table></div></div></body></html>`,
			expected: []string{"123456", "789012"},
		},
		{
			name: "复杂HTML内容中的种子ID（在outer表格内）",
			html: `<html><body><div id="outer"><div><table class="torrent-list"><tr><td><a href="details.php?id=111111&hit=1" class="torrent-link">种子A</a></td><td><span>其他内容</span></td><td><a href="details.php?id=222222&hit=1">种子B</a></td></tr></table></div></div></body></html>`,
			expected: []string{"111111", "222222"},
		},
		{
			name: "没有种子ID的HTML（在outer表格内）",
			html: `<html><body><div id="outer"><div><table><tr><td><a href="other.php?id=123">其他链接</a></td></tr></table></div></div></body></html>`,
			expected: []string{},
		},
		{
			name: "边界情况 - ID为空（在outer表格内）",
			html: `<html><body><div id="outer"><div><table><tr><td><a href="details.php?id=&hit=1">空ID</a></td><td><a href="details.php?id=123456&hit=1">正常ID</a></td></tr></table></div></div></body></html>`,
			expected: []string{"123456"},
		},
		{
			name: "包含用户详情链接的HTML（在outer表格外）",
			html: `<html><body>
				<a href="userdetails.php?id=87654">用户详情</a>
				<div id="outer">
					<div>
						<table>
							<tr>
								<td><a href="details.php?id=123456&hit=1">种子1</a></td>
								<td><a href="details.php?id=789012&page=0">种子2</a></td>
							</tr>
						</table>
					</div>
				</div>
			</body></html>`,
			expected: []string{"123456", "789012"},
		},
		{
			name: "混合包含各种非种子链接（只有表格内的被提取）",
			html: `<html><body>
				<a href="userdetails.php?id=87654">用户详情</a>
				<a href="user.php?id=12345">用户页面</a>
				<div id="outer">
					<div>
						<table>
							<tr>
								<td><a href="details.php?id=111111&hit=1">真正的种子</a></td>
								<td><a href="details.php?id=222222&page=1">另一个种子</a></td>
							</tr>
						</table>
					</div>
				</div>
				<a href="sendmessage.php?id=45678">发送消息</a>
				<a href="report.php?id=99999">举报</a>
			</body></html>`,
			expected: []string{"111111", "222222"},
		},
			}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTorrentIDs(tt.html)
			if len(result) != len(tt.expected) {
				t.Errorf("extractTorrentIDs() 返回长度 %d, 期望长度 %d", len(result), len(tt.expected))
				return
			}

			for i, id := range result {
				if id != tt.expected[i] {
					t.Errorf("extractTorrentIDs()[%d] = %v, expected %v", i, id, tt.expected[i])
				}
			}
		})
	}
}

// TestQueryTorrentList_ParameterValidation 测试 QueryTorrentList 参数校验
func TestQueryTorrentList_ParameterValidation(t *testing.T) {
	tests := []struct {
		name    string
		cookie  string
		info    *TVInfo
		wantErr bool
		errMsg  string
	}{
		{
			name:    "TVInfo为空",
			cookie:  "valid_cookie",
			info:    nil,
			wantErr: true,
			errMsg:  "TVInfo 参数不能为空",
		},
		{
			name:    "豆瓣ID为空",
			cookie:  "valid_cookie",
			info:    &TVInfo{DouBanID: "", Resolution: RES_2160P},
			wantErr: true,
			errMsg:  "豆瓣ID不能为空",
		},
		{
			name:    "豆瓣ID只有空格",
			cookie:  "valid_cookie",
			info:    &TVInfo{DouBanID: "   \t  ", Resolution: RES_2160P},
			wantErr: true,
			errMsg:  "豆瓣ID不能为空",
		},
		{
			name:    "Cookie为空",
			cookie:  "",
			info:    &TVInfo{DouBanID: "123456", Resolution: RES_2160P},
			wantErr: true,
			errMsg:  "Cookie不能为空",
		},
		{
			name:    "Cookie只有空格",
			cookie:  "   \t  ",
			info:    &TVInfo{DouBanID: "123456", Resolution: RES_2160P},
			wantErr: true,
			errMsg:  "Cookie不能为空",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := QueryTorrentList(tt.cookie, tt.info)
			if (err != nil) != tt.wantErr {
				t.Errorf("QueryTorrentList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("QueryTorrentList() error = %v, expected to contain %v", err.Error(), tt.errMsg)
			}
		})
	}
}


// TestQueryTorrentList_Success 测试成功的请求处理
func TestQueryTorrentList_Success(t *testing.T) {
	// 模拟HTML响应
	mockHTML := `<html><body>
		<a href="details.php?id=123456&hit=1">种子1</a>
		<a href="details.php?id=789012&hit=1">种子2</a>
		<a href="details.php?id=123456&hit=2">种子1重复</a>
		<a href="other.php?id=111">其他链接</a>
	</body></html>`

	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求头
		if r.Header.Get("Cookie") != "test_cookie" {
			t.Errorf("Cookie header = %v, expected %v", r.Header.Get("Cookie"), "test_cookie")
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("User-Agent header 不能为空")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockHTML))
	}))
	defer server.Close()

	// 测试 extractTorrentIDs 函数
	result := extractTorrentIDs(mockHTML)
	expected := []string{"123456", "789012"}

	if len(result) != len(expected) {
		t.Errorf("结果长度 %d，期望长度 %d", len(result), len(expected))
		return
	}

	for i, id := range result {
		if id != expected[i] {
			t.Errorf("结果[%d] = %v，期望 %v", i, id, expected[i])
		}
	}
}

// TestQueryTorrentList_EmptyResponse 测试空响应处理
func TestQueryTorrentList_EmptyResponse(t *testing.T) {
	// 测试空HTML响应的解析
	emptyHTML := ""
	result := extractTorrentIDs(emptyHTML)

	if len(result) != 0 {
		t.Errorf("期望空结果，但得到 %v", result)
	}

	// 测试只有空格的HTML响应
	spaceHTML := "   \n\t  "
	result = extractTorrentIDs(spaceHTML)

	if len(result) != 0 {
		t.Errorf("期望空结果，但得到 %v", result)
	}
}

// BenchmarkExtractTorrentIDs extractTorrentIDs 函数的基准测试
func BenchmarkExtractTorrentIDs(b *testing.B) {
	// 创建包含大量种子ID的HTML内容
	var htmlBuilder strings.Builder
	htmlBuilder.WriteString("<html><body>")
	for i := 0; i < 1000; i++ {
		htmlBuilder.WriteString(fmt.Sprintf(`<a href="details.php?id=%d&hit=1">种子%d</a>`, i, i))
	}
	htmlBuilder.WriteString("</body></html>")
	html := htmlBuilder.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractTorrentIDs(html)
	}
}


func TestQueryTorrentListInReality(t *testing.T) {
	// 这是一个示例，展示如何使用 QueryTorrentList 函数
	info := &TVInfo{
		DouBanID:   "37484739",
		Resolution: RES_1080P,
	}
	cookie := os.Getenv("SSD_COOKIE")
	torrentIDs, err := QueryTorrentList(cookie, info)
	if err != nil {
	    t.Fatalf("查询失败: %v\n", err)
	}
	assert.Equal(t, []string{"577692", "577598"}, torrentIDs)

}