// Package serial 提供串口设备管理和通信功能
package serial

import (
	"fmt"
	"strings"
	"time"

	"canpulse/internal/can"

	"go.bug.st/serial"
)

// Device 表示一个串口设备
type Device struct {
	port     serial.Port
	portName string
	baudRate int
}

// Config 串口设备配置
type Config struct {
	PortName string      // 串口名称，留空则自动检测
	BaudRate int         // 波特率
	Bitrate  can.Bitrate // CAN总线波特率
}

// ListPorts 列出所有可用的串口设备
func ListPorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, fmt.Errorf("failed to get port list: %w", err)
	}

	// 在macOS上只返回cu.*设备
	var cuPorts []string
	for _, port := range ports {
		if strings.Contains(port, "cu.") {
			cuPorts = append(cuPorts, port)
		}
	}

	return cuPorts, nil
}

// AutoDetectPort 自动检测可能的USB转CAN设备
func AutoDetectPort() (string, error) {
	ports, err := ListPorts()
	if err != nil {
		return "", err
	}

	// 查找可能的USB串口设备
	patterns := []string{"cu.usbserial", "cu.SLAB_USBtoUART", "cu.wchusbserial"}
	for _, port := range ports {
		for _, pattern := range patterns {
			if strings.Contains(port, pattern) {
				return port, nil
			}
		}
	}

	return "", fmt.Errorf("no USB serial device found")
}

// Open 打开串口设备并初始化CAN接口
func Open(config Config) (*Device, error) {
	// 自动检测串口（如果未指定）
	if config.PortName == "" {
		detectedPort, err := AutoDetectPort()
		if err != nil {
			return nil, fmt.Errorf("auto detect failed: %w", err)
		}
		config.PortName = detectedPort
	}

	// 打开串口
	mode := &serial.Mode{
		BaudRate: config.BaudRate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(config.PortName, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to open port %s: %w", config.PortName, err)
	}

	// 设置读取超时
	if err := port.SetReadTimeout(100 * time.Millisecond); err != nil {
		port.Close()
		return nil, fmt.Errorf("failed to set read timeout: %w", err)
	}

	device := &Device{
		port:     port,
		portName: config.PortName,
		baudRate: config.BaudRate,
	}

	// 初始化CAN设备
	if err := device.initCAN(config.Bitrate); err != nil {
		device.Close()
		return nil, fmt.Errorf("failed to initialize CAN: %w", err)
	}

	return device, nil
}

// initCAN 初始化CAN设备（SLCAN协议）
func (d *Device) initCAN(bitrate can.Bitrate) error {
	commands := []string{
		"C\r",                         // 关闭CAN通道
		bitrate.SLCANCommand() + "\r", // 设置波特率
		"O\r",                         // 打开CAN通道
	}

	for _, cmd := range commands {
		if _, err := d.port.Write([]byte(cmd)); err != nil {
			return fmt.Errorf("failed to send command '%s': %w", strings.TrimSpace(cmd), err)
		}
		time.Sleep(100 * time.Millisecond)

		// 读取并丢弃响应
		buf := make([]byte, 128)
		d.port.Read(buf)
	}

	return nil
}

// Read 从串口读取数据
func (d *Device) Read(buf []byte) (int, error) {
	return d.port.Read(buf)
}

// Send 发送CAN帧到总线
func (d *Device) Send(frame *can.Frame) error {
	if d.port == nil {
		return fmt.Errorf("device not opened")
	}

	// 将CAN帧转换为SLCAN格式
	data := frame.ToSLCAN()

	// 发送数据
	n, err := d.port.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write frame: %w", err)
	}

	if n != len(data) {
		return fmt.Errorf("incomplete write: wrote %d of %d bytes", n, len(data))
	}

	// 等待发送完成
	time.Sleep(10 * time.Millisecond)

	return nil
}

// Close 关闭串口设备
func (d *Device) Close() error {
	if d.port != nil {
		return d.port.Close()
	}
	return nil
}

// PortName 返回串口名称
func (d *Device) PortName() string {
	return d.portName
}

// BaudRate 返回串口波特率
func (d *Device) BaudRate() int {
	return d.baudRate
}
