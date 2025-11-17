# 读回执功能 - 后续实施步骤

## 🎉 已完成工作总结

### Phase 1: Server 侧核心功能 ✅ **100% 完成**

我们已经成功实现了 Mattermost Server 侧的所有核心功能：

1. **数据库层** ✅
   - 创建了 `channel_read_cursors` 表
   - 添加了必要的索引
   - 迁移文件：`000147_create_channel_read_cursors.up.sql`

2. **Model 层** ✅
   - `ChannelReadCursor` - 读游标模型
   - `ReadCursorAdvanceRequest` - API 请求模型
   - `ReadCursorEvent` - 事件模型
   - WebSocket 事件类型定义
   - 完整的单元测试

3. **Store 层** ✅
   - `ChannelReadCursorStore` 接口
   - SQL 实现（支持 PostgreSQL）
   - Upsert 操作防止游标回退
   - 完整的 CRUD 方法

4. **App 层** ✅
   - `AdvanceChannelReadCursor()` - 推进读游标
   - `AutoAdvanceReadCursorOnChannelView()` - 自动追踪
   - `CleanupOldReadCursors()` - 数据清理
   - WebSocket 事件发布

5. **API 层** ✅
   - `POST /api/v4/channels/{id}/read_cursor` - 推进游标
   - `GET /api/v4/channels/{id}/read_cursor` - 获取游标
   - 集成到 `viewChannel` API
   - 自动追踪用户阅读行为

### Git 提交记录
```
5a0f25d docs: update MVP progress - Phase 1 completed ✅
c9dd073 feat: add API layer for read receipts (Phase 1.5)
8efc693 feat: add App layer for read receipts (Phase 1.4)
43a7683 feat: add Store layer for read receipts (Phase 1.3)
86249d1 feat: add Model layer for read receipts (Phase 1.2)
5844416 feat: add database migration for channel_read_cursors table
```

---

## 🚀 Phase 2: ReadIndexService（高性能读索引服务）

### 概述
ReadIndexService 是一个独立的 Go 微服务，负责：
- 消费 Mattermost Server 发送的读游标事件
- 维护内存中的高性能索引（使用 RoaringBitmap）
- 提供 HTTP API 查询"谁读了某条消息"

### 架构设计
```
Mattermost Server
    ↓ (发送事件)
Redis Stream
    ↓ (消费事件)
ReadIndexService (内存索引)
    ↓ (查询 API)
Mattermost Server → 前端
```

### 实施步骤

#### 2.1 项目初始化
```bash
cd read-index-service
go mod init github.com/mattermost/mattermost-read-index-service
go get github.com/RoaringBitmap/roaring
go get github.com/go-redis/redis/v8
go get github.com/gorilla/mux
```

#### 2.2 核心代码结构
```
read-index-service/
├── cmd/
│   └── server/
│       └── main.go              # 服务入口
├── internal/
│   ├── index/
│   │   ├── channel_state.go    # 频道状态管理
│   │   ├── segment.go           # 分段位图
│   │   └── index.go             # 核心索引逻辑
│   ├── consumer/
│   │   └── redis_consumer.go   # Redis Stream 消费者
│   └── api/
│       └── handlers.go          # HTTP API 处理器
├── go.mod
└── README.md
```

#### 2.3 关键实现文件

**参考实现**：`docs/READ_INDEX_SERVICE_IMPLEMENTATION.go` 已经包含了完整的实现代码，包括：
- `ChannelState` - 频道状态结构
- `ReadSegment` - 分段位图
- `HandleReadCursorEvent()` - 事件处理
- `GetReadersForSeq()` - 查询已读用户
- Redis Stream 消费逻辑
- HTTP API 端点

#### 2.4 部署配置

**Docker Compose 示例**：
```yaml
version: '3.8'
services:
  read-index-service:
    build: ./read-index-service
    ports:
      - "8066:8066"
    environment:
      - REDIS_URL=redis://redis:6379/0
      - PORT=8066
      - WINDOW_SIZE=1000
    depends_on:
      - redis
  
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
```

#### 2.5 与 Mattermost 集成

