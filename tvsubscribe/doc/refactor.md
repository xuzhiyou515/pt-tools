# TVSubscribe 重构文档

## 重构目标
将原有的种子处理逻辑从基于 ID 的简单模式升级为基于完整信息的 TorrentInfo 结构化模式，提供更丰富的种子信息和更可靠的下载机制。

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

### 4. 调整配置文件 TODO
- 在 tvsubscribe\cmd\main.go 中 Config 移除 Subscribes
- Subscribes 使用独立的变量，从 ./subscribes.json 中读取数据
- Config 添加 Port 字段，用于配置监听的端口，默认监听8443端口

### 5、优化配置和订阅调整 TODO
使用gin框架，支持以下http接口来获取、调整配置和订阅。
- GET /getConfig 返回 Config json 结构
- POST /setConfig payload 是 Config 结构（未包含的字段不修改）
  - 更新内存的Config数据，并把更新后的Config写入配置文件
- GET /getSubscribeList 返回 Subcribes 列表
- POST /addSubscribe payload 是 tvsubscribe.TVInfo
  - 添加新的订阅，需要注意去重, 更新配置文件
- POST /delSubscribe payload 是 tvsubscribe.TVInfo
  - 删除订阅，更新配置文件
支持通过命令行来调整配置、订阅,实现方式是调用对应的http接口
tvsubscribe config --list 获取配置
tvsubscribe config --set "endpoint=xxxx" "cookie=xxxx" "interval_minutes=5" "wechat_server=xxxx" "wechat_token=xxxx" 设置配置（配置时不需要每项都设置）
tvsubscribe subscribe --list 获取订阅
tvsubscribe subscribe --add "douban_id=xxxxx" "resolution=1" 添加订阅 （resolution可以省略，默认是1）
tvsubscribe subscribe --add "douban_id=xxxxx" "resolution=1" 删除订阅 （resolution可以省略，默认是1）
以上命令都支持通过 --url 配置服务器地址 默认是 "127.0.0.1:8443"

