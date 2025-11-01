# TVSubscribe - 电视剧种子自动下载器

一个自动从 SpringSunday 网站查询并下载电视剧种子的工具，支持定时检查和自动添加到 Transmission 下载器。

## 功能特性

- 📺 支持多个电视剧订阅
- ⏰ 定时自动检查新种子
- 🎯 支持不同分辨率 (2160P/1080P)
- 📥 自动下载并添加到 Transmission
- ⚙️ 配置文件驱动，易于管理
- 🔄 配置文件热重载，修改后自动生效

## 配置说明

### 配置文件 (config.json)

```json
{
  "endpoint": "http://username:password@host:port/transmission/rpc",
  "cookie": "your_springsunday_cookie",
  "passkey": "your_springsunday_passkey",
  "interval_minutes": 60,
  "subscribes": [
    {
      "douban_id": "36391902",
      "resolution": 0
    },
    {
      "douban_id": "26798436",
      "resolution": 1
    }
  ]
}
```

### 配置字段说明

- `endpoint`: Transmission RPC 地址
- `cookie`: SpringSunday 网站的登录 Cookie
- `passkey`: SpringSunday 网站的 Passkey
- `interval_minutes`: 检查间隔（分钟），默认 60 分钟
- `subscribes`: 订阅的电视剧列表
  - `douban_id`: 豆瓣ID
  - `resolution`: 分辨率 (0=2160P, 1=1080P)

## 使用方法

1. 编辑 `config.json` 文件，填入你的配置信息
2. 运行程序：
   ```bash
   cd cmd
   go build -o tvsubscribe.exe
   ./tvsubscribe.exe
   ```

## 运行模式

程序启动后会：
1. 立即执行一次所有订阅的检查
2. 然后按照配置的间隔定时执行
3. 启动配置文件监视，支持热重载
4. 每次执行会：
   - 查询每个订阅的种子列表
   - 下载新种子到 `torrents/` 目录
   - 自动添加到 Transmission 下载

## 热重载功能

程序支持配置文件热重载，修改 `config.json` 文件后会自动重新加载配置并立即执行：

- ✅ 添加新的电视剧订阅 - 立即检查新订阅的种子
- ✅ 修改检查间隔时间 - 下一次定时任务使用新的间隔
- ✅ 更新 Cookie 或 Passkey - 立即使用新的认证信息
- ✅ 修改 Transmission 端点 - 立即使用新的下载端点

修改配置文件后，程序会：
1. 立即重新加载配置
2. 立即重新处理所有订阅的电视剧
3. 下一次定时任务使用新的配置

## 日志输出

程序会输出详细的日志信息，包括：
- 配置加载状态
- 每个订阅的处理进度
- 找到的种子数量
- 下载成功/失败信息
- 配置文件重载状态
- 程序启动和退出信息

## 程序控制

- **启动** - 程序启动后立即执行一次所有订阅检查
- **运行** - 按配置间隔定时执行，同时监控配置文件变化
- **退出** - 按 `Ctrl+C` 发送 SIGINT 信号优雅退出

## 注意事项

- 确保配置文件中包含有效的 Cookie 和 Passkey
- 确保 Transmission 服务正常运行
- 种子文件会保存在 `torrents/` 目录下
- 已存在的种子文件会自动跳过下载
- 使用 `Ctrl+C` 优雅退出，避免强制终止