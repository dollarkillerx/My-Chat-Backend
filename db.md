# My-Chat-Backend 数据库设计文档

## 概述

系统使用 PostgreSQL 数据库，分为两个独立的数据库：

| 数据库 | 服务 | 说明 |
|--------|------|------|
| `mychat` | SeaKing | 用户、好友、群组、会话管理 |
| `mychat_relay` | Relay | 消息事件存储 |

---

## SeaKing 数据库 (mychat)

### 1. users - 用户表

存储用户基本信息。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | VARCHAR(32) | PK | 用户ID (UUID) |
| username | VARCHAR(64) | UNIQUE, NOT NULL | 用户名 |
| nickname | VARCHAR(64) | | 昵称 |
| password | VARCHAR(256) | NOT NULL | 密码 (bcrypt 加密) |
| avatar | VARCHAR(512) | | 头像 URL |
| phone | VARCHAR(32) | INDEX | 手机号 |
| email | VARCHAR(128) | INDEX | 邮箱 |
| status | INTEGER | DEFAULT 1 | 状态: 0=禁用, 1=正常 |
| created_at | TIMESTAMP | DEFAULT NOW | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT NOW | 更新时间 |
| deleted_at | TIMESTAMP | INDEX | 软删除时间 |

**索引:**
- `idx_users_username` (username) - 唯一索引
- `idx_users_phone` (phone)
- `idx_users_email` (email)

**状态枚举:**
```go
UserStatusDisabled = 0  // 禁用
UserStatusNormal   = 1  // 正常
```

---

### 2. user_keys - 用户密钥表

存储用户的公私钥对（用于加密）。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | SERIAL | PK | 自增主键 |
| user_id | VARCHAR(32) | UNIQUE, NOT NULL | 用户ID |
| public_key | TEXT | NOT NULL | 公钥 (明文, Base64) |
| encrypted_private_key | TEXT | NOT NULL | 私钥 (用户密码加密, Base64) |
| key_salt | VARCHAR(64) | NOT NULL | 密钥派生盐值 |
| created_at | TIMESTAMP | DEFAULT NOW | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT NOW | 更新时间 |

**说明:**
- 公钥明文存储，用于加密会话密钥
- 私钥使用用户密码派生的密钥加密存储
- 新设备登录时下载加密私钥，用密码解密
- 算法: RSA-2048 或 X25519

**索引:**
- `idx_user_keys_user` (user_id) - 唯一索引

---

### 3. chat_keys - 私聊加密密钥表

存储私聊会话的对称加密密钥。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | SERIAL | PK | 自增主键 |
| conversation_id | VARCHAR(64) | NOT NULL, INDEX | 会话ID (d:uid1:uid2) |
| user_id | VARCHAR(32) | NOT NULL | 用户ID |
| encrypted_key | TEXT | NOT NULL | 对称密钥 (用该用户公钥加密) |
| created_at | TIMESTAMP | DEFAULT NOW | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT NOW | 更新时间 |

**约束:**
- `UNIQUE(conversation_id, user_id)` - 每用户每会话一条记录

**说明:**
- 首次私聊时，发起方生成对称密钥
- 对称密钥分别用双方公钥加密，各存一份
- 双方用自己私钥解密得到相同的对称密钥

---

### 4. group_keys - 群组加密密钥表

存储群聊的共享对称加密密钥。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | SERIAL | PK | 自增主键 |
| group_id | VARCHAR(32) | NOT NULL, INDEX | 群组ID |
| user_id | VARCHAR(32) | NOT NULL | 用户ID |
| encrypted_key | TEXT | NOT NULL | 群密钥 (用该用户公钥加密) |
| version | INTEGER | DEFAULT 1 | 密钥版本 |
| created_at | TIMESTAMP | DEFAULT NOW | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT NOW | 更新时间 |

**约束:**
- `UNIQUE(group_id, user_id, version)` - 每用户每版本一条记录

**说明:**
- 创建群时，群主生成对称密钥
- 群密钥用每个成员的公钥加密后分别存储
- 所有成员解密后得到相同的群密钥
- 成员退出时可选择更新密钥版本

---

### 5. friendships - 好友关系表

存储用户间的好友关系（双向存储）。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | SERIAL | PK | 自增主键 |
| user_id | VARCHAR(32) | NOT NULL, INDEX | 用户ID |
| friend_id | VARCHAR(32) | NOT NULL, INDEX | 好友ID |
| remark | VARCHAR(64) | | 好友备注 |
| status | INTEGER | DEFAULT 1 | 状态: 1=正常, 2=拉黑 |
| created_at | TIMESTAMP | DEFAULT NOW | 创建时间 |
| updated_at | TIMESTAMP | | 更新时间 |
| deleted_at | TIMESTAMP | INDEX | 软删除时间 |