**修改 App 层的事件发布**（`server/channels/app/channel_read_cursor.go`）：
```go
func (a *App) publishReadCursorEvent(rctx request.CTX, event *model.ReadCursorEvent) error {
    // 当前是占位符，需要实现真正的 Redis Stream 发布
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

**添加查询 ReadIndexService 的方法**：
```go
func (a *App) GetPostReadReceipts(rctx request.CTX, postId string, limit int) (*ReadReceiptsResponse, *model.AppError) {
    // 调用 ReadIndexService HTTP API
    url := fmt.Sprintf("%s/channels/%s/posts/%d/readers?limit=%d", 
        a.Config().ReadIndexServiceURL, channelId, seq, limit)
    // ... HTTP 请求逻辑
}
```

---

## 🎨 Phase 3: 前端 UI 实现

### 3.1 Redux Actions
创建 `webapp/channels/src/actions/read_receipts.ts`：
```typescript
export function fetchReadCounts(channelId: string, postIds: string[]): ActionFunc {
    return async (dispatch) => {
        const counts = await Client4.getReadCounts(channelId, postIds);
        dispatch({
            type: ActionTypes.RECEIVED_READ_COUNTS,
            data: {channelId, counts},
        });
        return {data: counts};
    };
}
```

### 3.2 UI 组件
- `PostReadIndicator.tsx` - 显示已读计数
- `PostReadReceiptsModal.tsx` - 已读用户列表弹窗
- 在 `Post` 组件中集成

### 3.3 WebSocket 监听
监听 `read_cursor_advanced` 事件，实时更新 UI

---

## 📝 快速开始指南

### 选项 A：完整实现（推荐用于生产）
1. 实现 ReadIndexService（参考 `docs/READ_INDEX_SERVICE_IMPLEMENTATION.go`）
2. 配置 Redis Stream
3. 实现前端 UI
4. 进行性能测试

### 选项 B：简化版本（快速验证）
1. **跳过 ReadIndexService**，直接在 Server 侧实现简单查询
2. 在 App 层添加：
   ```go
   func (a *App) GetPostReadUsers(channelId string, postSeq int64) ([]string, error) {
       cursors, _ := a.Srv().Store().ChannelReadCursor().GetForChannel(channelId)
       var users []string
       for _, cursor := range cursors {
           if cursor.LastPostSeq >= postSeq {
               users = append(users, cursor.UserId)
           }
       }
       return users, nil
   }
   ```
3. 添加 API 端点
4. 实现前端 UI

**注意**：选项 B 适合小规模部署（<100 人频道），大规模场景必须使用选项 A。

---

## 🧪 测试计划

### 单元测试
```bash
cd server
go test ./channels/store/sqlstore -run TestChannelReadCursor
go test ./channels/app -run TestAdvanceChannelReadCursor
go test ./channels/api4 -run TestReadCursor
```

### 集成测试
1. 启动 Mattermost Server
2. 创建测试频道和用户
3. 发送消息
4. 调用 API 推进游标
5. 验证数据库记录

### 性能测试
- 10k 用户频道的写入性能
- 并发读取性能
- 内存占用监控

---

## 📚 相关文档

- `READ_RECEIPTS_PRODUCTION_MVP.md` - MVP 实施方案
- `READ_RECEIPTS_IMPLEMENTATION_PLAN.md` - 原始设计文档
- `docs/READ_INDEX_SERVICE_IMPLEMENTATION.go` - ReadIndexService 完整实现

---

## 💡 建议

**对于当前阶段**，我建议：

1. **先验证 Server 侧功能**
   - 运行 Mattermost Server
   - 使用 Postman 或 curl 测试 API
   - 验证数据库记录正确

2. **实现简化版查询**（选项 B）
   - 快速验证整个流程
   - 为前端开发提供可用的 API

3. **前端 UI 开发**
   - 实现基本的已读计数显示
   - 验证用户体验

4. **最后优化**
   - 根据实际使用情况决定是否需要 ReadIndexService
   - 小规模部署可能不需要

---

## 🎯 当前状态

```
✅ Phase 1: Server 侧核心功能 (100%)
⏳ Phase 2: ReadIndexService (0%)
⏳ Phase 3: 前端 UI (0%)
⏳ Phase 4: 优化与测试 (0%)
```

**总体进度**: 25% 完成

**下一步行动**: 选择实施路径（选项 A 或 B）并继续开发。
