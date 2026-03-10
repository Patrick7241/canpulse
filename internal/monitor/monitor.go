// Package monitor 提供CAN总线监控和统计功能
package monitor

import (
	"fmt"
	"os"
	"time"

	"canpulse/internal/can"
	"canpulse/internal/logger"
)

// Options 监控器配置选项
type Options struct {
	LogFile       string // 日志文件路径
	DetectMode    bool   // 变化检测模式
	ShowBinary    bool   // 显示二进制格式
	MinCount      int    // 统计时的最小帧数
	StatsInterval int    // 统计输出间隔（秒）
}

// Monitor CAN总线监控器
type Monitor struct {
	stats      *Stats
	logger     *Logger
	options    Options
	frameCount int
}

// Logger 日志记录器
type Logger struct {
	file *os.File
}

// NewMonitor 创建监控器
func NewMonitor(opts Options) (*Monitor, error) {
	monitor := &Monitor{
		stats:   NewStats(opts.MinCount),
		options: opts,
	}

	// 创建日志记录器
	if opts.LogFile != "" {
		logger, err := NewLogger(opts.LogFile)
		if err != nil {
			return nil, err
		}
		monitor.logger = logger
	}

	return monitor, nil
}

// NewLogger 创建日志记录器
func NewLogger(filename string) (*Logger, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	// 写入CSV头
	fmt.Fprintf(file, "Timestamp,ID,DLC,Data,Binary\n")

	return &Logger{file: file}, nil
}

// ProcessFrame 处理一个CAN帧
func (m *Monitor) ProcessFrame(frame *can.Frame) {
	m.frameCount++

	// 更新统计并检测变化
	changeInfo := m.stats.Update(frame)

	// 记录到日志
	if m.logger != nil {
		m.logger.Log(frame)
	}

	// 显示帧（根据检测模式）
	if !m.options.DetectMode || changeInfo.Changed {
		m.printFrame(frame, changeInfo)
	}
}

// printFrame 打印CAN帧信息
func (m *Monitor) printFrame(frame *can.Frame, changeInfo ChangeInfo) {
	frameType := "标准帧"
	if frame.IsExtended {
		frameType = "扩展帧"
	}

	timestamp := frame.Timestamp.Format("15:04:05.000")

	// 格式化十六进制数据
	dataHex := ""
	for i, b := range frame.Data {
		if i > 0 {
			dataHex += " "
		}

		// 高亮变化的字节（黄色）
		if changeInfo.Changed && contains(changeInfo.ChangedBytes, i) {
			dataHex += fmt.Sprintf("\033[1;33m%02X\033[0m", b)
		} else {
			dataHex += fmt.Sprintf("%02X", b)
		}
	}

	// 变化标记
	changeMarker := ""
	if changeInfo.Changed && len(changeInfo.ChangedBytes) > 0 {
		changeMarker = fmt.Sprintf(" 🔄 [变化字节: %v]", changeInfo.ChangedBytes)
	}

	logger.Printf("[%d] %s | %s | ID: 0x%03X | DLC: %d | Hex: [%s]%s\n",
		m.frameCount, timestamp, frameType, frame.ID, frame.DLC, dataHex, changeMarker)

	// 显示二进制格式
	if m.options.ShowBinary {
		dataBin := ""
		for i, b := range frame.Data {
			if i > 0 {
				dataBin += " "
			}
			dataBin += fmt.Sprintf("%08b", b)
		}
		logger.Printf("     二进制: [%s]\n", dataBin)
	}

	// 解析OBD-II信息
	if info := can.ParseOBDResponse(frame); info != nil {
		logger.Printf("     └─ %s\n", can.FormatOBDInfo(info))
	}
}

// Log 记录CAN帧到日志文件
func (l *Logger) Log(frame *can.Frame) {
	if l.file == nil {
		return
	}

	timestamp := frame.Timestamp.Format("2006-01-02 15:04:05.000")

	// 格式化数据
	dataHex := ""
	dataBin := ""
	for i, b := range frame.Data {
		if i > 0 {
			dataHex += " "
			dataBin += " "
		}
		dataHex += fmt.Sprintf("%02X", b)
		dataBin += fmt.Sprintf("%08b", b)
	}

	fmt.Fprintf(l.file, "%s,0x%03X,%d,%s,%s\n",
		timestamp, frame.ID, frame.DLC, dataHex, dataBin)
}

// PrintStats 打印统计信息
func (m *Monitor) PrintStats() {
	m.stats.Print()
}

// StartStatsTicker 启动统计定时器
func (m *Monitor) StartStatsTicker() {
	if m.options.StatsInterval <= 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(time.Duration(m.options.StatsInterval) * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			m.PrintStats()
		}
	}()
}

// Close 关闭监控器
func (m *Monitor) Close() error {
	if m.logger != nil && m.logger.file != nil {
		return m.logger.file.Close()
	}
	return nil
}

// contains 检查切片中是否包含指定值
func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
