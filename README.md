# My-Chat-Backend

基于 Nostr 协议思想的中心化即时通讯后端系统。

## 架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                          Client                                  │
│                    (Mobile / Web / Desktop)                      │
└─────────────────────────────────────────────────────────────────┘
                    │                       │
          HTTP API  │                       │ WebSocket
      (register/login)                      │
                    ▼                       ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Gateway 集群                             │
│              (WebSocket连接管理 / 消息路由 / 鉴权)                 │
└─────────────────────────────────────────────────────────────────┘
                    │                       │
          JSON-RPC  │                       │ JSON-RPC
                    ▼                       ▼
┌───────────────────────────┐   ┌───────────────────────────────┐
│         SeaKing           │   │           Relay               │
│    (用户中心/关系中心)      │   │       (事件存储层)             │
│  • 用户注册/登录           │   │  • Event 存储                 │
│  • 好友管理               │   │  • 消息查询                    │
│  • 群组管理               │   │  • 已读回执                    │
│  • 会话管理               │   │  • 消息反应                    │
└───────────────────────────┘   └───────────────────────────────┘
            │                               │
            ▼                               ▼
┌───────────────────────────┐   ┌───────────────────────────────┐
│        PostgreSQL         │   │         PostgreSQL            │
│       (mychat DB)         │   │       (mychat_relay DB)       │
└───────────────────────────┘   └───────────────────────────────┘
```

## 通信架构

本系统仅使用两种通信方式：

1. **HTTP API** (仅 Gateway 暴露): 用于用户注册和登录
2. **WebSocket** (Gateway): 客户端与服务端的实时通信通道
3. **JSON-RPC 2.0** (内部): 服务间的内部通信

**注意**: SeaKing 和 Relay 服务不暴露 REST API，仅通过 JSON-RPC 与 Gateway 通信。

## 项目结构

```
My-Chat-Backend/
├── common/                 # 公共库
│   └── pkg/
│       ├── auth/          # JWT 认证
│       ├── client/        # 服务间 RPC 客户端
│       ├── config/        # 配置加载
│       ├── crypto/        # 加密工具
│       ├── errors/        # 错误定义
│       ├── log/           # 日志
│       ├── middleware/    # 中间件
│       └── protocol/      # 协议定义
├── gateway/               # 网关服务
│   ├── cmd/              # 入口
│   └── internal/
│       ├── conf/         # 配置
│       ├── handler/      # 消息处理
│       ├── server/       # HTTP服务
│       └── ws/           # WebSocket管理
├── seaking/              # 用户中心服务
│   ├── cmd/
│   └── internal/
│       ├── conf/
│       ├── model/        # 数据模型
│       ├── rpc/          # JSON-RPC
│       ├── server/
│       ├── service/      # 业务逻辑
│       └── storage/
├── relay/                # 事件存储服务
│   ├── cmd/
│   └── internal/
│       ├── conf/
│       ├── model/
│       ├── rpc/
│       ├── server/
│       ├── service/
│       └── storage/
├── scripts/              # 脚本
├── docker-compose.yml    # Docker 编排
├── Makefile             # 构建脚本
└── IM Rules.md          # 协议设计文档
```

## 快速开始

### 环境要求

- Go 1.21+
- PostgreSQL 14+
- Redis 7+
- Docker & Docker Compose (可选)

### 使用 Docker Compose 启动

```bash
# 启动基础设施
docker-compose up -d postgres redis

# 构建并启动所有服务
docker-compose up -d
```

### 本地开发

```bash
# 安装依赖
make tidy

# 构建所有服务
make build

# 数据库迁移
make migrate

# 启动服务 (分别在不同终端)
make run-seaking
make run-relay
make run-gateway
```

### 运行测试

```bash
make test
```

## 服务端口

| 服务 | 端口 | 说明 |
|------|------|------|
| Gateway | 8080 | HTTP API + WebSocket |
| SeaKing | 8081 | JSON-RPC (内部) |
| Relay | 8082 | JSON-RPC (内部) |

## Gateway API

### HTTP 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/register` | 用户注册 |
| POST | `/api/login` | 用户登录 |
| GET | `/api/stats` | 获取在线统计 |
| GET | `/health` | 健康检查 |

### 注册请求

```json
POST /api/register
{
    "username": "user1",
    "password": "password123",
    "nickname": "User One",
    "phone": "13800138000",
    "email": "user@example.com"
}
```

### 登录请求

```json
POST /api/login
{
    "username": "user1",
    "password": "password123",
    "device_id": "device-uuid",
    "platform": "ios"
}
```

### WebSocket 连接

连接地址: `ws://localhost:8080/ws?token=<JWT_TOKEN>`

### WebSocket 命令

| 命令 | 说明 | 方向 |
|------|------|------|
| `ping` | 心跳请求 | C -> S |
| `pong` | 心跳响应 | S -> C |
| `event` | 事件消息 | 双向 |
| `ack` | 消息确认 | S -> C |
| `error` | 错误响应 | S -> C |
| `result` | 结果响应 | S -> C |
| `subscribe` | 订阅会话 | C -> S |
| `unsubscribe` | 取消订阅 | C -> S |
| `sync` | 同步历史消息 | C -> S |

#### 好友命令

| 命令 | 说明 |
|------|------|
| `get_friends` | 获取好友列表 |
| `send_friend_request` | 发送好友请求 |
| `get_friend_requests` | 获取待处理好友请求 |
| `accept_friend_request` | 接受好友请求 |
| `reject_friend_request` | 拒绝好友请求 |
| `delete_friend` | 删除好友 |

