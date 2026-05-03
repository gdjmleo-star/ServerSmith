# ============================================================
# Stage 1: Build web frontend (static export)
# ============================================================
FROM node:20-alpine AS web-builder
WORKDIR /web
COPY web/package.json web/package-lock.json* ./
RUN npm ci
COPY web/ .
RUN npm run build

# ============================================================
# Stage 2: Build Go binary with embedded frontend
# ============================================================
FROM golang:1.22-alpine AS go-builder
WORKDIR /build
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ .
COPY --from=web-builder /web/out/ ./web-out/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-X main.Version=$(date -u +%Y%m%d)" -o /serversmith-server .

# ============================================================
# Stage 3: Runtime
# ============================================================
FROM alpine:3.19
RUN apk add --no-cache ca-certificates openssh-client openssh
COPY --from=go-builder /serversmith-server /usr/local/bin/
EXPOSE 18080
CMD ["/usr/local/bin/serversmith-server"]
