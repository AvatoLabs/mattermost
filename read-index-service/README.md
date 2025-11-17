# Mattermost Read Index Service

高性能读回执索引服务，使用 RoaringBitmap 提供快速的"谁读了这条消息"查询。

## 功能特性

- ✅ **高性能内存索引**：使用 RoaringBitmap 压缩存储
- ✅ **分段存储**：每 100 条消息一个段，优化查询性能
- ✅ **滑动窗口**：自动清理旧数据，控制内存使用
- ✅ **Redis Stream 集成**：实时消费读游标事件
- ✅ **HTTP API**：提供查询接口

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 配置环境变量

**重要**：ReadIndexService 使用 Mattermost 现有的 Redis 实例，不需要单独部署 Redis。

```bash
# 使用 Mattermost 的 Redis（通常在 6379 端口）
export REDIS_URL="redis://localhost:6379/0"
export PORT="8066"
export WINDOW_SIZE="1000"
```

### 3. 运行服务

```bash
go run cmd/server/main.go
```

### 4. 使用 Docker

```bash
docker build -t read-index-service .
docker run -p 8066:8066 \
  -e REDIS_URL=redis://redis:6379/0 \
  read-index-service
```

## API 端点

### 健康检查
```bash
GET /health
```

### 获取已读用户列表
```bash
GET /channels/{channel_id}/posts/{seq}/readers?limit=50
```

响应：
```json
{
  "count": 123,
  "readers": ["user1", "user2", ...],
  "truncated": false
}
```

### 批量获取已读计数
```bash
POST /read-counts
Content-Type: application/json

{
  "channel_id": "channel123",
  "seqs": [1234567890000, 1234567895000]
}
```

响应：
```json
{
  "1234567890000": 45,
  "1234567895000": 32
}
```

### 服务统计
```bash
GET /stats
```

## 架构设计

### 数据结构

```
ChannelState
├── UserCursors: map[user_id]last_seq  // 每个用户的读游标
├── UserIndex: map[user_id]bitmap_idx  // 用户到位图索引的映射
├── IndexToUser: []user_id             // 反向映射
└── Segments: []ReadSegment            // 分段位图
    ├── StartSeq: 0
    ├── EndSeq: 99
    └── Readers: RoaringBitmap         // 读过此段的用户位图
```

### 性能特性

- **写入性能**：O(segments) ≈ O(10)，每个事件只需更新少量段
- **查询性能**：O(segments) ≈ O(10)，合并相关段的位图
- **内存占用**：10k 用户 × 1000 条窗口 ≈ 20KB/频道

### 滑动窗口

- 默认保留最近 1000 条消息的索引
- 自动清理超出窗口的旧段
- 可通过 `WINDOW_SIZE` 环境变量配置

## 与 Mattermost 集成

### 1. Mattermost Server 发送事件

修改 `server/channels/app/channel_read_cursor.go`：

```go
func (a *App) publishReadCursorEvent(rctx request.CTX, event *model.ReadCursorEvent) error {
    if a.Srv().RedisClient() != nil {
        data, _ := json.Marshal(event)
        return a.Srv().RedisClient().XAdd(rctx.Context(), &redis.XAddArgs{
            Stream: "read_cursor_events",
            Values: map[string]interface{}{"data": data},
        }).Err()
    }
    return nil
}
```

### 2. Mattermost Server 查询索引

添加新的 App 方法：

```go
func (a *App) GetPostReadReceipts(channelId string, seq int64, limit int) (*ReadReceiptsResponse, error) {
    url := fmt.Sprintf("%s/channels/%s/posts/%d/readers?limit=%d", 
        a.Config().ReadIndexServiceURL, channelId, seq, limit)
    
    resp, err := http.Get(url)
    // ... 处理响应
}
```

## 监控

### Prometheus 指标（TODO）

- `read_index_channels_total` - 索引的频道数
- `read_index_users_total` - 索引的用户数
- `read_index_events_processed_total` - 处理的事件数
- `read_index_memory_bytes` - 内存使用量

## 开发

### 运行测试

```bash
go test ./...
```

### 构建

```bash
go build -o bin/read-index-service cmd/server/main.go
```

## 许可证

与 Mattermost 主项目相同
