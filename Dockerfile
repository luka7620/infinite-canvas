ARG BUN_IMAGE=oven/bun:1.3.13
ARG GO_IMAGE=golang:1.25-alpine
ARG NODE_IMAGE=node:22-bookworm-slim

FROM --platform=$BUILDPLATFORM ${BUN_IMAGE} AS web-build

WORKDIR /app/web
COPY web/package.json web/bun.lock ./
RUN --mount=type=cache,target=/root/.bun/install/cache bun install --frozen-lockfile --registry=https://registry.npmmirror.com --cache-dir=/root/.bun/install/cache
COPY VERSION /app/VERSION
COPY CHANGELOG.md /app/CHANGELOG.md
COPY web ./
RUN bun run build

FROM --platform=$BUILDPLATFORM ${GO_IMAGE} AS api-build

WORKDIR /app
ARG TARGETOS
ARG TARGETARCH
ENV GOPROXY=https://goproxy.cn|https://goproxy.io|https://proxy.golang.org|direct
COPY go.mod go.sum ./
COPY config ./config
COPY handler ./handler
COPY middleware ./middleware
COPY model ./model
COPY repository ./repository
COPY router ./router
COPY service ./service
COPY main.go ./
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -trimpath -o /server .

FROM ${NODE_IMAGE}

WORKDIR /app
COPY VERSION /app/VERSION
COPY CHANGELOG.md /app/CHANGELOG.md
COPY docker-start.sh /app/docker-start.sh
COPY --from=api-build /server /app/server
COPY --from=web-build /app/web/.next/standalone /app/web
COPY --from=web-build /app/web/.next/static /app/web/.next/static
COPY --from=web-build /app/web/public /app/web/public
RUN mv /app/web/server.js /app/web/server.cjs
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PROMPT_DATA_DIR=/app/data/prompts
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*
RUN mkdir -p /app/data/prompts && chmod +x /app/docker-start.sh

EXPOSE 3000
CMD ["/app/docker-start.sh"]
