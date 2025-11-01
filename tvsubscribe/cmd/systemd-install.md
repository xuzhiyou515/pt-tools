# Systemd 服务安装说明

## 前提条件

- Linux 系统（支持 systemd）
- Go 1.23.2 或更高版本
- Transmission 服务已安装并运行

## 安装步骤

### 1. 编译程序

```bash
cd cmd
go build -o tvsubscribe
```

### 2. 创建系统用户

```bash
sudo useradd -r -s /bin/false -d /opt/tvsubscribe tvsubscribe
```

### 3. 创建安装目录

```bash
sudo mkdir -p /opt/tvsubscribe
sudo chown tvsubscribe:tvsubscribe /opt/tvsubscribe
```

### 4. 复制文件

```bash
# 复制可执行文件
sudo cp tvsubscribe /opt/tvsubscribe/

# 复制配置文件（需要根据实际情况修改）
sudo cp config.json /opt/tvsubscribe/

# 设置权限
sudo chown tvsubscribe:tvsubscribe /opt/tvsubscribe/tvsubscribe
sudo chown tvsubscribe:tvsubscribe /opt/tvsubscribe/config.json
sudo chmod 755 /opt/tvsubscribe/tvsubscribe
sudo chmod 600 /opt/tvsubscribe/config.json
```

### 5. 安装 systemd 服务

```bash
# 复制服务文件
sudo cp tvsubscribe.service /etc/systemd/system/

# 重新加载 systemd
sudo systemctl daemon-reload

# 启用服务
sudo systemctl enable tvsubscribe.service

# 启动服务
sudo systemctl start tvsubscribe.service
```

## 服务管理

### 启动服务
```bash
sudo systemctl start tvsubscribe.service
```

### 停止服务
```bash
sudo systemctl stop tvsubscribe.service
```

### 重启服务
```bash
sudo systemctl restart tvsubscribe.service
```

### 查看服务状态
```bash
sudo systemctl status tvsubscribe.service
```

### 查看服务日志
```bash
sudo journalctl -u tvsubscribe.service -f
```

### 启用开机自启
```bash
sudo systemctl enable tvsubscribe.service
```

### 禁用开机自启
```bash
sudo systemctl disable tvsubscribe.service
```

## 配置文件说明

确保 `/opt/tvsubscribe/config.json` 文件包含正确的配置：

- `endpoint`: Transmission RPC 地址
- `cookie`: SpringSunday 网站的登录 Cookie
- `passkey`: SpringSunday 网站的 Passkey
- `interval_minutes`: 检查间隔（分钟）
- `subscribes`: 订阅的电视剧列表

## 安全注意事项

- 服务以专用用户 `tvsubscribe` 运行
- 配置文件权限设置为 600，仅限所有者读写
- 使用 systemd 的安全选项限制服务权限
- 日志通过 journald 管理

## 故障排除

### 检查服务状态
```bash
sudo systemctl status tvsubscribe.service
```

### 查看详细日志
```bash
sudo journalctl -u tvsubscribe.service -n 50 --no-pager
```

### 检查文件权限
```bash
ls -la /opt/tvsubscribe/
```

### 检查用户和组
```bash
id tvsubscribe
```