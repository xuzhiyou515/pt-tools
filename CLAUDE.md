- 修改代码后可以用部署脚本来部署后端程序并测试

## TVSubscribe 部署脚本用法

所有部署脚本都已更新为用户级部署，无需 sudo 权限，文件部署到 `~/.config/tvsubscribe/` 目录。

### 部署脚本

1. **完整部署脚本** `./tvsubscribe/deploy.sh`
   - 构建并部署后端和前端
   - 停止服务、复制文件、启动服务
   - 用法：`./tvsubscribe/deploy.sh`

2. **后端部署脚本** `./tvsubscribe/deploy-backend.sh`
   - 仅构建并部署 Go 后端程序
   - 用法：`./tvsubscribe/deploy-backend.sh`

3. **前端部署脚本** `./tvsubscribe/deploy-web.sh`
   - 仅构建并部署前端文件
   - 用法：`./tvsubscribe/deploy-web.sh`

### 服务管理（无需 sudo）

```bash
systemctl --user start tvsubscribe      # 启动服务
systemctl --user stop tvsubscribe       # 停止服务
systemctl --user restart tvsubscribe    # 重启服务
systemctl --user status tvsubscribe     # 查看状态
systemctl --user enable tvsubscribe     # 开机自启
systemctl --user disable tvsubscribe    # 禁用开机自启
```

### 查看日志

```bash
journalctl --user -u tvsubscribe -f     # 实时日志
journalctl --user -u tvsubscribe -n 50  # 最近 50 行日志
```

### 用户目录结构

```
~/.config/tvsubscribe/
├── config.json        # 配置文件
├── subscribes.json    # 订阅配置
├── tvsubscribe        # 后端程序
└── web/
    └── dist/          # 前端文件
```

### Web 界面访问

服务启动后，可通过以下地址访问：
- HTTP API: http://127.0.0.1:8443
- Web 管理界面: http://127.0.0.1:8443

### 注意事项

- 所有操作都不需要 sudo 权限
- 使用用户级 systemd 服务
- 配置文件位于用户目录，可自由修改
- 部署脚本会自动创建必要的目录结构