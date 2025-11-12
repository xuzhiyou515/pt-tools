package tvsubscribe

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

const (
	RES_2160P = iota
	RES_1080P
)

type TVInfo struct {
	ID         string `json:"id"`         // 订阅唯一标识
	DouBanID   string `json:"douban_id"`  // 豆瓣ID
	Name       string `json:"name"`       // 电视剧名称
	Resolution int    `json:"resolution"` // 分辨率
}

type TorrentInfo struct {
	ID           string // 种子id
	Info         string // 种子信息
	DownloadLink string // 种子下载链接
	Volume       string // 种子大小
}

// DoubanSearchResult 豆瓣搜索结果
type DoubanSearchResult struct {
	ID      string `json:"douban_id"` // 豆瓣ID
	Title   string `json:"title"`     // 标题
	Img     string `json:"img"`       // 图片URL
	Year    string `json:"year"`      // 年份
	Episode string `json:"episode"`   // 集数
}

// doubanAPIResponse 豆瓣API原始响应结构
type doubanAPIResponse struct {
	Episode   string `json:"episode"`
	Img       string `json:"img"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Type      string `json:"type"`      // 类型，我们需要筛选type为movie的
	Year      string `json:"year"`
	SubTitle  string `json:"sub_title"`
	ID        string `json:"id"`        // 豆瓣ID
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

// GetTVNameByDouBanID 根据豆瓣ID获取电视剧名称
func GetTVNameByDouBanID(douBanID string) (string, error) {
	if strings.TrimSpace(douBanID) == "" {
		return "", fmt.Errorf("豆瓣ID不能为空")
	}

	// 构建豆瓣API URL
	url := fmt.Sprintf("https://movie.douban.com/subject/%s/", douBanID)

	// 创建HTTP请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求豆瓣失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("豆瓣API请求失败，状态码: %d", resp.StatusCode)
	}

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取豆瓣响应失败: %v", err)
	}

	// 使用goquery解析HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("解析豆瓣HTML失败: %v", err)
	}

	// 尝试多种方式获取标题
	var title string

	// 方法1: 从h1标签获取
	doc.Find("h1 span").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" && title == "" {
			title = text
		}
	})

	// 方法2: 从property="v:itemreviewed"获取
	if title == "" {
		doc.Find("[property=v:itemreviewed]").Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if text != "" {
				title = text
			}
		})
	}

	// 方法3: 从title标签获取并清理
	if title == "" {
		pageTitle := doc.Find("title").Text()
		if strings.Contains(pageTitle, "豆瓣") {
			title = strings.TrimSpace(strings.Replace(pageTitle, "(豆瓣)", "", -1))
			title = strings.TrimSpace(strings.Replace(title, " (豆瓣)", "", -1))
		}
	}

	if title == "" {
		return "", fmt.Errorf("无法从豆瓣页面获取电视剧名称")
	}

	return title, nil
}

// SearchDouBan 搜索豆瓣
func SearchDouBan(name string) ([]DoubanSearchResult, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("搜索名称不能为空")
	}

	// 构建豆瓣API URL
	baseURL := "https://movie.douban.com/j/subject_suggest"
	params := url.Values{}
	params.Add("q", name)
	searchURL := baseURL + "?" + params.Encode()

	// 创建HTTP客户端，设置超时
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 创建请求
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置HTTP头，模拟浏览器访问
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json, text/javascript, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Set("Accept-Encoding", "identity") // 不使用压缩
	req.Header.Set("Referer", "https://movie.douban.com/")
	req.Header.Set("Connection", "keep-alive")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求豆瓣API失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("豆瓣API返回错误状态码: %d", resp.StatusCode)
	}

	// 读取响应体
	var reader io.Reader = resp.Body

	// 检查是否是gzip压缩
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("创建gzip读取器失败: %v", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %v", err)
	}

	// 记录原始响应用于调试
	responseStr := string(body)
	log.Printf("豆瓣API响应长度: %d", len(responseStr))

	// 清理响应
	cleanedBody := strings.TrimSpace(responseStr)

	// 检查JSON格式是否有效
	if !strings.HasPrefix(cleanedBody, "[") {
		return nil, fmt.Errorf("响应不是有效的JSON格式，期望数组格式。前200字符: %s", cleanedBody[:min(200, len(cleanedBody))])
	}

	// 解析JSON响应
	var apiResponses []doubanAPIResponse
	if err := json.Unmarshal([]byte(cleanedBody), &apiResponses); err != nil {
		return nil, fmt.Errorf("解析JSON响应失败: %v，响应内容前200字符: %s", err, cleanedBody[:min(200, len(cleanedBody))])
	}

// 筛选type为movie的条目并转换为返回格式
	var results []DoubanSearchResult
	for _, item := range apiResponses {
		if item.Type == "movie" && item.ID != "" && item.Title != "" {
			result := DoubanSearchResult{
				ID:      item.ID,
				Title:   item.Title,
				Img:     item.Img,
				Year:    item.Year,
				Episode: item.Episode,
			}
			results = append(results, result)
		}
	}

	// 按豆瓣ID排序（按数字大小降序排列，新ID在前）
	sort.Slice(results, func(i, j int) bool {
		// 尝试将ID转换为数字进行比较
		id1, err1 := strconv.Atoi(results[i].ID)
		id2, err2 := strconv.Atoi(results[j].ID)

		if err1 == nil && err2 == nil {
			// 两个都是数字，按数字降序排列
			return id1 > id2
		} else if err1 == nil {
			// 只有第一个是数字，数字ID排在前面
			return true
		} else if err2 == nil {
			// 只有第二个是数字，字符串ID排在后面
			return false
		} else {
			// 两个都不是数字，按字符串降序排列
			return results[i].ID > results[j].ID
		}
	})

	return results, nil
}

