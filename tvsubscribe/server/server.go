package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"tvsubscribe"
	"tvsubscribe/interfaces"
)

// ProcessTVSubscribesFunc 处理电视剧订阅的函数类型
type ProcessTVSubscribesFunc func()

// ProcessSingleTVFunc 处理单个电视剧的函数类型
type ProcessSingleTVFunc func(tvInfo tvsubscribe.TVInfo)

// Server HTTP服务器
type Server struct {
	configManager       interfaces.ConfigManager
	subscribeManager    interfaces.SubscribeManager
	engine              *gin.Engine
	processTVSubscribes ProcessTVSubscribesFunc
	processSingleTV     ProcessSingleTVFunc
}

// NewServer 创建新的HTTP服务器
func NewServer(configManager interfaces.ConfigManager, subscribeManager interfaces.SubscribeManager, processTVSubscribes ProcessTVSubscribesFunc, processSingleTV ProcessSingleTVFunc) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	server := &Server{
		configManager:       configManager,
		subscribeManager:    subscribeManager,
		engine:              engine,
		processTVSubscribes: processTVSubscribes,
		processSingleTV:     processSingleTV,
	}

	server.setupRoutes()
	return server
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 静态文件服务 - 如果存在构建文件则提供Web界面
	s.engine.Static("/assets", "./web/dist/assets")
	s.engine.StaticFile("/", "./web/dist/index.html")

	// 处理SPA路由 - 所有未匹配的路由都返回index.html
	s.engine.NoRoute(func(c *gin.Context) {
		// 如果是API请求，返回404
		path := c.Request.URL.Path
		if path == "/getConfig" ||
		   path == "/setConfig" ||
		   path == "/getSubscribeList" ||
		   path == "/addSubscribe" ||
		   path == "/delSubscribe" ||
		   path == "/triggerNow" ||
		   path == "/searchDouBan" ||
		   path == "/health" ||
		   path == "/proxy/image" {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "API接口不存在",
			})
			return
		}
		// 对于其他路径，返回index.html以支持Vue Router
		c.File("./web/dist/index.html")
	})

	// 配置相关API
	s.engine.GET("/getConfig", s.getConfig)
	s.engine.POST("/setConfig", s.setConfig)

	// 订阅相关API
	s.engine.GET("/getSubscribeList", s.getSubscribeList)
	s.engine.POST("/addSubscribe", s.addSubscribe)
	s.engine.POST("/delSubscribe", s.delSubscribe)
	s.engine.POST("/triggerNow", s.triggerNow)

	// 豆瓣搜索
	s.engine.GET("/searchDouBan", s.searchDouBan)

	// 豆瓣图片代理
	s.engine.GET("/proxy/image", s.proxyImage)

	// 健康检查
	s.engine.GET("/health", s.health)
}

// getConfig 获取配置
func (s *Server) getConfig(c *gin.Context) {
	configInterface := s.configManager.GetConfig()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    configInterface,
	})
}

// setConfig 设置配置
func (s *Server) setConfig(c *gin.Context) {
	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的JSON格式: " + err.Error(),
		})
		return
	}

	// 更新配置
	if err := s.configManager.UpdateConfig(updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 配置更新成功后，立即执行一次电视剧订阅处理
	go func() {
		log.Println("配置已更新，立即执行电视剧订阅处理")
		s.processTVSubscribes()
	}()

	// 返回更新后的配置
	updatedConfig := s.configManager.GetConfig()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "配置更新成功，正在立即处理订阅",
		"data":    updatedConfig,
	})
}

// getSubscribeList 获取订阅列表
func (s *Server) getSubscribeList(c *gin.Context) {
	subscribes := s.subscribeManager.GetSubscribes()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    subscribes,
	})
}

// addSubscribe 添加订阅
func (s *Server) addSubscribe(c *gin.Context) {
	var tvInfo tvsubscribe.TVInfo
	if err := c.ShouldBindJSON(&tvInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的JSON格式: " + err.Error(),
		})
		return
	}

	// 验证必填字段
	if tvInfo.DouBanID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "豆瓣ID不能为空",
		})
		return
	}
	if tvInfo.Resolution <= 0 {
		tvInfo.Resolution = 1 // 默认分辨率
	}

	// 添加订阅
	if err := s.subscribeManager.AddSubscribe(tvInfo); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 订阅添加成功后，立即查询和下载该订阅的种子
	go func() {
		log.Printf("新订阅添加成功，立即处理豆瓣ID: %s, 分辨率: %d", tvInfo.DouBanID, tvInfo.Resolution)
		s.processSingleTV(tvInfo)
	}()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "订阅添加成功，正在立即查询和下载种子",
		"data":    tvInfo,
	})
}