#### 会话命令

| 命令 | 说明 |
|------|------|
| `get_conversations` | 获取会话列表 |
| `create_conversation` | 创建会话 |
| `get_conversation_members` | 获取会话成员 |

#### 群组命令

| 命令 | 说明 |
|------|------|
| `get_groups` | 获取群组列表 |
| `create_group` | 创建群组 |
| `get_group_info` | 获取群组信息 |
| `get_group_members` | 获取群组成员 |

#### 用户命令

| 命令 | 说明 |
|------|------|
| `get_user_info` | 获取用户信息 |

## 消息类型 (Kind)

| Kind | 名称 | 持久化 | 说明 |
|------|------|--------|------|
| 1 | 文本消息 | ✅ | 基础消息 |
| 3 | 文件消息 | ✅ | 图片/语音/文件 |
| 5 | 撤销消息 | ✅ | 软删除 |
| 7 | 编辑消息 | ✅ | 编辑已发送消息 |
| 10 | 已读回执 | ✅ | 水位线模式 |
| 11 | 正在输入 | ❌ | 仅转发 |
| 12 | 消息反应 | ✅ | Emoji 回应 |
| 13 | 转发消息 | ✅ | 单条/合并转发 |

## 服务间通信 (JSON-RPC 2.0)

### SeaKing RPC 方法

```
seaking.register              - 用户注册
seaking.login                 - 用户登录
seaking.validateToken         - 验证 JWT Token
seaking.getUserInfo           - 获取用户信息
seaking.checkAccess           - 检查会话访问权限
seaking.getConversation       - 获取会话信息
seaking.getConversationMembers - 获取会话成员
seaking.createConversation    - 创建会话
seaking.getUserConversations  - 获取用户会话列表
seaking.getFriends            - 获取好友列表
seaking.sendFriendRequest     - 发送好友请求
seaking.getPendingFriendRequests - 获取待处理好友请求
seaking.acceptFriendRequest   - 接受好友请求
seaking.rejectFriendRequest   - 拒绝好友请求
seaking.deleteFriend          - 删除好友
seaking.getUserGroups         - 获取用户群组列表
seaking.createGroup           - 创建群组
seaking.getGroupInfo          - 获取群组信息
seaking.getGroupMembers       - 获取群组成员
```

### Relay RPC 方法

```
relay.storeEvent         - 存储事件
relay.getEvent           - 获取事件
relay.queryEvents        - 查询事件
relay.syncEvents         - 同步最新事件
relay.updateReadReceipt  - 更新已读回执
relay.validateRevoke     - 验证撤销权限
relay.validateEdit       - 验证编辑权限
```

## 配置示例

### Gateway 配置

```toml
[ServiceConfiguration]
Name = "gateway"
Port = "8080"
Debug = true

[RedisConfiguration]
Host = "localhost"
Port = "6379"

[LoggerConfiguration]
Level = "debug"
Path = "./logs"
File = "gateway"

[JWTConfiguration]
Secret = "your-jwt-secret"
ExpireHour = 168

[GatewayConfiguration]
WSPath = "/ws"
MaxConnPerUser = 5
HeartbeatTimeout = 30
WriteTimeout = 10
ReadTimeout = 10
SeaKingAddr = "http://localhost:8081"
RelayAddr = "http://localhost:8082"
```

### SeaKing 配置

```toml
[ServiceConfiguration]
Name = "seaking"
Port = "8081"
Debug = true

[PostgresConfiguration]
Host = "localhost"
Port = "5432"
User = "postgres"
Password = "postgres"
Database = "mychat"

[RedisConfiguration]
Host = "localhost"
Port = "6379"

[LoggerConfiguration]
Level = "debug"
Path = "./logs"
File = "seaking"

[JWTConfiguration]
Secret = "your-jwt-secret"
ExpireHour = 168
```

### Relay 配置

```toml
[ServiceConfiguration]
Name = "relay"
Port = "8082"
Debug = true

[PostgresConfiguration]
Host = "localhost"
Port = "5432"
User = "postgres"
Password = "postgres"
Database = "mychat_relay"

[RedisConfiguration]
Host = "localhost"
Port = "6379"

[LoggerConfiguration]
Level = "debug"
Path = "./logs"
File = "relay"

[RelayConfiguration]
MaxEventsPerQuery = 100
RevokeTimeWindow = 120
EditTimeWindow = 86400
```

## 开发进度

### 已完成

- [x] 项目架构搭建
- [x] 协议定义 (MsgPack)
- [x] Gateway WebSocket 管理
- [x] Gateway HTTP API (注册/登录)
- [x] Gateway WebSocket 命令 (好友/群组/会话/用户)
- [x] SeaKing 用户管理
- [x] SeaKing 好友管理
- [x] SeaKing 群组管理
- [x] SeaKing 会话管理
- [x] SeaKing JSON-RPC 服务
- [x] Relay 事件存储
- [x] Relay 消息查询
- [x] Relay 已读回执
- [x] Relay 消息反应
- [x] Relay JSON-RPC 服务
- [x] 服务间 JSON-RPC 通信
- [x] 消息撤销验证 (2分钟窗口)
- [x] 消息编辑验证 (24小时窗口)
- [x] 单元测试

### 待实现

- [ ] 文件管理系统 (S3/OSS)
- [ ] 消息搜索 (Elasticsearch/MeiliSearch)
- [ ] 推送通知
- [ ] 消息加密 (E2E)

## 协议文档

详细协议设计请参考 [IM Rules.md](./IM%20Rules.md)

## 许可证

MIT License
