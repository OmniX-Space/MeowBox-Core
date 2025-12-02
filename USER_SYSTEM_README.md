# 用户系统使用指南

## 新功能概述

已为 Meow Embedded Music Server 添加完整的用户系统，支持：

### ✨ 核心功能

1. **用户认证**
   - 用户注册（用户名 + 邮箱 + 密码）
   - 用户登录
   - 会话管理（基于Token）
   - 安全的密码加密存储

2. **个人歌单管理**
   - 每个用户独立的"我喜欢"歌单
   - 创建自定义歌单
   - 添加/删除歌曲到歌单
   - 删除自定义歌单

3. **音乐搜索和播放**
   - 在线搜索音乐
   - 实时播放音乐
   - 添加到个人歌单
   - 查看歌单歌曲列表

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 启动服务器

```bash
go run .
```

### 3. 访问应用

打开浏览器访问：
- **新版应用（推荐）**: `http://localhost:2233/app`
- **旧版界面（兼容）**: `http://localhost:2233/`

## API 文档

### 认证相关

#### 注册
```http
POST /api/auth/register
Content-Type: application/json

{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}
```

**响应**:
```json
{
  "token": "session_token_here",
  "user": {
    "id": "user_id",
    "username": "testuser",
    "email": "test@example.com",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

#### 登录
```http
POST /api/auth/login
Content-Type: application/json

{
  "username": "testuser",
  "password": "password123"
}
```

#### 登出
```http
POST /api/auth/logout
Authorization: Bearer <token>
```

#### 获取当前用户信息
```http
GET /api/auth/me
Authorization: Bearer <token>
```

### 歌单管理（需要认证）

#### 获取用户所有歌单
```http
GET /api/user/playlists
Authorization: Bearer <token>
```

#### 创建歌单
```http
POST /api/user/playlists/create
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "我的歌单",
  "description": "这是我的歌单描述"
}
```

#### 添加歌曲到歌单
```http
POST /api/user/playlists/add-song?playlist_id=<playlist_id>
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "歌曲名",
  "artist": "歌手名",
  "audio_url": "/path/to/audio",
  "audio_full_url": "/path/to/full/audio",
  "cover_url": "/path/to/cover",
  "lyric_url": "/path/to/lyric",
  "duration": 240
}
```

#### 从歌单移除歌曲
```http
DELETE /api/user/playlists/remove-song?playlist_id=<playlist_id>&title=<song_title>&artist=<artist_name>
Authorization: Bearer <token>
```

#### 删除歌单
```http
DELETE /api/user/playlists/delete?playlist_id=<playlist_id>
Authorization: Bearer <token>
```

### 音乐搜索

```http
GET /stream_pcm?song=<歌曲名>&singer=<歌手名>
```

## 数据存储

用户数据存储在以下文件中：

- `./files/users.json` - 用户账户信息
- `./files/user_playlists.json` - 用户歌单数据
- `./files/playlists.json` - 旧版全局歌单（向后兼容）

## 功能特性

### 🔒 安全性

- 密码使用 bcrypt 加密存储
- 基于 Token 的会话管理
- API 请求认证保护

### 📱 响应式设计

- 现代化 UI 界面
- 支持桌面和移动设备
- 流畅的用户体验

### 🎵 音乐功能

- 实时搜索音乐
- 在线播放
- 个人收藏管理
- 自定义歌单

### 🔄 向后兼容

- 保留旧版 API 端点
- 支持旧版界面访问
- 平滑升级路径

## 使用示例

### JavaScript 示例

```javascript
// 注册用户
const registerResponse = await fetch('/api/auth/register', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    username: 'myuser',
    email: 'user@example.com',
    password: 'securepass123'
  })
});
const { token, user } = await registerResponse.json();

// 获取歌单
const playlistsResponse = await fetch('/api/user/playlists', {
  headers: { 'Authorization': `Bearer ${token}` }
});
const playlists = await playlistsResponse.json();

// 搜索音乐
const musicResponse = await fetch('/stream_pcm?song=告白气球&singer=周杰伦');
const song = await musicResponse.json();

// 添加到歌单
await fetch(`/api/user/playlists/add-song?playlist_id=${playlists[0].id}`, {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify(song)
});
```

## 故障排除

### 问题：无法登录
- 检查用户名和密码是否正确
- 确认用户已注册
- 检查服务器日志

### 问题：Token 过期
- 重新登录获取新 Token
- Token 存储在 localStorage 中，清除浏览器缓存后需要重新登录

### 问题：添加歌曲失败
- 确认已登录并有有效 Token
- 检查歌单 ID 是否正确
- 确认歌曲信息完整

## 开发者信息

- 后端框架：Go (Golang)
- 前端框架：React 18 + TailwindCSS
- 认证方式：Token-based
- 数据存储：JSON 文件

## 更新日志

### v2.0.0 (当前版本)
- ✅ 添加用户注册和登录功能
- ✅ 实现个人歌单管理
- ✅ 创建现代化 Web 界面
- ✅ 支持在线音乐搜索和播放
- ✅ 向后兼容旧版 API

---

**享受您的音乐之旅！ 🎵**