**约束:**
- `UNIQUE(user_id, friend_id)` - 防止重复好友关系

**索引:**
- `idx_friendships_user` (user_id)
- `idx_friendships_friend` (friend_id)

**状态枚举:**
```go
FriendStatusNormal  = 1  // 正常
FriendStatusBlocked = 2  // 拉黑
```

---

### 3. friend_requests - 好友请求表

存储好友申请记录。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | SERIAL | PK | 自增主键 |
| from_uid | VARCHAR(32) | NOT NULL, INDEX | 发起者ID |
| to_uid | VARCHAR(32) | NOT NULL, INDEX | 接收者ID |
| message | VARCHAR(256) | | 验证消息 |
| status | INTEGER | DEFAULT 0 | 状态: 0=待处理, 1=同意, 2=拒绝 |
| created_at | TIMESTAMP | DEFAULT NOW | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT NOW | 更新时间 |
| deleted_at | TIMESTAMP | INDEX | 软删除时间 |

**索引:**
- `idx_friend_requests_to` (to_uid, status)

**状态枚举:**
```go
FriendRequestPending  = 0  // 待处理
FriendRequestAccepted = 1  // 已同意
FriendRequestRejected = 2  // 已拒绝
```

---

### 4. groups - 群组表

存储群组基本信息。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | VARCHAR(32) | PK | 群组ID (UUID) |
| name | VARCHAR(64) | NOT NULL | 群名称 |
| description | VARCHAR(512) | | 群描述 |
| avatar | VARCHAR(512) | | 群头像 URL |
| owner_id | VARCHAR(32) | NOT NULL, INDEX | 群主ID |
| max_members | INTEGER | DEFAULT 500 | 最大成员数 |
| status | INTEGER | DEFAULT 1 | 状态: 0=解散, 1=正常 |
| created_at | TIMESTAMP | DEFAULT NOW | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT NOW | 更新时间 |
| deleted_at | TIMESTAMP | INDEX | 软删除时间 |

**索引:**
- `idx_groups_owner` (owner_id)

**状态枚举:**
```go
GroupStatusDissolved = 0  // 已解散
GroupStatusNormal    = 1  // 正常
```

---

### 5. group_members - 群成员表

存储群组成员关系。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | SERIAL | PK | 自增主键 |
| group_id | VARCHAR(32) | NOT NULL, INDEX | 群组ID |
| user_id | VARCHAR(32) | NOT NULL, INDEX | 用户ID |
| role | INTEGER | DEFAULT 0 | 角色: 0=成员, 1=管理员, 2=群主 |
| nickname | VARCHAR(64) | | 群内昵称 |
| muted | BOOLEAN | DEFAULT FALSE | 是否被禁言 |
| muted_at | TIMESTAMP | | 禁言时间 |
| joined_at | TIMESTAMP | DEFAULT NOW | 加入时间 |
| created_at | TIMESTAMP | | 创建时间 |
| updated_at | TIMESTAMP | | 更新时间 |
| deleted_at | TIMESTAMP | INDEX | 软删除时间 |

**约束:**
- `UNIQUE(group_id, user_id)` - 防止重复加入

**索引:**
- `idx_group_members_group` (group_id)
- `idx_group_members_user` (user_id)

**角色枚举:**
```go
GroupRoleMember = 0  // 普通成员
GroupRoleAdmin  = 1  // 管理员
GroupRoleOwner  = 2  // 群主
```

---

### 6. conversations - 会话表

存储会话（聊天）信息。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | VARCHAR(64) | PK | 会话ID (cid) |
| type | INTEGER | NOT NULL | 类型: 1=单聊, 2=群聊 |
| name | VARCHAR(64) | | 会话名称 |
| avatar | VARCHAR(512) | | 会话头像 |
| created_at | TIMESTAMP | DEFAULT NOW | 创建时间 |
| updated_at | TIMESTAMP | DEFAULT NOW | 更新时间 |
| deleted_at | TIMESTAMP | INDEX | 软删除时间 |

**会话ID生成规则:**
```go
// 单聊: d:{uid1}:{uid2} (uid 按字典序排列)
GenerateDirectCid("user1", "user2") // "d:user1:user2"

// 群聊: g:{group_id}
GenerateGroupCid("group123") // "g:group123"
```

