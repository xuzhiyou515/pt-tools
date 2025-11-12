# TVSubscribe 管理界面

基于 Vue 3 + Element Plus 的 TVSubscribe 管理界面，用于管理配置和电视剧订阅。

## 功能特性

- **配置管理**: 管理服务器配置、Cookie、检查间隔、微信通知等
- **订阅管理**: 添加、删除电视剧订阅，支持不同分辨率选择
- **响应式设计**: 适配不同屏幕尺寸
- **实时操作**: 直接调用后端API进行配置和订阅的增删改查

## 技术栈

- Vue 3
- Vue Router 4
- Element Plus
- Axios
- Vite

## 快速开始

### 1. 安装依赖

```bash
cd tvsubscribe/web
npm install
```

### 2. 启动开发服务器

```bash
npm run dev
```

开发服务器将在 http://localhost:3000 启动，并自动代理API请求到 http://localhost:8443

### 3. 构建生产版本

```bash
npm run build
```

构建后的文件将输出到 `dist` 目录。

### 4. 预览生产版本

```bash
npm run preview
```

## API 接口

管理界面通过以下 HTTP 接口与后端通信：

### 配置管理
- `GET /api/getConfig` - 获取当前配置
- `POST /api/setConfig` - 更新配置

### 订阅管理
- `GET /api/getSubscribeList` - 获取订阅列表
- `POST /api/addSubscribe` - 添加新订阅
- `POST /api/delSubscribe` - 删除订阅

### 健康检查
- `GET /api/health` - 健康检查

## 配置说明

### 服务器地址 (endpoint)
PT 站点的完整地址，用于获取种子信息和下载链接。

### Cookie
从浏览器复制的相关站点的 Cookie，用于身份验证。

### 检查间隔 (interval_minutes)
自动检查新种子的时间间隔，单位为分钟。

### 微信通知
- `wechat_server`: 微信服务器地址
- `wechat_token`: 微信推送 Token

### 监听端口 (port)
HTTP 服务器监听的端口，默认 8443。

## 项目结构

```
web/
├── src/
│   ├── components/        # 公共组件
│   ├── views/            # 页面组件
│   │   ├── Config.vue    # 配置管理页面
│   │   └── Subscribe.vue # 订阅管理页面
│   ├── router/           # 路由配置
│   ├── App.vue           # 根组件
│   └── main.js           # 入口文件
├── index.html            # HTML 模板
├── package.json          # 项目配置
├── vite.config.js        # Vite 配置
└── README.md             # 说明文档
```

## 注意事项

1. 确保后端服务 (tvsubscribe) 已启动并运行在正确的端口
2. 配置中的 Cookie 和敏感信息请注意保护
3. 微信通知功能需要配置正确的服务器地址和 Token
4. 订阅的豆瓣ID可以从豆瓣电影页面获取