# My-Chat-Backend

基于 Nostr 协议思想的中心化即时通讯后端系统，支持端到端加密。

## 特性

- 参考 Nostr 协议的事件模型，轻量可扩展
- 端到端加密（E2EE）- 服务端无法解密消息内容
- 支持单聊、群聊、文件、撤销、编辑、转发等完整 IM 功能
- JSON-RPC 2.0 + WebSocket 通信
- 微服务架构，可横向扩展

## 架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                          Client                                  │
│                    (Mobile / Web / Desktop)                      │
│              [生成密钥对 / 加密消息 / 解密消息]                    │
└─────────────────────────────────────────────────────────────────┘
                    │                       │
        JSON-RPC    │                       │ WebSocket
    (register/login/│                       │ (加密消息推送)
     friends/groups)│                       │
                    ▼                       ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Gateway 集群                             │
│              (JSON-RPC接口 / WebSocket消息推送 / 鉴权)            │
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
│  • 密钥管理 🔐            │   │  • 加密消息 (无法解密)          │
└───────────────────────────┘   └───────────────────────────────┘
            │                               │
            ▼                               ▼
┌───────────────────────────┐   ┌───────────────────────────────┐
│        PostgreSQL         │   │         PostgreSQL            │
│       (mychat DB)         │   │       (mychat_relay DB)       │
│  • 用户公钥 (明文)         │   │  • 加密消息内容                │
│  • 用户私钥 (密码加密)     │   │                               │
│  • 会话密钥 (公钥加密)     │   │                               │
└───────────────────────────┘   └───────────────────────────────┘
```

## 通信架构

本系统使用两种通信方式：

1. **JSON-RPC 2.0**: 所有业务操作（注册、登录、好友、群组、会话等）
2. **WebSocket**: 仅用于实时消息推送和接收

**注意**:
- 客户端通过 Gateway 的 JSON-RPC 接口进行业务操作
- 客户端通过 Gateway 的 WebSocket 接收消息推送
- SeaKing 和 Relay 是内部服务，不直接暴露给客户端

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
│       ├── handler/      # WebSocket消息处理
│       ├── rpc/          # JSON-RPC处理
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

### 测试覆盖率

| 模块 | 覆盖率 | 说明 |
|------|--------|------|
| common/pkg/crypto | 80.5% | RSA/AES 加密工具 |
| seaking/internal/model | 100% | 数据模型 |
| seaking/internal/service | - | 业务逻辑 |

测试文件列表:
- `common/pkg/crypto/rsa_test.go` - RSA 加密测试 (18 tests)
- `common/pkg/crypto/aes_test.go` - AES 加密测试 (14 tests)
- `seaking/internal/model/key_test.go` - 密钥模型测试
- `seaking/internal/model/user_test.go` - 用户模型测试
- `seaking/internal/model/group_test.go` - 群组模型测试
- `seaking/internal/model/relation_test.go` - 好友关系模型测试
- `seaking/internal/model/conversation_test.go` - 会话模型测试
- `seaking/internal/service/key/service_test.go` - 密钥服务测试
- `seaking/internal/service/conversation/conversation_test.go` - 会话服务测试
- `seaking/internal/service/relation/relation_test.go` - 好友服务测试

## 服务端口

| 服务 | 端口 | 说明 |
|------|------|------|
| Gateway | 8080 | JSON-RPC + WebSocket |
| SeaKing | 8081 | JSON-RPC (内部) |
| Relay | 8082 | JSON-RPC (内部) |

## Gateway 客户端接口

### JSON-RPC 接口

端点: `POST /api/rpc`

请求格式:
```json
{
    "jsonrpc": "2.0",
    "method": "方法名",
    "params": { ... },
    "id": 1
}
```

#### 认证相关（无需Token）

| 方法 | 说明 | 参数 |
|------|------|------|
| `register` | 用户注册 | `username`, `password`, `nickname`, `public_key`, `encrypted_private_key`, `key_salt`, `phone?`, `email?` |
| `login` | 用户登录 | `username`, `password`, `device_id`, `platform` |

**注意**: 注册时客户端需生成 RSA-2048 密钥对，私钥使用密码加密后上传。

#### 用户相关（需要Token）

| 方法 | 说明 | 参数 |
|------|------|------|
| `getUserInfo` | 获取用户信息 | `uid?` (不传则获取自己) |

#### 好友相关（需要Token）

| 方法 | 说明 | 参数 |
|------|------|------|
| `getFriends` | 获取好友列表 | 无 |
| `sendFriendRequest` | 发送好友请求 | `to_uid`, `message?` |
| `getPendingFriendRequests` | 获取待处理好友请求 | 无 |
| `acceptFriendRequest` | 接受好友请求 | `request_id` |
| `rejectFriendRequest` | 拒绝好友请求 | `request_id` |
| `deleteFriend` | 删除好友 | `friend_id` |

#### 会话相关（需要Token）

| 方法 | 说明 | 参数 |
|------|------|------|
| `getConversations` | 获取会话列表 | 无 |
| `createConversation` | 创建会话 | `type`, `member_ids`, `name?` |
| `getConversationMembers` | 获取会话成员 | `cid` |

#### 群组相关（需要Token）

| 方法 | 说明 | 参数 |
|------|------|------|
| `getGroups` | 获取群组列表 | 无 |
| `createGroup` | 创建群组 | `name`, `description?`, `member_ids?` |
| `getGroupInfo` | 获取群组信息 | `group_id` |
| `getGroupMembers` | 获取群组成员 | `group_id` |

#### 加密相关（需要Token）

| 方法 | 说明 | 参数 |
|------|------|------|
| `getUserPublicKey` | 获取用户公钥 | `uid` |
| `getChatKey` | 获取私聊会话密钥 | `cid` |
| `createChatKey` | 创建私聊会话密钥 | `cid`, `keys[]` |
| `getGroupKey` | 获取群组密钥 | `group_id`, `version?` |
| `createGroupKey` | 创建/更新群组密钥 | `group_id`, `keys[]`, `version` |

### 示例

注册:
```json
{
    "jsonrpc": "2.0",
    "method": "register",
    "params": {
        "username": "user1",
        "password": "password123",
        "nickname": "User One"
    },
    "id": 1
}
```

登录:
```json
{
    "jsonrpc": "2.0",
    "method": "login",
    "params": {
        "username": "user1",
        "password": "password123",
        "device_id": "device-uuid",
        "platform": "ios"
    },
    "id": 2
}
```

获取好友列表（需要在Header中带Token）:
```
Authorization: Bearer <token>
```
```json
{
    "jsonrpc": "2.0",
    "method": "getFriends",
    "params": {},
    "id": 3
}
```

### WebSocket 接口

连接地址: `ws://localhost:8080/ws?token=<JWT_TOKEN>`

