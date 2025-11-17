package index

import (
	"sync"

	"github.com/RoaringBitmap/roaring"
)

// Service 管理所有频道的读索引
type Service struct {
	channels   map[string]*ChannelState
	mu         sync.RWMutex
	windowSize int64
}

// ChannelState 表示一个频道的读索引状态
type ChannelState struct {
	ChannelID   string
	MaxSeq      int64
	UserCursors map[string]int64  // user_id -> last_seq
	UserIndex   map[string]uint32 // user_id -> bitmap index
	IndexToUser []string          // index -> user_id (反向映射)
	Segments    []*ReadSegment
	WindowSize  int64
	mu          sync.RWMutex
}

// ReadSegment 表示一个消息段的读索引
type ReadSegment struct {
	StartSeq int64
	EndSeq   int64
	Readers  *roaring.Bitmap // 读过此段的用户位图
}

// ReadCursorEvent 读游标事件
type ReadCursorEvent struct {
	Type        string `json:"type"`
	EventID     string `json:"event_id"`
	ChannelID   string `json:"channel_id"`
	UserID      string `json:"user_id"`
	PrevLastSeq int64  `json:"prev_last_seq"`
	NewLastSeq  int64  `json:"new_last_seq"`
	Timestamp   int64  `json:"timestamp"`
}

// NewService 创建新的索引服务
func NewService(windowSize int64) *Service {
	return &Service{
		channels:   make(map[string]*ChannelState),
		windowSize: windowSize,
	}
}

// HandleEvent 处理读游标事件
func (s *Service) HandleEvent(event *ReadCursorEvent) error {
	s.mu.RLock()
	cs, exists := s.channels[event.ChannelID]
	s.mu.RUnlock()

	if !exists {
		cs = s.createChannelState(event.ChannelID)
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	// 确保用户有位图索引
	userIdx, exists := cs.UserIndex[event.UserID]
	if !exists {
		userIdx = uint32(len(cs.IndexToUser))
		cs.UserIndex[event.UserID] = userIdx
		cs.IndexToUser = append(cs.IndexToUser, event.UserID)
	}

	// 获取旧游标
	oldSeq := cs.UserCursors[event.UserID]
	if event.NewLastSeq <= oldSeq {
		return nil // 序号没有增加
	}

	// 更新游标
	cs.UserCursors[event.UserID] = event.NewLastSeq

	// 更新段位图
	for _, seg := range cs.Segments {
		if seg.StartSeq > event.NewLastSeq {
			break
		}
		if seg.EndSeq > oldSeq {
			seg.Readers.Add(userIdx)
		}
	}

	// 更新最大序号
	if event.NewLastSeq > cs.MaxSeq {
		cs.MaxSeq = event.NewLastSeq
		cs.ensureSegmentsCover(event.NewLastSeq)
	}

	// 清理旧段
	cs.pruneOldSegments()

	return nil
}

// GetReadersForSeq 获取读过某条消息的用户列表
func (s *Service) GetReadersForSeq(channelID string, seq int64, limit int) ([]string, int, error) {
	s.mu.RLock()
	cs, exists := s.channels[channelID]
	s.mu.RUnlock()

	if !exists {
		return []string{}, 0, nil
	}

	cs.mu.RLock()
	defer cs.mu.RUnlock()

	// 合并所有 EndSeq >= seq 的段
	merged := roaring.New()
	for _, seg := range cs.Segments {
		if seg.EndSeq >= seq {
			merged.Or(seg.Readers)
		}
	}

	count := int(merged.GetCardinality())
	readers := make([]string, 0, min(count, limit))

	iter := merged.Iterator()
	for iter.HasNext() && len(readers) < limit {
		userIdx := iter.Next()
		if int(userIdx) < len(cs.IndexToUser) {
			readers = append(readers, cs.IndexToUser[userIdx])
		}
	}

	return readers, count, nil
}

// GetReadCounts 批量获取已读计数
func (s *Service) GetReadCounts(channelID string, seqs []int64) map[int64]int {
	s.mu.RLock()
	cs, exists := s.channels[channelID]
	s.mu.RUnlock()

	result := make(map[int64]int)
	if !exists {
		for _, seq := range seqs {
			result[seq] = 0
		}
		return result
	}

	cs.mu.RLock()
	defer cs.mu.RUnlock()

	for _, seq := range seqs {
		merged := roaring.New()
		for _, seg := range cs.Segments {
			if seg.EndSeq >= seq {
				merged.Or(seg.Readers)
			}
		}
		result[seq] = int(merged.GetCardinality())
	}

	return result
}

// GetStats 获取服务统计信息
func (s *Service) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	channels := make([]map[string]interface{}, 0)
	for _, cs := range s.channels {
		cs.mu.RLock()
		channelStats := map[string]interface{}{
			"channel_id":  cs.ChannelID,
			"max_seq":     cs.MaxSeq,
			"users_count": len(cs.UserCursors),
			"segments":    len(cs.Segments),
		}
		cs.mu.RUnlock()
		channels = append(channels, channelStats)
	}

	return map[string]interface{}{
		"channels_count": len(s.channels),
		"channels":       channels,
	}
}

// 内部方法

func (s *Service) createChannelState(channelID string) *ChannelState {
	s.mu.Lock()
	defer s.mu.Unlock()

	cs := &ChannelState{
		ChannelID:   channelID,
		MaxSeq:      0,
		UserCursors: make(map[string]int64),
		UserIndex:   make(map[string]uint32),
		IndexToUser: make([]string, 0),
		Segments:    make([]*ReadSegment, 0),
		WindowSize:  s.windowSize,
	}

	s.channels[channelID] = cs
	return cs
}

func (cs *ChannelState) ensureSegmentsCover(maxSeq int64) {
	segmentSize := int64(100) // 每段 100 条消息

	if len(cs.Segments) == 0 {
		cs.Segments = append(cs.Segments, &ReadSegment{
			StartSeq: 0,
			EndSeq:   segmentSize - 1,
			Readers:  roaring.New(),
		})
	}

	lastSeg := cs.Segments[len(cs.Segments)-1]
	for lastSeg.EndSeq < maxSeq {
		newSeg := &ReadSegment{
			StartSeq: lastSeg.EndSeq + 1,
			EndSeq:   lastSeg.EndSeq + segmentSize,
			Readers:  roaring.New(),
		}
		cs.Segments = append(cs.Segments, newSeg)
		lastSeg = newSeg
	}
}

func (cs *ChannelState) pruneOldSegments() {
	threshold := cs.MaxSeq - cs.WindowSize
	if threshold <= 0 {
		return
	}

	newSegments := make([]*ReadSegment, 0)
	for _, seg := range cs.Segments {
		if seg.EndSeq >= threshold {
			newSegments = append(newSegments, seg)
		}
	}
	cs.Segments = newSegments
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
