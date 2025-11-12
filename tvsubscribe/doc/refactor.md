# TVSubscribe 重构文档

## 重构目标

将原有的种子处理逻辑从基于 ID 的简单模式升级为基于完整信息的 TorrentInfo 结构化模式，提供更丰富的种子信息和更可靠的下载机制，并实现现代化的管理界面和API服务。

## 重构内容

### 1. 新增 TorrentInfo 结构体

在 `tvsubscribe/torrentList.go` 中添加以下结构：

```go
type TorrentInfo struct {
    ID           string // 种子id，原QueryTorrentList的返回值
    Info         string // 种子详细信息，从.torrent-smalldescr span[title]的title属性提取
    DownloadLink string // 种子下载链接，直接从页面的download.php链接获取
    Volume       string // 种子大小，从包含GB/MB/KB单位的td内容提取
}
```

### 2. 修改 QueryTorrentList 函数

- **返回值修改**：从 `([]string, error)` 改为 `([]TorrentInfo, error)`
- **新增 extractTorrentInfos 函数**：替换原有的 extractTorrentIDs 函数，提供完整的种子信息提取
- **信息提取逻辑改进**：
  - ID：从 `details.php?id=` 链接中提取种子ID
  - Info：从 `.torrent-smalldescr span[title]` 的 title 属性中提取详细的中文描述
  - DownloadLink：从页面的 `download.php` 链接中提取完整下载地址
  - Volume：从包含大小单位的 td 中提取种子大小信息

### 3. 重构下载功能

在 `tvsubscribe/downloadTorrent.go` 中：

- **移除 buildDownloadURL 函数**：不再需要通过ID和passkey构建下载链接
- **修改 DownloadTorrent 函数**：参数从 `[]string, passkey` 改为 `[]TorrentInfo`
- **新增 downloadATorrentFromInfo 函数**：使用 TorrentInfo 结构体进行下载
- **直接使用下载链接**：优先使用页面上的完整下载链接，提高可靠性
- **移除 passkey 依赖**：简化配置和参数传递

### 4. 调整配置文件

- 在 `tvsubscribe/cmd/main.go` 中 Config 移除 Subscribes
- Subscribes 使用独立的变量，从 `./subscribes.json` 中读取数据
- Config 添加 Port 字段，用于配置监听的端口，默认监听8443端口

### 5、优化配置和订阅调整

使用gin框架，支持以下http接口来获取、调整配置和订阅：

- GET `/getConfig` 返回 Config json 结构
- POST `/setConfig` payload 是 Config 结构（未包含的字段不修改）
  - 更新内存的Config数据，并把更新后的Config写入配置文件
- GET `/getSubscribeList` 返回 Subcribes 列表
- POST `/addSubscribe` payload 是 tvsubscribe.TVInfo
  - 添加新的订阅，需要注意去重, 更新配置文件
- POST `/delSubscribe` payload 是 tvsubscribe.TVInfo
  - 删除订阅，更新配置文件

支持通过命令行来调整配置、订阅,实现方式是调用对应的http接口：

```
tvsubscribe config --list 获取配置
tvsubscribe config --set "endpoint=xxxx" "cookie=xxxx" "interval_minutes=5" "wechat_server=xxxx" "wechat_token=xxxx" 设置配置（配置时不需要每项都设置）
tvsubscribe subscribe --list 获取订阅
tvsubscribe subscribe --add "douban_id=xxxxx" "resolution=1" 添加订阅 （resolution可以省略，默认是1）
tvsubscribe subscribe --del "douban_id=xxxxx" "resolution=1" 删除订阅 （resolution可以省略，默认是1）
```

以上命令都支持通过 --url 配置服务器地址 默认是 "127.0.0.1:8443"

使用现有满足需求的库实现命令行的解析，比如 flag

### 6、实现管理页面

使用vue框架调用5中实现的http接口实现简单的管理界面

### 7、添加搜索豆瓣功能

tvsubscribe后台实现豆瓣搜素http接口：

```
GET /searchDouBan?name=xxxxx
```

返回结果结构 如下示例：

```json
[
  {
    "douban_id": "xxxxxx",
    "title": "xxxxxx",
    "img": "xxxx",
    "year": "2025",
    "episode": "13"
  },
  {
    "douban_id": "xxxxxx",
    "title": "xxxxxx",
    "img": "xxxx",
    "year": "2025",
    "episode": "13"
  }
]
```

可以通过豆瓣的api实现：

请求url格式：`https://movie.douban.com/j/subject_suggest?q=%E9%97%B4%E8%B0%8d`

请求结果：

