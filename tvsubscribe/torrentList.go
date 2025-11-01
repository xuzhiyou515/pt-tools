package tvsubscribe

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	RES_2160P = iota
	RES_1080P
)

type TVInfo struct {
	DouBanID   string `json:"douban_id"`
	Resolution int    `json:"resolution"`
}

// 搜索链接
// 豆瓣ID 36391902 分辨率 2160P https://springsunday.net/torrents.php?standard1=1&team9=1&incldead=0&spstate=0&pick=0&inclbookmarked=0&search=36391902&search_area=5&search_mode=0
// 豆瓣ID 36391902 分辨率 1080P https://springsunday.net/torrents.php?standard2=1&team9=1&incldead=0&spstate=0&pick=0&inclbookmarked=0&search=36391902&search_area=5&search_mode=0

func QueryTorrentList(cookie string, info *TVInfo) ([]string, error) {
	// 参数校验
	if info == nil {
		return nil, fmt.Errorf("TVInfo 参数不能为空")
	}
	if strings.TrimSpace(info.DouBanID) == "" {
		return nil, fmt.Errorf("豆瓣ID不能为空")
	}
	if strings.TrimSpace(cookie) == "" {
		return nil, fmt.Errorf("Cookie不能为空")
	}

	// 构建搜索URL
	searchURL := buildSearchURL(info)

	// 创建HTTP请求
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Cookie", cookie)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP请求失败，状态码: %d", resp.StatusCode)
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查响应是否为空
	if len(body) == 0 {
		return []string{}, nil
	}

	// 解析种子ID
	torrentIDs := extractTorrentIDs(string(body))

	return torrentIDs, nil
}

// buildSearchURL 根据TVInfo构建搜索URL
func buildSearchURL(info *TVInfo) string {
	baseURL := "https://springsunday.net/torrents.php?"

	// 根据分辨率选择standard参数
	var standardParam string
	switch info.Resolution {
	case RES_2160P:
		standardParam = "standard1=1"
	case RES_1080P:
		standardParam = "standard2=1"
	default:
		standardParam = "standard2=1" // 默认使用1080P
	}

	// 构建完整URL
	url := fmt.Sprintf("%s%s&team9=1&incldead=0&spstate=0&pick=0&inclbookmarked=0&search=%s&search_area=5&search_mode=0",
		baseURL, standardParam, info.DouBanID)

	return url
}

// extractTorrentIDs 从HTML内容中提取种子ID
func extractTorrentIDs(htmlContent string) []string {
	var torrentIDs []string

	if strings.TrimSpace(htmlContent) == "" {
		return torrentIDs
	}

	// 解析HTML文档
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return torrentIDs
	}

	// 查找XPath对应的选择器: #outer > div > table，在其中查找种子链接
	doc.Find("#outer > div > table a[href*='details.php?id']").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			if torrentID := extractTorrentIDFromURL(href); torrentID != "" {
				torrentIDs = append(torrentIDs, torrentID)
			}
		}
	})

	return torrentIDs
}

// extractTorrentIDFromURL 从URL中提取种子ID
func extractTorrentIDFromURL(url string) string {
	// 查找 details.php?id=数字 的模式
	if strings.Contains(url, "details.php?id=") {
		// 分割URL获取ID部分
		parts := strings.Split(url, "details.php?id=")
		if len(parts) > 1 {
			idPart := parts[1]
			// 移除其他参数，只保留数字ID
			if ampIndex := strings.Index(idPart, "&"); ampIndex != -1 {
				idPart = idPart[:ampIndex]
			}
			return strings.TrimSpace(idPart)
		}
	}
	return ""
}

