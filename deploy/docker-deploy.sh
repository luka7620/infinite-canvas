#!/usr/bin/env bash
set -euo pipefail

RAW_BASE_URL="${RAW_BASE_URL:-https://raw.githubusercontent.com/luka7620/infinite-canvas/main/deploy}"
INSTALL_DIR="${INSTALL_DIR:-$(pwd)}"
START_SERVICES="${START_SERVICES:-1}"

red='\033[0;31m'
green='\033[0;32m'
yellow='\033[1;33m'
blue='\033[0;34m'
nc='\033[0m'

info() { echo -e "${blue}[INFO]${nc} $*"; }
success() { echo -e "${green}[OK]${nc} $*"; }
warn() { echo -e "${yellow}[WARN]${nc} $*"; }
error() { echo -e "${red}[ERROR]${nc} $*"; }

has_cmd() {
    command -v "$1" >/dev/null 2>&1
}

download() {
    local url="$1"
    local target="$2"
    if has_cmd curl; then
        curl -fsSL "$url" -o "$target"
    elif has_cmd wget; then
        wget -q "$url" -O "$target"
    else
        error "需要先安装 curl 或 wget"
        exit 1
    fi
}

secret() {
    if has_cmd openssl; then
        openssl rand -hex 32
    else
        od -An -N32 -tx1 /dev/urandom | tr -d ' \n'
        echo
    fi
}

set_env() {
    local key="$1"
    local value="$2"
    local file=".env"
    local escaped
    escaped=$(printf '%s' "$value" | sed 's/[\/&]/\\&/g')
    if grep -q "^${key}=" "$file"; then
        sed -i "s/^${key}=.*/${key}=${escaped}/" "$file"
    else
        printf '%s=%s\n' "$key" "$value" >>"$file"
    fi
}

set_env_from_current() {
    local key="$1"
    local value="${!key:-}"
    if [ -n "$value" ]; then
        set_env "$key" "$value"
    fi
}

set_env_or_secret() {
    local key="$1"
    local value="${!key:-}"
    if [ -n "$value" ]; then
        set_env "$key" "$value"
    else
        set_env "$key" "$(secret)"
    fi
}

env_file_value() {
    local key="$1"
    awk -F= -v key="$key" '$1 == key { sub(/^[^=]*=/, ""); print; exit }' .env 2>/dev/null || true
}

main() {
    if ! has_cmd docker; then
        error "需要先安装 Docker 和 Docker Compose"
        exit 1
    fi
    if ! docker compose version >/dev/null 2>&1; then
        error "当前 Docker 不支持 docker compose 命令"
        exit 1
    fi

    if ! mkdir -p "$INSTALL_DIR"; then
        error "无法创建 $INSTALL_DIR，请先进入有写权限的部署目录，或显式指定 INSTALL_DIR"
        exit 1
    fi
    if [ ! -w "$INSTALL_DIR" ]; then
        error "$INSTALL_DIR 不可写，请先进入有写权限的部署目录，或使用 sudo 运行"
        exit 1
    fi
    cd "$INSTALL_DIR"

    if [ -f docker-compose.yml ] || [ -f .env ]; then
        warn "$INSTALL_DIR 已存在部署文件，将只补齐缺失文件并保留现有 .env"
    fi

    info "下载部署文件到 $INSTALL_DIR"
    download "$RAW_BASE_URL/docker-compose.server.yml" docker-compose.yml
    download "$RAW_BASE_URL/.env.server.example" .env.example

    if [ ! -f .env ]; then
        cp .env.example .env
        for key in APP_IMAGE BIND_HOST SERVER_PORT ADMIN_USERNAME POSTGRES_USER POSTGRES_DB DATABASE_MAX_OPEN_CONNS DATABASE_MAX_IDLE_CONNS REDIS_DB REDIS_CACHE_TTL_SECONDS REDIS_KEY_PREFIX TZ; do
            set_env_from_current "$key"
        done
        set_env_or_secret ADMIN_PASSWORD
        set_env_or_secret JWT_SECRET
        set_env_or_secret POSTGRES_PASSWORD
        set_env_or_secret REDIS_PASSWORD
        chmod 600 .env || true
        success "已生成 .env 和随机密钥"
        info "管理员密码保存在 $INSTALL_DIR/.env 的 ADMIN_PASSWORD"
    else
        warn "检测到已有 .env，未覆盖密钥和数据库密码"
    fi

    mkdir -p data postgres_data redis_data
    success "已准备数据目录"

    if [ "$START_SERVICES" = "1" ]; then
        server_port="${SERVER_PORT:-$(env_file_value SERVER_PORT)}"
        server_port="${server_port:-3000}"
        info "启动服务"
        docker compose up -d --remove-orphans
        success "已启动。访问地址: http://localhost:$server_port"
        info "查看日志: cd $INSTALL_DIR && docker compose logs -f app"
    else
        success "部署文件已准备完成"
        info "启动命令: cd $INSTALL_DIR && docker compose up -d"
    fi
}

main "$@"
