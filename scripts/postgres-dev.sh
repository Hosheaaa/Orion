#!/usr/bin/env bash

# 启动/停止/查看本地 PostgreSQL（开发环境）
# 可通过环境变量覆盖：PG_BIN_DIR、PG_DATA_DIR、PG_LOG_FILE

set -euo pipefail

PG_BIN_DIR=${PG_BIN_DIR:-"/opt/homebrew/opt/postgresql@16/bin"}
PG_DATA_DIR=${PG_DATA_DIR:-"/opt/homebrew/var/postgresql@16"}
PG_LOG_FILE=${PG_LOG_FILE:-"$PG_DATA_DIR/pg-dev.log"}

PG_CTL="$PG_BIN_DIR/pg_ctl"

usage() {
  cat <<'EOF'
用法: ./postgres-dev.sh <start|stop|status|restart>

环境变量:
  PG_BIN_DIR   PostgreSQL 可执行文件目录，默认 /opt/homebrew/opt/postgresql@16/bin
  PG_DATA_DIR  PostgreSQL 数据目录，默认 /opt/homebrew/var/postgresql@16
  PG_LOG_FILE  启动日志输出文件，默认 PG_DATA_DIR/pg-dev.log
EOF
}

ensure_pgctl() {
  if [[ ! -x "$PG_CTL" ]]; then
    echo "找不到 pg_ctl：$PG_CTL" >&2
    echo "请检查是否已安装 PostgreSQL，或设置 PG_BIN_DIR。" >&2
    exit 1
  fi
}

ensure_datadir() {
  if [[ ! -d "$PG_DATA_DIR" ]]; then
    echo "数据目录不存在：$PG_DATA_DIR" >&2
    echo "请确认 Postgres 是否初始化完成，或设置 PG_DATA_DIR。" >&2
    exit 1
  fi
}

start_pg() {
  ensure_pgctl
  ensure_datadir
  mkdir -p "$(dirname "$PG_LOG_FILE")"
  "$PG_CTL" -D "$PG_DATA_DIR" -l "$PG_LOG_FILE" start
}

stop_pg() {
  ensure_pgctl
  ensure_datadir
  "$PG_CTL" -D "$PG_DATA_DIR" stop -m fast
}

status_pg() {
  ensure_pgctl
  ensure_datadir
  "$PG_CTL" -D "$PG_DATA_DIR" status
}

restart_pg() {
  stop_pg || true
  start_pg
}

if [[ $# -ne 1 ]]; then
  usage
  exit 1
fi

case "$1" in
  start)
    start_pg
    ;;
  stop)
    stop_pg
    ;;
  status)
    status_pg
    ;;
  restart)
    restart_pg
    ;;
  *)
    usage
    exit 1
    ;;
esac
