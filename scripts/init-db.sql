-- My-Chat-Backend 数据库初始化脚本
-- 注意：这个脚本在 postgres 容器启动时执行

-- ============================================
-- 创建数据库
-- ============================================

-- 创建 relay 使用的数据库
CREATE DATABASE mychat_relay;

-- 授权
GRANT ALL PRIVILEGES ON DATABASE mychat TO postgres;
GRANT ALL PRIVILEGES ON DATABASE mychat_relay TO postgres;

-- ============================================
-- SeaKing 数据库表结构 (mychat)
-- ============================================

\c mychat;

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(32) PRIMARY KEY,
    username VARCHAR(64) UNIQUE NOT NULL,
    nickname VARCHAR(64),
    password VARCHAR(256) NOT NULL,
    avatar VARCHAR(512),
    phone VARCHAR(32),
    email VARCHAR(128),
    status INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_phone ON users(phone);
CREATE INDEX idx_users_email ON users(email);

-- 好友关系表
CREATE TABLE IF NOT EXISTS friendships (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(32) NOT NULL,
    friend_id VARCHAR(32) NOT NULL,
    status INTEGER DEFAULT 1,
    remark VARCHAR(64),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, friend_id)
);

CREATE INDEX idx_friendships_user ON friendships(user_id);
CREATE INDEX idx_friendships_friend ON friendships(friend_id);

-- 好友请求表
CREATE TABLE IF NOT EXISTS friend_requests (
    id SERIAL PRIMARY KEY,
    from_uid VARCHAR(32) NOT NULL,
    to_uid VARCHAR(32) NOT NULL,
    message VARCHAR(256),
    status INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_friend_requests_to ON friend_requests(to_uid, status);

-- 群组表
CREATE TABLE IF NOT EXISTS groups (
    id VARCHAR(32) PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    description VARCHAR(256),
    avatar VARCHAR(512),
    owner_id VARCHAR(32) NOT NULL,
    max_members INTEGER DEFAULT 500,
    status INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_groups_owner ON groups(owner_id);

-- 群成员表
CREATE TABLE IF NOT EXISTS group_members (
    id SERIAL PRIMARY KEY,
    group_id VARCHAR(32) NOT NULL,
    user_id VARCHAR(32) NOT NULL,
    role INTEGER DEFAULT 0,
    nickname VARCHAR(64),
    muted BOOLEAN DEFAULT FALSE,
    muted_at TIMESTAMP,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(group_id, user_id)
);

CREATE INDEX idx_group_members_group ON group_members(group_id);
CREATE INDEX idx_group_members_user ON group_members(user_id);

-- 会话表
CREATE TABLE IF NOT EXISTS conversations (
    id VARCHAR(32) PRIMARY KEY,
    type INTEGER NOT NULL,
    name VARCHAR(64),
    avatar VARCHAR(512),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 会话成员表
CREATE TABLE IF NOT EXISTS conversation_members (
    id SERIAL PRIMARY KEY,
    conversation_id VARCHAR(32) NOT NULL,
    user_id VARCHAR(32) NOT NULL,
    role INTEGER DEFAULT 0,
    muted BOOLEAN DEFAULT FALSE,
    pinned BOOLEAN DEFAULT FALSE,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(conversation_id, user_id)
);

CREATE INDEX idx_conv_members_conv ON conversation_members(conversation_id);
CREATE INDEX idx_conv_members_user ON conversation_members(user_id);

-- 用户密钥表 (加密)
CREATE TABLE IF NOT EXISTS user_keys (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(32) UNIQUE NOT NULL,
    public_key TEXT NOT NULL,
    encrypted_private_key TEXT NOT NULL,
    key_salt VARCHAR(64) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_keys_user ON user_keys(user_id);

-- 私聊密钥表 (加密)
CREATE TABLE IF NOT EXISTS chat_keys (
    id SERIAL PRIMARY KEY,
    conversation_id VARCHAR(64) NOT NULL,
    user_id VARCHAR(32) NOT NULL,
    encrypted_key TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(conversation_id, user_id)
);

CREATE INDEX idx_chat_keys_conv ON chat_keys(conversation_id);

-- 群组密钥表 (加密)
CREATE TABLE IF NOT EXISTS group_keys (
    id SERIAL PRIMARY KEY,
    group_id VARCHAR(32) NOT NULL,
    user_id VARCHAR(32) NOT NULL,
    encrypted_key TEXT NOT NULL,
    version INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(group_id, user_id, version)
);

CREATE INDEX idx_group_keys_group ON group_keys(group_id);
CREATE INDEX idx_group_keys_user ON group_keys(user_id);

-- ============================================
-- Relay 数据库表结构 (mychat_relay)
-- ============================================

\c mychat_relay;

-- 事件表
CREATE TABLE IF NOT EXISTS events (
    mid BIGSERIAL PRIMARY KEY,
    cid VARCHAR(32) NOT NULL,
    kind INTEGER NOT NULL,
    sender VARCHAR(32) NOT NULL,
    timestamp BIGINT NOT NULL,
    flags INTEGER DEFAULT 0,
    tags JSONB,
    data JSONB,
    sig VARCHAR(256),
    ext JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_events_cid_mid ON events(cid, mid);
CREATE INDEX idx_events_cid_timestamp ON events(cid, timestamp DESC);
CREATE INDEX idx_events_sender ON events(sender);
CREATE INDEX idx_events_kind ON events(kind);

-- 已读回执表
CREATE TABLE IF NOT EXISTS read_receipts (
    id SERIAL PRIMARY KEY,
    cid VARCHAR(32) NOT NULL,
    uid VARCHAR(32) NOT NULL,
    last_read_mid BIGINT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(cid, uid)
);

CREATE INDEX idx_receipts_cid ON read_receipts(cid);

-- 消息反应表
CREATE TABLE IF NOT EXISTS reactions (
    id SERIAL PRIMARY KEY,
    cid VARCHAR(32) NOT NULL,
    mid BIGINT NOT NULL,
    uid VARCHAR(32) NOT NULL,
    emoji VARCHAR(32) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(cid, mid, uid, emoji)
);

CREATE INDEX idx_reactions_mid ON reactions(mid);
CREATE INDEX idx_reactions_cid_mid ON reactions(cid, mid);

-- ============================================
-- 完成
-- ============================================

\echo 'Database initialization completed!'
