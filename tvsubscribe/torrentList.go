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

type TorrentInfo struct {
	ID           string // 种子id
	Info         string // 种子信息
	DownloadLink string // 种子下载链接
	Volume       string // 种子大小
}

// 搜索链接
// 豆瓣ID 36391902 分辨率 2160P https://springsunday.net/torrents.php?standard1=1&team9=1&incldead=0&spstate=0&pick=0&inclbookmarked=0&search=36391902&search_area=5&search_mode=0
// 豆瓣ID 36391902 分辨率 1080P https://springsunday.net/torrents.php?standard2=1&team9=1&incldead=0&spstate=0&pick=0&inclbookmarked=0&search=36391902&search_area=5&search_mode=0

func QueryTorrentList(cookie string, info *TVInfo) ([]TorrentInfo, error) {
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
		return []TorrentInfo{}, nil
	}

	// 解析种子信息
	torrentInfos := extractTorrentInfos(string(body))

	return torrentInfos, nil
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

// extractTorrentInfos 从HTML内容中提取种子详细信息
func extractTorrentInfos(htmlContent string) []TorrentInfo {
	if strings.TrimSpace(htmlContent) == "" {
		return []TorrentInfo{}
	}

	// 解析HTML文档
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return []TorrentInfo{}
	}

	// 使用map来去重
	uniqueIDs := make(map[string]bool)
	var torrentInfos []TorrentInfo

	// 查找种子列表行：#outer > div > table 中的每一行
	doc.Find("#outer > div > table tr").Each(func(i int, s *goquery.Selection) {
		// 查找种子详情链接
		detailLink := s.Find("a[href*='details.php?id']")
		if detailLink.Length() == 0 {
			return // 跳过没有详情链接的行
		}

		href, exists := detailLink.Attr("href")
		if !exists {
			return
		}

		torrentID := extractTorrentIDFromURL(href)
		if torrentID == "" {
			return
		}

		// 如果ID还没有出现过，则添加到结果中
		if !uniqueIDs[torrentID] {
			uniqueIDs[torrentID] = true

			// 提取种子信息：在 torrent-smalldescr div 中查找具有 title 属性的 span 标签
			info := ""
			s.Find(".torrent-smalldescr span[title]").Each(func(j int, span *goquery.Selection) {
				title, exists := span.Attr("title")
				if exists && title != "" {
					// 选择最长的 title（通常是详细的描述）
					if len(title) > len(info) {
						info = title
					}
				}
			})

			// 提取下载链接：在当前行中查找包含 download.php 的链接
			downloadLink := ""
			downloadLinks := s.Find("a[href*='download.php']")
			downloadLinks.Each(func(j int, a *goquery.Selection) {
				if href, exists := a.Attr("href"); exists && href != "" && downloadLink == "" {
					downloadLink = "https://springsunday.net/" + href
				}
			})

			// 提取种子大小：在当前行的 td 中查找包含大小单位的文本
			volume := ""
			s.Find("td").Each(func(j int, td *goquery.Selection) {
				tdText := strings.TrimSpace(td.Text())
				// 检查是否包含大小信息（通常包含 GB、MB 等单位）
				if strings.Contains(tdText, "GB") || strings.Contains(tdText, "MB") || strings.Contains(tdText, "KB") {
					volume = strings.ReplaceAll(tdText, "<br>", " ")
					volume = strings.TrimSpace(volume)
				}
			})

			torrentInfo := TorrentInfo{
				ID:           torrentID,
				Info:         info,
				DownloadLink: downloadLink,
				Volume:       volume,
			}

			torrentInfos = append(torrentInfos, torrentInfo)
		}
	})

	return torrentInfos
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

