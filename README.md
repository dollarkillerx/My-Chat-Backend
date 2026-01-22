# My-Chat-Backend

基于 Nostr 协议思想的中心化即时通讯后端系统。

## 架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                          Client                                  │
│                    (Mobile / Web / Desktop)                      │
└─────────────────────────────────────────────────────────────────┘
                                │
                                │ WebSocket
                                ▼
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
│       ├── api/          # REST API
│       ├── conf/
│       ├── model/        # 数据模型
│       ├── rpc/          # JSON-RPC
│       ├── server/
│       ├── service/      # 业务逻辑
│       └── storage/
├── relay/                # 事件存储服务
│   ├── cmd/
│   └── internal/
│       ├── api/
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
| Gateway | 8080 | WebSocket 入口 |
| SeaKing | 8081 | 用户中心 API |
| Relay | 8082 | 事件存储 API |

## API 文档

### SeaKing API

#### 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/register` | 用户注册 |
| POST | `/api/v1/login` | 用户登录 |

#### 用户接口 (需认证)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/profile` | 获取个人资料 |
| PUT | `/api/v1/profile` | 更新个人资料 |
| PUT | `/api/v1/password` | 修改密码 |

#### 好友接口 (需认证)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/friends` | 获取好友列表 |
| POST | `/api/v1/friends/request` | 发送好友请求 |
| POST | `/api/v1/friends/accept` | 接受好友请求 |
| POST | `/api/v1/friends/reject` | 拒绝好友请求 |
| DELETE | `/api/v1/friends/:uid` | 删除好友 |
| POST | `/api/v1/friends/block` | 拉黑好友 |
| POST | `/api/v1/friends/unblock` | 取消拉黑 |
| GET | `/api/v1/friends/requests` | 待处理请求列表 |

#### 群组接口 (需认证)

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/groups` | 获取群组列表 |
| POST | `/api/v1/groups` | 创建群组 |
| GET | `/api/v1/groups/:id` | 获取群组详情 |
| PUT | `/api/v1/groups/:id` | 更新群组信息 |
| DELETE | `/api/v1/groups/:id` | 解散群组 |
| GET | `/api/v1/groups/:id/members` | 获取群成员 |
| POST | `/api/v1/groups/:id/members` | 添加成员 |
| DELETE | `/api/v1/groups/:id/members/:uid` | 移除成员 |
| POST | `/api/v1/groups/:id/leave` | 退出群组 |
| POST | `/api/v1/groups/:id/transfer` | 转让群主 |
| POST | `/api/v1/groups/:id/admin` | 设置管理员 |

### Relay API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/events` | 存储事件 |
| GET | `/api/v1/events/:mid` | 获取单个事件 |
| POST | `/api/v1/events/query` | 查询事件 |
| GET | `/api/v1/events/sync` | 同步最新事件 |
| POST | `/api/v1/receipts` | 更新已读回执 |
| GET | `/api/v1/receipts` | 获取已读回执 |
| POST | `/api/v1/reactions` | 添加消息反应 |
| DELETE | `/api/v1/reactions` | 移除消息反应 |
| GET | `/api/v1/reactions/:mid` | 获取消息反应 |

### Gateway WebSocket

连接地址: `ws://localhost:8080/ws?token=<JWT_TOKEN>`

#### 命令类型

| 命令 | 说明 |
|------|------|
| `ping` | 心跳请求 |
| `pong` | 心跳响应 |
| `event` | 事件消息 |
| `ack` | 消息确认 |
| `error` | 错误响应 |
| `subscribe` | 订阅会话 |
| `unsubscribe` | 取消订阅 |
| `sync` | 同步历史消息 |

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

## 服务间通信

服务间使用 JSON-RPC 2.0 协议通信：

### SeaKing RPC 方法

```
seaking.checkAccess       - 检查会话访问权限
seaking.getConversation   - 获取会话信息
seaking.getConversationMembers - 获取会话成员
seaking.createConversation - 创建会话
seaking.getUserConversations - 获取用户会话列表
seaking.validateToken     - 验证 JWT Token
seaking.getUserInfo       - 获取用户信息
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
SeaKingAddr = "http://localhost:8081/api/rpc"
RelayAddr = "http://localhost:8082/api/rpc"
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
```

## 开发进度

### 已完成 ✅

- [x] 项目架构搭建
- [x] 协议定义 (MsgPack)
- [x] Gateway WebSocket 管理
- [x] SeaKing 用户管理
- [x] SeaKing 好友管理
- [x] SeaKing 群组管理
- [x] SeaKing 会话管理
- [x] Relay 事件存储
- [x] Relay 消息查询
- [x] Relay 已读回执
- [x] Relay 消息反应
- [x] 服务间 JSON-RPC 通信
- [x] 消息撤销验证 (2分钟窗口)
- [x] 消息编辑验证 (24小时窗口)
- [x] 单元测试

### 待实现 🚧

- [ ] 文件管理系统 (S3/OSS)
- [ ] 消息搜索 (Elasticsearch/MeiliSearch)
- [ ] 推送通知
- [ ] 消息加密 (E2E)

## 协议文档

详细协议设计请参考 [IM Rules.md](./IM%20Rules.md)

## 许可证

MIT License
