package server

import (
	"fmt"
	"log"
	"net/http"

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
		   path == "/health" {
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

// delSubscribe 删除订阅
func (s *Server) delSubscribe(c *gin.Context) {
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