```json
[
    {
        "episode": "13",
        "img": "https://img9.doubanio.com\/view\/photo\/s_ratio_poster\/public\/p2924788556.jpg",
        "title": "间谍过家家 第三季",
        "url": "https:\/\/movie.douban.com\/subject\/36700483\/?suggest=%E9%97%B4%E8%B0%8D",
        "type": "movie",
        "year": "2025",
        "sub_title": "SPY×FAMILY Season 3",
        "id": "36700483"
    },
    {
        "episode": "12",
        "img": "https://img1.doubanio.com\/view\/photo\/s_ratio_poster\/public\/p2869306649.jpg",
        "title": "间谍过家家 第一季",
        "url": "https:\/\/movie.douban.com\/subject\/35258427\/?suggest=%E9%97%B4%E8%B0%8D",
        "type": "movie",
        "year": "2022",
        "sub_title": "SPY×FAMILY",
        "id": "35258427"
    },
    {
        "episode": "12",
        "img": "https://img3.doubanio.com\/view\/photo\/s_ratio_poster\/public\/p2899072942.jpg",
        "title": "间谍过家家 第二季",
        "url": "https:\/\/movie.douban.com\/subject\/36190888\/?suggest=%E9%97%B4%E8%B0%8d",
        "type": "movie",
        "year": "2023",
        "sub_title": "SPY×FAMILY Season 2",
        "id": "36190888"
    }
]
```

需要筛选出type为movie的，避免混入书籍、音乐

### 8、优化订阅管理

TVInfo添加ID字段，在添加订阅时赋予一个唯一id，兼容处理在加载订阅配置文件时如果没有id也重新赋予一个唯一id

调整delSubscribe http 接口 payload 为 id数组

管理网页的删除、批量删除都按照调整后的逻辑处理

添加立即触发功能 http 接口：

```
POST /triggerNow payload 为 id数组
```

管理网页添加对应的立即触发、批量立即触发功能

## 重构实现状态

### ✅ 已完成的功能

1. **TorrentInfo 结构体** - 完整的种子信息结构
   - ID: 种子唯一标识
   - Info: 种子详细信息
   - DownloadLink: 直接下载链接
   - Volume: 种子大小

2. **种子查询重构** - 基于TorrentInfo的完整信息提取
   - 重构QueryTorrentList函数返回值类型
   - 新增extractTorrentInfos函数
   - 改进HTML解析逻辑，支持更多种子信息

3. **下载功能重构** - 基于TorrentInfo的可靠下载
   - 移除buildDownloadURL函数
   - 新增downloadATorrentFromInfo函数
   - 直接使用页面上的完整下载链接

4. **配置管理重构** - 独立的配置和订阅管理
   - Config和Subscribes分离存储
   - 新增Port字段用于HTTP服务
   - 实现配置管理器接口

5. **HTTP API服务** - 基于Gin框架的完整API
   - 配置管理API (getConfig/setConfig)
   - 订阅管理API (getSubscribeList/addSubscribe/delSubscribe)
   - 健康检查API (health)

6. **CLI命令行工具** - 完整的命令行管理功能
   - 配置管理命令 (config --list/--set)
   - 订阅管理命令 (subscribe --list/--add/--del)
   - 远程服务器支持 (--url参数)

7. **Vue管理界面** - 现代化Web管理界面
   - 基于Vue 3 + Element Plus
   - 配置管理页面
   - 订阅管理页面（支持多选批量删除）
   - 响应式设计

8. **静态文件集成** - 内置Web服务
   - 后端直接host Vue管理界面
   - SPA路由支持
   - 单端口部署 (默认8443)

9. **TVInfo增强** - 智能名称获取
   - 新增Name字段
   - 自动从豆瓣获取电视剧名称
   - 支持多种解析策略

10. **微信通知功能** - 实时状态推送
    - 种子下载成功/失败通知
    - 配置微信服务器和Token
    - 异步通知发送

11. **豆瓣搜索功能** - 智能订阅添加
    - HTTP API: GET /searchDouBan?name=xxx
    - 支持电视剧名称搜索
    - 返回豆瓣ID、标题、海报、年份、集数
    - 结果按豆瓣ID排序（新内容优先）
    - 图片代理服务避免403错误
    - 支持gzip压缩的API响应

12. **订阅管理优化** - 精确的批量操作
    - 基于唯一ID的订阅管理
    - 支持ID数组批量删除: POST /delSubscribe
    - 立即触发功能: POST /triggerNow
    - 向后兼容旧的删除格式
    - 前端批量选择和操作功能
    - 线程安全的并发操作

## 技术栈

### 后端
- **Go 1.19+** - 主要开发语言
- **Gin** - HTTP Web框架
- **goquery** - HTML解析库
- **sync** - 并发控制和线程安全
- **crypto/rand** - 唯一ID生成
- **encoding/json** - JSON数据处理