// delSubscribe 删除订阅（支持两种模式：旧版兼容和新版ID数组）
func (s *Server) delSubscribe(c *gin.Context) {
	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的JSON格式: " + err.Error(),
		})
		return
	}

	// 检查是新版本ID数组格式还是旧版本TVInfo格式
	if ids, ok := requestBody["ids"]; ok {
		// 新版本：ID数组格式
		idArray, ok := ids.([]interface{})
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "ids字段必须是数组格式",
			})
			return
		}

		// 转换为字符串数组
		var idsToDelete []string
		for _, id := range idArray {
			if idStr, ok := id.(string); ok {
				if idStr != "" {
					idsToDelete = append(idsToDelete, idStr)
				}
			}
		}

		if len(idsToDelete) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "ID列表不能为空",
			})
			return
		}

		// 批量删除
		if err := s.subscribeManager.RemoveSubscribesByID(idsToDelete); err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": fmt.Sprintf("成功删除 %d 个订阅", len(idsToDelete)),
		})

	} else {
		// 旧版本：TVInfo格式（向后兼容）
		var tvInfo tvsubscribe.TVInfo
		if err := c.ShouldBindJSON(&tvInfo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "无效的JSON格式: " + err.Error(),
			})
			return
		}

		// 验证必填字段
		if tvInfo.DouBanID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "豆瓣ID不能为空",
			})
			return
		}
		if tvInfo.Resolution <= 0 {
			tvInfo.Resolution = 1 // 默认分辨率
		}

		// 删除订阅
		if err := s.subscribeManager.RemoveSubscribe(tvInfo); err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "订阅删除成功",
			"data":    tvInfo,
		})
	}
}

// triggerNow 立即触发订阅处理
func (s *Server) triggerNow(c *gin.Context) {
	var requestBody map[string]interface{}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的JSON格式: " + err.Error(),
		})
		return
	}

	// 检查ids字段
	idsInterface, ok := requestBody["ids"]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "缺少ids字段",
		})
		return
	}

	// 转换为字符串数组
	idArray, ok := idsInterface.([]interface{})
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "ids字段必须是数组格式",
		})
		return
	}

	var idsToTrigger []string
	for _, id := range idArray {
		if idStr, ok := id.(string); ok {
			if idStr != "" {
				idsToTrigger = append(idsToTrigger, idStr)
			}
		}
	}

	if len(idsToTrigger) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "ID列表不能为空",
		})
		return
	}

	// 获取要触发的订阅信息
	var subscribesToTrigger []tvsubscribe.TVInfo
	for _, id := range idsToTrigger {
		subscribe, err := s.subscribeManager.GetSubscribeByID(id)
		if err != nil {
			// 记录错误但继续处理其他订阅
			log.Printf("获取订阅失败 ID=%s: %v", id, err)
			continue
		}
		subscribesToTrigger = append(subscribesToTrigger, subscribe)
	}

	if len(subscribesToTrigger) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "未找到有效的订阅",
		})
		return
	}

	// 异步触发订阅处理
	go func() {
		log.Printf("开始处理 %d 个立即触发的订阅", len(subscribesToTrigger))
		for _, subscribe := range subscribesToTrigger {
			log.Printf("处理订阅 ID=%s, 豆瓣ID=%s, 分辨率=%d",
				subscribe.ID, subscribe.DouBanID, subscribe.Resolution)
			s.processSingleTV(subscribe)
		}
		log.Printf("完成处理 %d 个立即触发的订阅", len(subscribesToTrigger))
	}()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("已触发 %d 个订阅的处理", len(subscribesToTrigger)),
		"data": gin.H{
			"triggered_count": len(subscribesToTrigger),
			"total_requested": len(idsToTrigger),
		},
	})
}

// searchDouBan 搜索豆瓣
func (s *Server) searchDouBan(c *gin.Context) {
	// 获取查询参数
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "搜索关键词不能为空，请提供name参数",
		})
		return
	}

	// 调用豆瓣搜索功能
	results, err := tvsubscribe.SearchDouBan(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "搜索豆瓣失败: " + err.Error(),
		})
		return
	}

	// 返回搜索结果
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
	})
}

// proxyImage 图片代理
func (s *Server) proxyImage(c *gin.Context) {
	imageURL := c.Query("url")
	if imageURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请提供图片URL参数",
		})
		return
	}

	// 只允许代理豆瓣的图片
	if !strings.HasPrefix(imageURL, "https://img") || !strings.Contains(imageURL, "doubanio.com") {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "只允许代理豆瓣图片",
		})
		return
	}

	// 创建HTTP客户端，设置超时
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 创建请求
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "创建图片请求失败",
		})
		return
	}

	// 设置请求头，模拟浏览器
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://movie.douban.com/")
	req.Header.Set("Accept", "image/webp,image/apng,image/*,*/*;q=0.8")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取图片失败",
		})
		return
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{
			"success": false,
			"message": fmt.Sprintf("图片服务器返回错误: %d", resp.StatusCode),
		})
		return
	}

	// 获取Content-Type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	// 设置响应头
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=3600") // 缓存1小时
	c.Header("Access-Control-Allow-Origin", "*")

	// 直接返回图片数据
	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		// 如果复制失败，可能是因为客户端已经断开连接
		return
	}
}

// health 健康检查
func (s *Server) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "服务运行正常",
	})
}

// Start 启动服务器
func (s *Server) Start() error {
	configMap := s.configManager.GetConfig()
	port := 8443 // 默认端口
	if p, ok := configMap["port"].(int); ok {
		port = p
	} else if p, ok := configMap["port"].(float64); ok {
		port = int(p)
	}
	addr := fmt.Sprintf(":%d", port)
	log.Printf("HTTP服务器启动在端口 %d", port)
	return s.engine.Run(addr)
}

// GetEngine 获取gin引擎（用于测试）
func (s *Server) GetEngine() *gin.Engine {
	return s.engine
}