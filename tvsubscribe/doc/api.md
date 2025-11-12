# TVSubscribe API 文档

## 概述

TVSubscribe 提供完整的 RESTful API 接口，支持配置管理、订阅管理、豆瓣搜索等功能。

**基础信息**
- 基础URL: `http://localhost:8443`
- 响应格式: JSON
- 字符编码: UTF-8

## 通用响应格式

所有API响应都遵循统一格式：

```json
{
  "success": true,
  "message": "操作成功",
  "data": {
    // 具体数据内容
  }
}
```

错误响应格式：
```json
{
  "success": false,
  "message": "错误描述"
}
```

## 配置管理 API

### 获取配置

**请求**
```http
GET /getConfig
```

**响应示例**
```json
{
  "success": true,
  "data": {
    "endpoint": "https://springsunday.net",
    "cookie": "your_cookie_here",
    "interval_minutes": 60,
    "wechat_server": "https://your-wechat-bot.com/webhook",
    "wechat_token": "your_token",
    "port": 8443
  }
}
```

### 设置配置

**请求**
```http
POST /setConfig
Content-Type: application/json

{
  "interval_minutes": 30,
  "wechat_server": "https://new-server.com/webhook"
}
```

**响应**
```json
{
  "success": true,
  "message": "配置更新成功，正在立即处理订阅",
  "data": {
    // 更新后的完整配置
  }
}
```

## 订阅管理 API

### 获取订阅列表

**请求**
```http
GET /getSubscribeList
```

**响应示例**
```json
{
  "success": true,
  "data": [
    {
      "id": "a1b2c3d4e5f6",
      "douban_id": "36391902",
      "name": "庆余年 第二季",
      "resolution": 1
    },
    {
      "id": "b7c8d9e0f1g2",
      "douban_id": "26798436",
      "name": "琅琊榜",
      "resolution": 0
    }
  ]
}
```

### 添加订阅

**请求**
```http
POST /addSubscribe
Content-Type: application/json

{
  "douban_id": "36391902",
  "resolution": 1
}
```

**响应**
```json
{
  "success": true,
  "message": "订阅添加成功，正在立即查询和下载种子",
  "data": {
    "id": "c3d4e5f6g7h8",
    "douban_id": "36391902",
    "name": "庆余年 第二季",
    "resolution": 1
  }
}
```

### 删除订阅

#### 新格式：批量删除（推荐）

**请求**
```http
POST /delSubscribe
Content-Type: application/json

{
  "ids": ["a1b2c3d4e5f6", "b7c8d9e0f1g2"]
}
```

**响应**
```json
{
  "success": true,
  "message": "成功删除 2 个订阅"
}
```

#### 旧格式：单个删除（兼容）

**请求**
```http
POST /delSubscribe
Content-Type: application/json

{
  "douban_id": "36391902",
  "resolution": 1
}
```

**响应**
```json
{
  "success": true,
  "message": "订阅删除成功",
  "data": {
    "douban_id": "36391902",
    "resolution": 1
  }
}
```

### 立即触发订阅处理

**请求**
```http
POST /triggerNow
Content-Type: application/json

{
  "ids": ["a1b2c3d4e5f6", "b7c8d9e0f1g2"]
}
```

**响应**
```json
{
  "success": true,
  "message": "已触发 2 个订阅的处理",
  "data": {
    "triggered_count": 2,
    "total_requested": 2
  }
}
```

## 豆瓣搜索 API

### 搜索电视剧

**请求**
```http
GET /searchDouBan?name=庆余年
```

**响应示例**
```json
{
  "success": true,
  "data": [
    {
      "douban_id": "36391902",
      "title": "庆余年 第二季",
      "img": "https://img9.doubanio.com/view/photo/s_ratio_poster/public/p2881234567.jpg",
      "year": "2024",
      "episode": "33"
    },
    {
      "douban_id": "25853071",
      "title": "庆余年",
      "img": "https://img1.doubanio.com/view/photo/s_ratio_poster/public/p2901234567.jpg",
      "year": "2019",
      "episode": "46"
    }
  ]
}
```

**注意**
- 结果按豆瓣ID降序排列（新内容优先）
- 自动过滤非movie类型的内容
- 支持中文名称搜索

## 图片代理 API

### 获取豆瓣图片

