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

// TestExtractTorrentInfos 测试 extractTorrentInfos 函数
func TestExtractTorrentInfos(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected []TorrentInfo
	}{
		{
			name:     "空HTML内容",
			html:     "",
			expected: []TorrentInfo{},
		},
		{
			name:     "只有空格的HTML内容",
			html:     "   \n\t  ",
			expected: []TorrentInfo{},
		},
		{
			name: "单个种子信息（在outer表格内）",
			html: `<html><body><div id="outer"><div><table><tr><td class="embedded"><div class="torrent-title"><a href="details.php?id=123456&hit=1">Sword.and.Beloved.S01E26-E27.2025.2160p</a></div><div class="torrent-smalldescr"><span title="天地剑心 / 狐妖小红娘·王权篇 / 狐妖小红娘之王权篇 | 第26-27集 | 成毅 / 李一桐 / 郭俊辰 [国语] [简繁英字幕]">详细信息</span></div></td><td width="110"><a href="download.php?id=123456&passkey=test">下载</a></td><td>1.5 GB</td></tr></table></div></div></body></html>`,
			expected: []TorrentInfo{
				{ID: "123456", Info: "天地剑心 / 狐妖小红娘·王权篇 / 狐妖小红娘之王权篇 | 第26-27集 | 成毅 / 李一桐 / 郭俊辰 [国语] [简繁英字幕]", DownloadLink: "https://springsunday.net/download.php?id=123456&passkey=test", Volume: "1.5 GB"},
			},
		},
		{
			name: "多个种子信息（在outer表格内）",
			html: `<html><body><div id="outer"><div><table>
				<tr><td class="embedded"><div class="torrent-title"><a href="details.php?id=123456&hit=1">Sword.and.Beloved.S01E26-E27.2025.2160p</a></div><div class="torrent-smalldescr"><span title="天地剑心 / 狐妖小红娘·王权篇 / 狐妖小红娘之王权篇 | 第26-27集 | 成毅 / 李一桐 / 郭俊辰 [国语] [简繁英字幕]">详细信息</span></div></td><td width="110"><a href="download.php?id=123456&passkey=test">下载</a></td><td>1.5 GB</td></tr>
				<tr><td class="embedded"><div class="torrent-title"><a href="details.php?id=789012&hit=1">Sword.and.Beloved.S01E24-E25.2025.2160p</a></div><div class="torrent-smalldescr"><span title="天地剑心 / 狐妖小红娘·王权篇 / 狐妖小红娘之王权篇 | 第24-25集 | 成毅 / 李一桐 / 郭俊辰 [国语] [简繁英字幕]">详细信息</span></div></td><td width="110"><a href="download.php?id=789012&passkey=test">下载</a></td><td>2.1 GB</td></tr>
			</table></div></div></body></html>`,
			expected: []TorrentInfo{
				{ID: "123456", Info: "天地剑心 / 狐妖小红娘·王权篇 / 狐妖小红娘之王权篇 | 第26-27集 | 成毅 / 李一桐 / 郭俊辰 [国语] [简繁英字幕]", DownloadLink: "https://springsunday.net/download.php?id=123456&passkey=test", Volume: "1.5 GB"},
				{ID: "789012", Info: "天地剑心 / 狐妖小红娘·王权篇 / 狐妖小红娘之王权篇 | 第24-25集 | 成毅 / 李一桐 / 郭俊辰 [国语] [简繁英字幕]", DownloadLink: "https://springsunday.net/download.php?id=789012&passkey=test", Volume: "2.1 GB"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTorrentInfos(tt.html)
			if len(result) != len(tt.expected) {
				t.Errorf("extractTorrentInfos() 返回长度 %d, 期望长度 %d", len(result), len(tt.expected))
				return
			}

			for i, info := range result {
				if info.Info != tt.expected[i].Info {
					t.Errorf("extractTorrentInfos()[%d].Info = %v, expected %v", i, info.Info, tt.expected[i].Info)
				}
				if info.DownloadLink != tt.expected[i].DownloadLink {
					t.Errorf("extractTorrentInfos()[%d].DownloadLink = %v, expected %v", i, info.DownloadLink, tt.expected[i].DownloadLink)
				}
				if info.Volume != tt.expected[i].Volume {
					t.Errorf("extractTorrentInfos()[%d].Volume = %v, expected %v", i, info.Volume, tt.expected[i].Volume)
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
	mockHTML := `<html><body><div id="outer"><div><table>
		<tr><td><a href="details.php?id=123456&hit=1">种子1</a></td></tr>
		<tr><td><a href="details.php?id=789012&hit=1">种子2</a></td></tr>
		<tr><td><a href="details.php?id=123456&hit=2">种子1重复</a></td></tr>
		<tr><td><a href="other.php?id=111">其他链接</a></td></tr>
	</table></div></div></body></html>`

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

	// 测试 extractTorrentInfos 函数
	result := extractTorrentInfos(mockHTML)
	expected := []string{"123456", "789012"}

	if len(result) != len(expected) {
		t.Errorf("结果长度 %d，期望长度 %d", len(result), len(expected))
		return
	}

	for i, info := range result {
		if info.ID != expected[i] {
			t.Errorf("结果[%d].ID = %v，期望 %v", i, info.ID, expected[i])
		}
	}
}

// TestQueryTorrentList_EmptyResponse 测试空响应处理
func TestQueryTorrentList_EmptyResponse(t *testing.T) {
	// 测试空HTML响应的解析
	emptyHTML := ""
	result := extractTorrentInfos(emptyHTML)

	if len(result) != 0 {
		t.Errorf("期望空结果，但得到 %v", result)
	}

	// 测试只有空格的HTML响应
	spaceHTML := "   \n\t  "
	result = extractTorrentInfos(spaceHTML)

	if len(result) != 0 {
		t.Errorf("期望空结果，但得到 %v", result)
	}
}

// BenchmarkExtractTorrentInfos extractTorrentInfos 函数的基准测试
func BenchmarkExtractTorrentInfos(b *testing.B) {
	// 创建包含大量种子ID的HTML内容
	var htmlBuilder strings.Builder
	htmlBuilder.WriteString("<html><body><div id=\"outer\"><div><table>")
	for i := 0; i < 1000; i++ {
		htmlBuilder.WriteString(fmt.Sprintf(`<tr><td><a href="details.php?id=%d&hit=1">种子%d</a></td></tr>`, i, i))
	}
	htmlBuilder.WriteString("</table></div></div></body></html>")
	html := htmlBuilder.String()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractTorrentInfos(html)
	}
}


func TestQueryTorrentListInReality(t *testing.T) {
	// 这是一个示例，展示如何使用 QueryTorrentList 函数
	info := &TVInfo{
		DouBanID:   "37484739",
		Resolution: RES_1080P,
	}
	cookie := os.Getenv("SSD_COOKIE")
	torrentInfos, err := QueryTorrentList(cookie, info)
	if err != nil {
	    t.Fatalf("查询失败: %v\n", err)
	}
	expectedIDs := []string{"579091", "577692", "577598"}
	actualIDs := make([]string, len(torrentInfos))
	for i, info := range torrentInfos {
		actualIDs[i] = info.ID
	}
	assert.Equal(t, expectedIDs, actualIDs)
	fmt.Printf("torrentInfos: %v", torrentInfos)
}