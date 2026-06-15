#!/bin/sh
set -eu

PORT=9080 /app/server &
api_pid=$!

cd /app/web
HOSTNAME=0.0.0.0 PORT=3000 API_BASE_URL=http://127.0.0.1:9080 node server.cjs &
web_pid=$!

stop() {
    kill "$api_pid" "$web_pid" 2>/dev/null || true
}

trap stop INT TERM

while kill -0 "$api_pid" 2>/dev/null && kill -0 "$web_pid" 2>/dev/null; do
    sleep 2
done

stop
wait "$api_pid" 2>/dev/null || true
wait "$web_pid" 2>/dev/null || true
exit 1