**请求**
```http
GET /proxy/image?url=https://img9.doubanio.com/view/photo/s_ratio_poster/public/p2881234567.jpg
```

**响应**
- 直接返回图片内容
- Content-Type 根据图片格式自动设置
- 缓存时间: 1小时
- 只允许代理豆瓣图片

**用途**
- 避免豆瓣防盗链导致的403错误
- 通过服务器转发，保证图片正常显示

## 健康检查 API

### 服务状态

**请求**
```http
GET /health
```

**响应**
```json
{
  "success": true,
  "message": "服务运行正常"
}
```

## 数据结构说明

### TVInfo 结构

```json
{
  "id": "a1b2c3d4e5f6",           // 订阅唯一标识符（自动生成）
  "douban_id": "36391902",        // 豆瓣电视剧ID
  "name": "庆余年 第二季",         // 电视剧名称（自动获取）
  "resolution": 1                 // 分辨率 (0=2160P, 1=1080P)
}
```

### DoubanSearchResult 结构

```json
{
  "douban_id": "36391902",        // 豆瓣ID
  "title": "庆余年 第二季",         // 中文标题
  "img": "https://...",          // 海报图片URL
  "year": "2024",                // 上映年份
  "episode": "33"                // 集数
}
```

### Config 结构

```json
{
  "endpoint": "https://springsunday.net",  // PT站点地址
  "cookie": "your_cookie",                 // 登录Cookie
  "interval_minutes": 60,                  // 检查间隔（分钟）
  "wechat_server": "https://...",          // 微信通知服务器（可选）
  "wechat_token": "your_token",            // 微信通知Token（可选）
  "port": 8443                             // HTTP服务端口
}
```

## 错误码说明

| HTTP状态码 | 说明 |
|-----------|------|
| 200 | 请求成功 |
| 400 | 请求参数错误 |
| 403 | 无权限访问（如非豆瓣图片代理） |
| 404 | 资源不存在 |
| 409 | 资源冲突（如重复添加订阅） |
| 500 | 服务器内部错误 |
| 502 | 网关错误（如图片获取失败） |

## 使用示例

### JavaScript (Axios)

```javascript
// 获取订阅列表
const response = await axios.get('http://localhost:8443/getSubscribeList');
const subscribes = response.data.data;

// 批量删除订阅
const deleteResponse = await axios.post('http://localhost:8443/delSubscribe', {
  ids: ['a1b2c3d4e5f6', 'b7c8d9e0f1g2']
});

// 立即触发订阅
const triggerResponse = await axios.post('http://localhost:8443/triggerNow', {
  ids: ['a1b2c3d4e5f6']
});

// 豆瓣搜索
const searchResponse = await axios.get('http://localhost:8443/searchDouBan', {
  params: { name: '庆余年' }
});
```

### cURL

```bash
# 获取配置
curl -X GET http://localhost:8443/getConfig

# 添加订阅
curl -X POST http://localhost:8443/addSubscribe \
  -H "Content-Type: application/json" \
  -d '{"douban_id": "36391902", "resolution": 1}'

# 批量删除
curl -X POST http://localhost:8443/delSubscribe \
  -H "Content-Type: application/json" \
  -d '{"ids": ["a1b2c3d4e5f6", "b7c8d9e0f1g2"]}'

# 立即触发
curl -X POST http://localhost:8443/triggerNow \
  -H "Content-Type: application/json" \
  -d '{"ids": ["a1b2c3d4e5f6"]}'

# 豆瓣搜索
curl -X GET "http://localhost:8443/searchDouBan?name=庆余年"
```

## 注意事项

1. **ID字段**：所有订阅都自动生成唯一ID，推荐使用ID进行操作
2. **向后兼容**：删除API同时支持新旧两种格式
3. **异步处理**：添加订阅和立即触发都是异步操作
4. **图片代理**：仅限豆瓣图片使用，避免403错误
5. **错误容错**：批量操作中无效ID会自动跳过
6. **线程安全**：所有API都是线程安全的

## 版本更新

### v2.0 新增功能
- ✅ 唯一ID支持
- ✅ 批量操作（删除/触发）
- ✅ 豆瓣搜索功能
- ✅ 图片代理服务
- ✅ 立即触发功能
- ✅ 向后兼容设计