**类型枚举:**
```go
ConversationTypeDirect = 1  // 单聊
ConversationTypeGroup  = 2  // 群聊
```

---

### 7. conversation_members - 会话成员表

存储会话参与者信息。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | SERIAL | PK | 自增主键 |
| conversation_id | VARCHAR(64) | NOT NULL, INDEX | 会话ID |
| user_id | VARCHAR(32) | NOT NULL, INDEX | 用户ID |
| role | INTEGER | DEFAULT 0 | 角色 |
| last_read_mid | BIGINT | DEFAULT 0 | 最后已读消息ID |
| muted | BOOLEAN | DEFAULT FALSE | 是否免打扰 |
| pinned | BOOLEAN | DEFAULT FALSE | 是否置顶 |
| joined_at | TIMESTAMP | DEFAULT NOW | 加入时间 |
| created_at | TIMESTAMP | | 创建时间 |
| updated_at | TIMESTAMP | | 更新时间 |
| deleted_at | TIMESTAMP | INDEX | 软删除时间 |

**约束:**
- `UNIQUE(conversation_id, user_id)` - 防止重复加入

**索引:**
- `idx_conv_members_conv` (conversation_id)
- `idx_conv_members_user` (user_id)

---

## 表关系说明：groups / group_members / conversations / conversation_members

### 关系图

```
┌─────────┐          ┌───────────────┐
│ groups  │──────────│ group_members │──────────┐
└─────────┘   1:N    └───────────────┘          │
     │                                          │
     │ 创建群聊时                                 │ 同一个用户
     │ 自动生成                                   │
     ▼                                          ▼
┌───────────────┐    ┌──────────────────────┐
│ conversations │────│ conversation_members │
└───────────────┘ 1:N└──────────────────────┘
```

### 职责区分

| 表 | 职责 | 场景 |
|---|---|---|
| **groups** | 群组元信息（名称、头像、群主） | 群组管理 |
| **group_members** | 群成员 + 角色（群主/管理员/成员） | 群权限控制 |
| **conversations** | 会话（单聊/群聊）统一抽象 | 消息路由 |
| **conversation_members** | 会话成员 + 已读/置顶/免打扰 | 用户会话列表 |

### 为什么要分开？

1. **单聊没有群组**：单聊只需要 `conversations` + `conversation_members`，不涉及 `groups`

2. **职责分离**：
   - `group_members.role` → 管理权限（谁能踢人、禁言）
   - `conversation_members` → 消息状态（已读水位、是否置顶）

3. **会话ID生成规则**：
   ```go
   // 单聊: d:{uid1}:{uid2}
   // 群聊: g:{group_id}  ← 群聊会话与群组关联
   ```

### 数据流示例

**创建群聊时：**
```
1. 插入 groups (id=abc123, name="技术群")
2. 插入 group_members (group_id=abc123, user_id=xxx, role=2群主)
3. 插入 conversations (id="g:abc123", type=2群聊)
4. 插入 conversation_members (conversation_id="g:abc123", user_id=xxx)
```

群聊场景下 `group_members` 和 `conversation_members` 的成员是同步的，但存储的信息不同。

---

## Relay 数据库 (mychat_relay)

### 1. events - 消息事件表

存储所有消息事件。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| mid | BIGSERIAL | PK | 消息ID (自增) |
| cid | VARCHAR(64) | NOT NULL, INDEX | 会话ID |
| kind | INTEGER | NOT NULL, INDEX | 消息类型 |
| sender | VARCHAR(32) | NOT NULL, INDEX | 发送者ID |
| timestamp | BIGINT | NOT NULL, INDEX | 时间戳 (秒) |
| flags | INTEGER | DEFAULT 0 | 标志位 |
| tags | JSONB | | 标签数组 |
| data | JSONB | | 消息内容 |
| sig | VARCHAR(256) | | 签名 |
| ext | JSONB | | 扩展字段 |
| created_at | TIMESTAMP | DEFAULT NOW | 创建时间 |
| deleted_at | TIMESTAMP | INDEX | 软删除时间 |

**索引:**
- `idx_events_cid_mid` (cid, mid) - 会话消息查询
- `idx_events_cid_timestamp` (cid, timestamp DESC) - 时间顺序查询
- `idx_events_sender` (sender)
- `idx_events_kind` (kind)

