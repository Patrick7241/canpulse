// Package can 提供CAN总线协议相关的功能
package can

import (
	"encoding/hex"
	"fmt"
	"time"
)

// Frame 表示一个CAN帧
type Frame struct {
	Timestamp  time.Time // 接收时间戳
	ID         uint32    // CAN ID
	DLC        uint8     // 数据长度码 (Data Length Code)
	Data       []byte    // 数据内容
	IsExtended bool      // 是否为扩展帧
}

// Parser 负责解析CAN帧数据
type Parser struct{}

// NewParser 创建一个新的CAN帧解析器
func NewParser() *Parser {
	return &Parser{}
}

// Parse 从原始字节数据解析CAN帧
// 支持SLCAN协议格式:
//   - t<ID><DLC><DATA>\r - 标准帧 (11位ID)
//   - T<ID><DLC><DATA>\r - 扩展帧 (29位ID)
//   - r<ID><DLC>\r       - 远程帧
func (p *Parser) Parse(data []byte) (*Frame, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	dataStr := string(data)
	if len(dataStr) < 5 {
		return nil, fmt.Errorf("data too short: %d bytes", len(dataStr))
	}

	switch dataStr[0] {
	case 't': // 标准帧
		return p.parseStandardFrame(dataStr)
	case 'T': // 扩展帧
		return p.parseExtendedFrame(dataStr)
	case 'r', 'R': // 远程帧（暂不支持）
		return nil, fmt.Errorf("remote frame not supported")
	default:
		return nil, fmt.Errorf("unknown frame type: %c", dataStr[0])
	}
}

// parseStandardFrame 解析标准帧（11位ID）
// 格式: t<ID:3位十六进制><DLC:1位><DATA:DLC*2位十六进制>
func (p *Parser) parseStandardFrame(dataStr string) (*Frame, error) {
	if len(dataStr) < 5 {
		return nil, fmt.Errorf("invalid standard frame length")
	}

	frame := &Frame{
		Timestamp:  time.Now(),
		IsExtended: false,
	}

	// 解析ID (3个十六进制字符)
	idStr := dataStr[1:4]
	if _, err := fmt.Sscanf(idStr, "%x", &frame.ID); err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	// 解析DLC
	dlc := int(dataStr[4] - '0')
	if dlc < 0 || dlc > 8 {
		return nil, fmt.Errorf("invalid DLC: %d", dlc)
	}
	frame.DLC = uint8(dlc)

	// 解析数据
	if len(dataStr) >= 5+dlc*2 {
		dataHex := dataStr[5 : 5+dlc*2]
		data, err := hex.DecodeString(dataHex)
		if err != nil {
			return nil, fmt.Errorf("invalid data hex: %w", err)
		}
		frame.Data = data
	}

	return frame, nil
}

// parseExtendedFrame 解析扩展帧（29位ID）
// 格式: T<ID:8位十六进制><DLC:1位><DATA:DLC*2位十六进制>
func (p *Parser) parseExtendedFrame(dataStr string) (*Frame, error) {
	if len(dataStr) < 10 {
		return nil, fmt.Errorf("invalid extended frame length")
	}

	frame := &Frame{
		Timestamp:  time.Now(),
		IsExtended: true,
	}

	// 解析ID (8个十六进制字符)
	idStr := dataStr[1:9]
	if _, err := fmt.Sscanf(idStr, "%x", &frame.ID); err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	// 解析DLC
	dlc := int(dataStr[9] - '0')
	if dlc < 0 || dlc > 8 {
		return nil, fmt.Errorf("invalid DLC: %d", dlc)
	}
	frame.DLC = uint8(dlc)

	// 解析数据
	if len(dataStr) >= 10+dlc*2 {
		dataHex := dataStr[10 : 10+dlc*2]
		data, err := hex.DecodeString(dataHex)
		if err != nil {
			return nil, fmt.Errorf("invalid data hex: %w", err)
		}
		frame.Data = data
	}

	return frame, nil
}

// String 返回CAN帧的字符串表示
func (f *Frame) String() string {
	frameType := "标准帧"
	if f.IsExtended {
		frameType = "扩展帧"
	}

	dataHex := ""
	for i, b := range f.Data {
		if i > 0 {
			dataHex += " "
		}
		dataHex += fmt.Sprintf("%02X", b)
	}

	return fmt.Sprintf("%s | ID: 0x%03X | DLC: %d | Data: [%s]",
		frameType, f.ID, f.DLC, dataHex)
}

// IsOBDResponse 判断是否为OBD-II响应帧
// OBD-II响应ID范围: 0x7E8 - 0x7EF
func (f *Frame) IsOBDResponse() bool {
	return !f.IsExtended && f.ID >= 0x7E8 && f.ID <= 0x7EF && len(f.Data) > 2
}

// ToSLCAN 将CAN帧转换为SLCAN格式的字节流（用于发送）
// 标准帧格式: t<ID:3位十六进制><DLC:1位><DATA:DLC*2位十六进制>\r
// 扩展帧格式: T<ID:8位十六进制><DLC:1位><DATA:DLC*2位十六进制>\r
func (f *Frame) ToSLCAN() []byte {
	var result string

	if f.IsExtended {
		// 扩展帧: T<ID:8位><DLC:1位><DATA>
		result = fmt.Sprintf("T%08X%d", f.ID, f.DLC)
	} else {
		// 标准帧: t<ID:3位><DLC:1位><DATA>
		result = fmt.Sprintf("t%03X%d", f.ID, f.DLC)
	}

	// 添加数据
	for _, b := range f.Data {
		result += fmt.Sprintf("%02X", b)
	}

	// 添加结束符
	result += "\r"

	return []byte(result)
}

// NewFrame 创建一个新的CAN帧
func NewFrame(id uint32, data []byte, isExtended bool) (*Frame, error) {
	if len(data) > 8 {
		return nil, fmt.Errorf("data length exceeds 8 bytes")
	}

	return &Frame{
		Timestamp:  time.Now(),
		ID:         id,
		DLC:        uint8(len(data)),
		Data:       data,
		IsExtended: isExtended,
	}, nil
}
