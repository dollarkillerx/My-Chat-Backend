# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 复制 common 模块
COPY common/ ./common/

# 复制 gateway 模块
COPY gateway/ ./gateway/

# 构建
WORKDIR /app/gateway
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/gateway ./cmd

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin/gateway .
COPY gateway/configs/config.toml ./configs/

EXPOSE 8080

CMD ["./gateway", "-c", "config", "-cPath", "./,./configs/"]