### 前端
- **Vue 3** - 前端框架
- **Element Plus** - UI组件库
- **Vue Router 4** - 路由管理
- **Axios** - HTTP客户端
- **Vite** - 构建工具

## 项目结构

```
tvsubscribe/
├── cmd/                    # 命令行入口
│   ├── main.go            # 主程序入口和HTTP服务
│   └── cli.go             # CLI命令处理逻辑
├── server/                 # HTTP服务器
│   └── server.go          # Gin路由和API实现
├── web/                    # Vue管理界面
│   ├── src/               # Vue源码
│   │   ├── views/         # 页面组件
│   │   │   ├── Home.vue   # 首页/配置管理
│   │   │   └── Subscribe.vue # 订阅管理
│   │   ├── components/    # 公共组件
│   │   └── router/        # 路由配置
│   ├── dist/              # 构建输出（git忽略）
│   └── package.json       # 前端依赖
├── subscribe/              # 订阅管理
│   └── manager.go         # 订阅管理器实现
├── config/                 # 配置管理
│   └── config.go          # 配置管理器实现
├── interfaces.go           # 核心接口定义
├── torrentList.go          # 种子查询和解析
└── downloadTorrent.go      # 种子下载逻辑
```

## 部署方式

### 1. 单进程部署（推荐）
```bash
# 构建
go build -o tvsubscribe cmd/*.go

# 运行（自动启动Web界面）
./tvsubscribe
```

### 2. 开发模式
```bash
# 后端服务
go run cmd/*.go

# 前端开发服务器
cd web && npm run dev
```

## 配置文件

### config.json
```json
{
  "endpoint": "https://springsunday.net",
  "cookie": "your_cookie",
  "interval_minutes": 60,
  "wechat_server": "your_wechat_server",
  "wechat_token": "your_wechat_token",
  "port": 8443
}
```

### subscribes.json
```json
[
  {
    "id": "a1b2c3d4e5f6",
    "douban_id": "36391902",
    "name": "庆余年 第二季",
    "resolution": 1
  }
]
```

## API接口

| 方法 | 路径 | 功能 | 参数 |
|------|------|------|------|
| GET | /getConfig | 获取配置 | - |
| POST | /setConfig | 设置配置 | Config结构体 |
| GET | /getSubscribeList | 获取订阅列表 | - |
| POST | /addSubscribe | 添加订阅 | TVInfo结构体 |
| POST | /delSubscribe | 删除订阅 | IDs数组或TVInfo（兼容） |
| POST | /triggerNow | 立即触发订阅 | IDs数组 |
| GET | /searchDouBan | 豆瓣搜索 | name参数 |
| GET | /proxy/image | 图片代理 | url参数 |
| GET | /health | 健康检查 | - |

## 使用示例

### Web界面
1. 访问 http://localhost:8443
2. 在配置管理页面设置服务器参数
3. 在订阅管理页面添加/删除订阅
4. 使用豆瓣搜索功能快速添加新订阅

### 命令行
```bash
# 查看配置
./tvsubscribe config --list

# 设置配置
./tvsubscribe config --set "interval_minutes=30"

# 添加订阅
./tvsubscribe subscribe --add "douban_id=36391902" "resolution=1"
```

## 重构效果

### 改进前
- 简单的ID模式种子处理
- 基础的配置文件管理
- 命令行交互界面

### 改进后
- 完整的TorrentInfo结构化处理
- HTTP API + Web界面 + CLI三重管理方式
- 现代化的用户界面
- 智能名称获取和通知功能
- 豆瓣搜索和批量操作
- 生产级的服务架构
- 线程安全的数据管理

## 后续优化建议

### 1. 性能优化
- 实现种子信息缓存
- 添加并发下载支持
- 优化豆瓣API调用频率

### 2. 功能扩展
- 支持多个PT站点
- 添加种子质量筛选
- 实现下载历史记录
- 支持剧集进度跟踪

### 3. 用户体验
- 添加实时下载进度显示
- 支持批量导入订阅
- 移动端适配
- 多语言支持

### 4. 运维监控
- 添加Prometheus监控指标
- 实现日志分级和轮转
- 支持Docker容器化部署
- 健康检查和自动重启

## 版本历史

### v1.0 (原始版本)
- 基础种子下载功能
- 简单配置管理
- 命令行界面

### v2.0 (重构版本)
- TorrentInfo结构化处理
- HTTP API服务
- Vue管理界面
- 豆瓣搜索功能
- 批量操作支持
- 立即触发功能
- 微信通知集成
- 线程安全改进