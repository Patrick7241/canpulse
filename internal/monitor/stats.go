// Package monitor 提供CAN总线监控和统计功能
package monitor

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"canpulse/internal/can"
	"canpulse/internal/logger"
)

// IDStats 存储单个CAN ID的统计信息
type IDStats struct {
	Count      int       // 帧数量
	LastData   []byte    // 最后一次的数据
	LastUpdate time.Time // 最后更新时间
	FirstSeen  time.Time // 首次出现时间
}

// ChangeInfo 存储数据变化信息
type ChangeInfo struct {
	Changed      bool  // 是否发生变化
	ChangedBytes []int // 变化的字节索引
}

// Stats 统计信息管理器
type Stats struct {
	data     map[uint32]*IDStats
	mutex    sync.RWMutex
	minCount int // 统计时的最小帧数过滤
}

// NewStats 创建统计信息管理器
func NewStats(minCount int) *Stats {
	return &Stats{
		data:     make(map[uint32]*IDStats),
		minCount: minCount,
	}
}

// Update 更新统计信息并检测数据变化
func (s *Stats) Update(frame *can.Frame) ChangeInfo {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	stats, exists := s.data[frame.ID]
	if !exists {
		stats = &IDStats{
			FirstSeen: frame.Timestamp,
		}
		s.data[frame.ID] = stats
	}

	stats.Count++
	stats.LastUpdate = frame.Timestamp

	// 检测数据变化
	changeInfo := ChangeInfo{}

	if len(stats.LastData) > 0 && len(stats.LastData) == len(frame.Data) {
		for i := 0; i < len(frame.Data); i++ {
			if stats.LastData[i] != frame.Data[i] {
				changeInfo.Changed = true
				changeInfo.ChangedBytes = append(changeInfo.ChangedBytes, i)
			}
		}
	} else if len(stats.LastData) == 0 {
		// 首次出现算作变化
		changeInfo.Changed = true
	}

	// 保存当前数据
	stats.LastData = make([]byte, len(frame.Data))
	copy(stats.LastData, frame.Data)

	return changeInfo
}

// GetAll 获取所有统计信息（按ID排序）
func (s *Stats) GetAll() map[uint32]*IDStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make(map[uint32]*IDStats)
	for id, stats := range s.data {
		if stats.Count >= s.minCount {
			result[id] = stats
		}
	}
	return result
}

// Print 打印统计信息
func (s *Stats) Print() {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 按ID排序
	ids := make([]uint32, 0, len(s.data))
	for id := range s.data {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	logger.Println("\n========== CAN总线统计 ==========")
	logger.Printf("总共监测到 %d 个不同的CAN ID\n\n", len(s.data))
	logger.Println("ID       | 帧数  | 最后数据")
	logger.Println("---------|-------|--------------------------------")

	for _, id := range ids {
		stats := s.data[id]
		if stats.Count >= s.minCount {
			dataHex := ""
			for i, b := range stats.LastData {
				if i > 0 {
					dataHex += " "
				}
				dataHex += fmt.Sprintf("%02X", b)
			}
			logger.Printf("0x%03X   | %5d | %s\n", id, stats.Count, dataHex)
		}
	}
	logger.Println("=========================================\n")
}
