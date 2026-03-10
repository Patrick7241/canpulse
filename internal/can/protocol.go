// Package can 提供CAN总线协议相关的功能
package can

import "fmt"

// Bitrate 表示CAN总线波特率
type Bitrate string

const (
	Bitrate10K  Bitrate = "10K"
	Bitrate20K  Bitrate = "20K"
	Bitrate50K  Bitrate = "50K"
	Bitrate100K Bitrate = "100K"
	Bitrate125K Bitrate = "125K"
	Bitrate250K Bitrate = "250K"
	Bitrate500K Bitrate = "500K"
	Bitrate800K Bitrate = "800K"
	Bitrate1M   Bitrate = "1M"
)

// SLCANCommand 返回SLCAN协议的波特率设置命令
func (b Bitrate) SLCANCommand() string {
	switch b {
	case Bitrate10K:
		return "S0"
	case Bitrate20K:
		return "S1"
	case Bitrate50K:
		return "S2"
	case Bitrate100K:
		return "S3"
	case Bitrate125K:
		return "S4"
	case Bitrate250K:
		return "S5"
	case Bitrate500K:
		return "S6"
	case Bitrate800K:
		return "S7"
	case Bitrate1M:
		return "S8"
	default:
		return "S6" // 默认500K
	}
}

// OBDPIDInfo 存储OBD-II PID的解析信息
type OBDPIDInfo struct {
	Mode        uint8  // OBD模式
	PID         uint8  // PID编号
	Name        string // 参数名称
	Description string // 参数描述
	Unit        string // 单位
}

// ParseOBDResponse 解析OBD-II响应帧
func ParseOBDResponse(frame *Frame) *OBDPIDInfo {
	if !frame.IsOBDResponse() {
		return nil
	}

	info := &OBDPIDInfo{
		Mode: frame.Data[1],
		PID:  frame.Data[2],
	}

	// 只处理Mode 01的响应（Mode 01 + 0x40 = 0x41）
	if info.Mode != 0x41 {
		return info
	}

	// 根据PID解析数据
	switch info.PID {
	case 0x0C: // 发动机转速
		if len(frame.Data) >= 5 {
			rpm := (uint16(frame.Data[3])<<8 | uint16(frame.Data[4])) / 4
			info.Name = "发动机转速"
			info.Description = fmt.Sprintf("%d", rpm)
			info.Unit = "RPM"
		}
	case 0x0D: // 车速
		if len(frame.Data) >= 4 {
			speed := frame.Data[3]
			info.Name = "车速"
			info.Description = fmt.Sprintf("%d", speed)
			info.Unit = "km/h"
		}
	case 0x05: // 冷却液温度
		if len(frame.Data) >= 4 {
			temp := int(frame.Data[3]) - 40
			info.Name = "冷却液温度"
			info.Description = fmt.Sprintf("%d", temp)
			info.Unit = "°C"
		}
	case 0x0F: // 进气温度
		if len(frame.Data) >= 4 {
			temp := int(frame.Data[3]) - 40
			info.Name = "进气温度"
			info.Description = fmt.Sprintf("%d", temp)
			info.Unit = "°C"
		}
	case 0x11: // 节气门位置
		if len(frame.Data) >= 4 {
			position := float64(frame.Data[3]) * 100.0 / 255.0
			info.Name = "节气门位置"
			info.Description = fmt.Sprintf("%.1f", position)
			info.Unit = "%"
		}
	case 0x2F: // 燃油液位
		if len(frame.Data) >= 4 {
			level := float64(frame.Data[3]) * 100.0 / 255.0
			info.Name = "燃油液位"
			info.Description = fmt.Sprintf("%.1f", level)
			info.Unit = "%"
		}
	default:
		info.Name = "未知参数"
		info.Description = fmt.Sprintf("PID 0x%02X", info.PID)
	}

	return info
}

// FormatOBDInfo 格式化OBD信息为字符串
func FormatOBDInfo(info *OBDPIDInfo) string {
	if info.Name == "" {
		return fmt.Sprintf("OBD-II 响应: Mode=0x%02X, PID=0x%02X", info.Mode, info.PID)
	}
	return fmt.Sprintf("OBD-II 响应: Mode=0x%02X, PID=0x%02X | %s: %s %s",
		info.Mode, info.PID, info.Name, info.Description, info.Unit)
}
