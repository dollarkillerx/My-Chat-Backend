# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 复制 common 模块
COPY common/ ./common/

# 复制 seaking 模块
COPY seaking/ ./seaking/

# 构建
WORKDIR /app/seaking
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/seaking ./cmd

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin/seaking .
COPY seaking/configs/config.toml ./configs/

EXPOSE 8081

CMD ["./seaking", "-c", "config", "-cPath", "./,./configs/"]
