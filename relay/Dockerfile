# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 复制 common 模块
COPY common/ ./common/

# 复制 relay 模块
COPY relay/ ./relay/

# 构建
WORKDIR /app/relay
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/relay ./cmd

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin/relay .
COPY relay/configs/config.toml ./configs/

EXPOSE 8082

CMD ["./relay", "-c", "config", "-cPath", "./,./configs/"]