**消息类型 (Kind):**
```go
KindText       = 1   // 文本消息
KindFile       = 3   // 文件消息
KindRevoke     = 5   // 撤销消息
KindEdit       = 7   // 编辑消息
KindReadReceipt = 10 // 已读回执
KindTyping     = 11  // 正在输入
KindReaction   = 12  // 消息反应
KindForward    = 13  // 转发消息
```

---

### 2. read_receipts - 已读回执表

存储用户在各会话的已读状态（水位线模式）。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | SERIAL | PK | 自增主键 |
| cid | VARCHAR(64) | NOT NULL, INDEX | 会话ID |
| uid | VARCHAR(32) | NOT NULL | 用户ID |
| last_read_mid | BIGINT | NOT NULL | 最后已读消息ID |
| updated_at | TIMESTAMP | DEFAULT NOW | 更新时间 |

**约束:**
- `UNIQUE(cid, uid)` - 每用户每会话一条记录

**索引:**
- `idx_receipts_cid` (cid)

**与 conversation_members.last_read_mid 的关系:**

两者存储相同数据但服务于不同目的：

```
┌─────────────────┐         ┌─────────────────┐
│    SeaKing      │         │     Relay       │
│   (用户/会话)    │         │    (消息)       │
├─────────────────┤         ├─────────────────┤
│ conversation_   │  同步   │  read_receipts  │
│ members         │◄───────►│                 │
│ .last_read_mid  │         │ .last_read_mid  │
└─────────────────┘         └─────────────────┘
       │                           │
       ▼                           ▼
  "你有3条未读"              "对方已读" ✓✓
```

| 属性 | conversation_members | read_receipts |
|------|---------------------|---------------|
| **所属数据库** | mychat (SeaKing) | mychat_relay (Relay) |
| **职责** | 会话列表展示 | 消息已读回执 |
| **用途** | 计算未读数、置顶、免打扰 | 显示"对方已读"、群消息已读人数 |

**设计理由:**
1. **服务隔离** - SeaKing 和 Relay 是独立服务，各自维护所需数据，避免跨库查询
2. **查询优化** - 获取会话列表只查 SeaKing，查看消息详情只查 Relay
3. **性能考虑** - `conversation_members.last_read_mid` 是缓存副本，避免跨服务调用

---

### 3. reactions - 消息反应表

存储消息的 Emoji 反应。

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| id | SERIAL | PK | 自增主键 |
| cid | VARCHAR(64) | NOT NULL, INDEX | 会话ID |
| mid | BIGINT | NOT NULL, INDEX | 目标消息ID |
| uid | VARCHAR(32) | NOT NULL | 用户ID |
| emoji | VARCHAR(32) | NOT NULL | Emoji 表情 |
| created_at | TIMESTAMP | DEFAULT NOW | 创建时间 |
| deleted_at | TIMESTAMP | INDEX | 软删除时间 |

**约束:**
- `UNIQUE(cid, mid, uid, emoji)` - 同一用户同一消息同一表情只能反应一次

**索引:**
- `idx_reactions_mid` (mid)
- `idx_reactions_cid_mid` (cid, mid)

---

## ER 图

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            mychat (SeaKing)                                  │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────┐     ┌──────────────┐     ┌─────────────────┐                  │
│  │  users  │────<│ friendships  │>────│  users (friend) │                  │
│  └────┬────┘     └──────────────┘     └─────────────────┘                  │
│       │                                                                      │
│       │          ┌────────────────┐                                         │
│       └─────────<│friend_requests │                                         │
│       │          └────────────────┘                                         │
│       │                                                                      │
│       │          ┌─────────┐     ┌───────────────┐                          │
│       └─────────<│ groups  │────<│ group_members │>────┐                    │
│       │          └─────────┘     └───────────────┘     │                    │
│       │                                                │                    │
│       │          ┌───────────────┐     ┌──────────────────────┐            │
│       └─────────<│ conversations │────<│ conversation_members │>───────────┘│
│                  └───────────────┘     └──────────────────────┘             │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────────┐
│                           mychat_relay (Relay)                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌──────────┐     ┌───────────────┐     ┌────────────┐                     │
│  │  events  │────<│ read_receipts │     │ reactions  │>────┐               │
│  └────┬─────┘     └───────────────┘     └────────────┘     │               │
│       │                                        │           │               │
│       └────────────────────────────────────────┴───────────┘               │
│                              (by mid)                                       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 消息加密设计