WebSocket 仅用于消息推送相关操作:

| 命令 | 说明 | 方向 |
|------|------|------|
| `ping` | 心跳请求 | C -> S |
| `pong` | 心跳响应 | S -> C |
| `event` | 事件消息 | 双向 |
| `ack` | 消息确认 | S -> C |
| `error` | 错误响应 | S -> C |
| `subscribe` | 订阅会话 | C -> S |
| `unsubscribe` | 取消订阅 | C -> S |
| `sync` | 同步历史消息 | C -> S |

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

## 端到端加密 (E2EE)

### 加密架构

```
用户注册: 客户端生成密钥对 → 公钥明文上传 → 私钥用密码加密后上传
新设备登录: 下载加密私钥 → 用密码解密 → 本地缓存
私聊: 生成会话密钥 → 用双方公钥各加密一份 → 消息用会话密钥加密
群聊: 生成群密钥 → 用每个成员公钥加密一份 → 消息用群密钥加密
```

### 加密算法

| 用途 | 算法 |
|-----|------|
| 用户密钥对 | RSA-2048 |
| 私钥保护 | AES-256-GCM + PBKDF2 |
| 消息加密 | AES-256-GCM |

### 加密消息格式

```json
{
  "encrypted": true,
  "ciphertext": "base64...",
  "nonce": "base64..."
}
```

### 安全特性

- **服务端零知识**: 服务端只存储加密数据，无法解密消息内容
- **多设备同步**: 私钥加密存储在服务端，新设备登录后用密码解密
- **会话密钥隔离**: 每个会话独立密钥，互不影响
- **可选密钥轮换**: 群成员退出时可更新群密钥

## 内部服务通信 (JSON-RPC 2.0)

### SeaKing RPC 方法

```
# 认证
seaking.register              - 用户注册 (含密钥上传)
seaking.login                 - 用户登录 (返回加密私钥)
seaking.validateToken         - 验证 JWT Token

# 用户
seaking.getUserInfo           - 获取用户信息
seaking.getUserPublicKey      - 获取用户公钥

# 好友
seaking.getFriends            - 获取好友列表
seaking.sendFriendRequest     - 发送好友请求
seaking.getPendingFriendRequests - 获取待处理好友请求
seaking.acceptFriendRequest   - 接受好友请求
seaking.rejectFriendRequest   - 拒绝好友请求
seaking.deleteFriend          - 删除好友

# 群组
seaking.getUserGroups         - 获取用户群组列表
seaking.createGroup           - 创建群组
seaking.getGroupInfo          - 获取群组信息
seaking.getGroupMembers       - 获取群组成员

# 会话
seaking.checkAccess           - 检查会话访问权限
seaking.getConversation       - 获取会话信息
seaking.getConversationMembers - 获取会话成员
seaking.createConversation    - 创建会话
seaking.getUserConversations  - 获取用户会话列表

# 加密密钥
seaking.getChatKey            - 获取私聊会话密钥
seaking.createChatKey         - 创建私聊会话密钥
seaking.getGroupKey           - 获取群组密钥
seaking.createGroupKey        - 创建/更新群组密钥
seaking.getMemberPublicKeys   - 批量获取成员公钥
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
- [x] Gateway JSON-RPC 接口
- [x] Gateway WebSocket 消息推送
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
- [x] 端到端加密 - 用户密钥管理
- [x] 端到端加密 - 私聊会话密钥
- [x] 端到端加密 - 群聊密钥分发
- [x] 端到端加密 - 密钥轮换

### 待实现

- [ ] 客户端 SDK 实现
- [ ] 消息推送集成

## 文档

- [IM Rules.md](./IM%20Rules.md) - 协议设计文档（消息类型、加密流程、交互协议）
- [db.md](./db.md) - 数据库设计文档（表结构、索引、关系说明）

## 许可证

MIT License
