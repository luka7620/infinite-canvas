# 部署说明

## 服务器 Docker 一键部署

服务器已安装 Docker 和 Docker Compose 插件后，先创建并进入一个空目录，再运行：

```bash
mkdir -p infinite-canvas
cd infinite-canvas
curl -fsSL https://raw.githubusercontent.com/luka7620/infinite-canvas/main/deploy/docker-deploy.sh | bash
```

脚本会在当前目录准备部署文件，自动生成 `.env`、`ADMIN_PASSWORD`、`JWT_SECRET`、`POSTGRES_PASSWORD`、`REDIS_PASSWORD`，创建 `data/`、`postgres_data/`、`redis_data/`，并启动应用本体、PostgreSQL 和 Redis。

默认只监听本机地址，适合由服务器上已有的 Nginx 反向代理：

```text
http://127.0.0.1:3000
```

默认镜像是 `luka762/infinite-canvas:latest`。如果使用自己的镜像或端口，可以在同一行命令里传环境变量：

```bash
curl -fsSL https://raw.githubusercontent.com/luka7620/infinite-canvas/main/deploy/docker-deploy.sh | APP_IMAGE=你的镜像:latest SERVER_PORT=3000 bash
```

如果当前用户没有 Docker 权限，可以把最后一行改成 `curl ... | sudo bash`；如需显式指定目录，可加 `INSTALL_DIR=/path/to/dir`。

常用命令：

```bash
docker compose ps
docker compose logs -f app
docker compose up -d
docker compose down
```

部署目录里的本地目录就是持久化数据；迁移服务器时先 `docker compose down`，再整体打包迁移这个部署目录。

如果需要多实例部署，可以参考 `deploy/docker-compose.lb.yml` 和 `deploy/Caddyfile`，让多个 app 副本共享同一个 Postgres 和 Redis，再由 Caddy 或服务器上的 Nginx 做负载均衡。

## Render 部署

点击下面按钮即可部署到 Render：

[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy?repo=https://github.com/basketikun/infinite-canvas)

### 部署步骤

1. 点击 `Deploy to Render`。
2. 登录 Render，并按页面提示连接 GitHub。
3. 填写 `ADMIN_PASSWORD`，然后点击确认部署。

部署完成后，打开 Render 分配的 `.onrender.com` 域名即可访问。

### 免费版说明

默认使用 Render 免费 Web Service：

- 空闲约 15 分钟后会休眠，下次访问会自动唤醒。
- 免费版本地文件不是持久化存储，SQLite 数据可能在重启、重新部署后丢失。
- 适合体验和演示，不适合长期保存正式数据。

如果要长期使用，建议升级 Render 付费实例并挂载 Persistent Disk，或改用 PostgreSQL。

### 管理员账号

默认管理员用户名：

```text
admin
```

管理员密码是在 Render 部署页面里填写的 `ADMIN_PASSWORD`。