### 架构概述

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              简化加密架构                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────┐              ┌──────────────────────┐              ┌─────────────┐
│  │   用户 A    │              │        服务器         │              │   用户 B    │
│  ├─────────────┤              ├──────────────────────┤              ├─────────────┤
│  │ 公钥 A      │─────────────>│ 公钥 A (明文)        │              │ 公钥 B      │
│  │ 私钥 A      │─────────────>│ 私钥 A (密码加密)    │              │ 私钥 B      │
│  │ (本地缓存)   │              │ 私钥 B (密码加密)    │<─────────────│ (本地缓存)   │
│  └─────────────┘              │ 公钥 B (明文)        │              └─────────────┘
│                               │                      │
│                               │ 对称密钥 (公钥加密)   │
│                               │ 加密消息             │
│                               └──────────────────────┘
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 密钥类型

| 密钥类型 | 存储位置 | 加密方式 | 用途 |
|---------|---------|---------|------|
| **用户公钥** | 服务器 | 明文 | 加密对称密钥 |
| **用户私钥** | 服务器 | 用户密码加密 | 解密对称密钥 |
| **私聊对称密钥** | 服务器 | 用户公钥加密 | 加密私聊消息 |
| **群聊对称密钥** | 服务器 | 用户公钥加密 | 加密群聊消息 |

### 用户注册流程

```
1. 客户端生成密钥对
   - 生成 RSA-2048 公私钥对
   - 公钥: public_key
   - 私钥: private_key

2. 加密私钥
   - 使用密码派生密钥: derived_key = PBKDF2(password, salt)
   - 加密私钥: encrypted_private_key = AES-GCM(private_key, derived_key)

3. 上传到服务器
   - user_keys.public_key = public_key
   - user_keys.encrypted_private_key = encrypted_private_key
   - user_keys.key_salt = salt
```

### 新设备登录流程

```
1. 用户输入账号密码登录

2. 从服务器下载:
   - encrypted_private_key
   - key_salt

3. 解密私钥:
   - derived_key = PBKDF2(password, key_salt)
   - private_key = AES-GCM-Decrypt(encrypted_private_key, derived_key)

4. 本地缓存私钥用于后续解密
```

### 私聊加密流程

```
Alice 首次给 Bob 发消息:

1. Alice 生成对称密钥:
   chat_key = random(256 bits)

2. 加密对称密钥并存储:
   - 用 Alice 公钥加密: encrypted_key_a = RSA(chat_key, alice_public_key)
   - 用 Bob 公钥加密:   encrypted_key_b = RSA(chat_key, bob_public_key)
   - 存入 chat_keys 表 (各存一条)

3. 发送消息:
   - 用 chat_key 加密消息: ciphertext = AES-GCM(message, chat_key)
   - 存入 events 表

Bob 收到消息:

4. 从 chat_keys 获取 encrypted_key_b
5. 用自己私钥解密: chat_key = RSA-Decrypt(encrypted_key_b, bob_private_key)
6. 解密消息: message = AES-GCM-Decrypt(ciphertext, chat_key)
```

### 群聊加密流程

```
创建群组时:

1. 群主生成群对称密钥:
   group_key = random(256 bits)

2. 为每个成员加密群密钥:
   FOR each member:
     encrypted_key = RSA(group_key, member_public_key)
     INSERT INTO group_keys (group_id, user_id, encrypted_key)

发送群消息:

3. 用 group_key 加密: ciphertext = AES-GCM(message, group_key)
4. 存入 events 表

新成员加入:

5. 管理员获取当前 group_key
6. 用新成员公钥加密后存入 group_keys

成员退出 (可选更新密钥):

7. 生成新 group_key，version + 1
8. 重新为剩余成员加密分发
```

### events 表加密消息格式

```json
{
  "encrypted": true,
  "algorithm": "aes-256-gcm",
  "ciphertext": "base64编码的加密内容",
  "nonce": "base64编码的随机数"
}
```

### 加密算法选择

| 用途 | 算法 | 说明 |
|-----|------|------|
| 用户密钥对 | RSA-2048 | 简单易实现，性能可接受 |
| 私钥加密 | AES-256-GCM + PBKDF2 | 使用密码派生密钥 |
| 消息加密 | AES-256-GCM | 对称加密，高性能 |

### 安全注意事项

1. **密码强度** - 私钥安全依赖用户密码，建议强制密码复杂度
2. **私钥缓存** - 客户端应安全存储解密后的私钥（如 Keychain/KeyStore）
3. **密钥备份** - 用户忘记密码将无法解密私钥，可考虑恢复码机制
4. **传输安全** - 所有通信必须使用 TLS
5. **群密钥更新** - 敏感群可在成员退出时更新群密钥

---

