# syntax=docker/dockerfile:1.6

############################
# Stage 1: build Next.js console (static export)
############################
FROM node:22-alpine AS console-builder
WORKDIR /app

COPY console/package.json console/package-lock.json* ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi

COPY console/ ./
RUN npm run build:export


############################
# Stage 2: build Go binary with the console embedded
############################
FROM golang:1.23-alpine AS go-builder
WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY backend/go.mod backend/go.sum* ./backend/
RUN cd backend && go mod download

COPY backend/ ./backend/
COPY --from=console-builder /app/out/. ./backend/internal/server/web/

ENV CGO_ENABLED=0 GOOS=linux
RUN cd backend && go build -trimpath -ldflags="-s -w" -o /out/nextonebotfilter ./cmd/nextonebotfilter


############################
# Stage 3: runtime
############################
FROM alpine:3.20 AS runtime
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata tini wget \
    && addgroup -S app && adduser -S app -G app

ENV TZ=Asia/Shanghai

COPY --from=go-builder /out/nextonebotfilter /app/nextonebotfilter

COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh /app/nextonebotfilter \
    && mkdir -p /app/data \
    && chown -R app:app /app

USER app

# 8787: console (UI + API), 3939: OneBot reverse-WS gateway (default)
EXPOSE 8787 3939

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8787/api/health || exit 1

VOLUME ["/app/data"]

ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/docker-entrypoint.sh"]
CMD ["/app/nextonebotfilter", "-db", "/app/data/nextonebotfilter.db", "-console", ":8787", "-log", "/app/data/nextonebotfilter.log"]
