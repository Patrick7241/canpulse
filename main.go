// CANPulse - 汽车CAN总线监听工具
//
// 这是一个用Go语言编写的CAN总线监听工具，支持通过USB转CAN模块
// （如智潜物联、周立功等）读取和解析汽车CAN总线数据。
//
// 主要功能:
//   - 自动检测USB串口设备
//   - 支持SLCAN协议（兼容周立功等常见设备）
//   - 实时解析CAN帧数据（标准帧和扩展帧）
//   - 自动解析常见的OBD-II PID
//   - 变化检测模式 - 识别车身控制信号
//   - 数据记录和统计分析
//   - 发送CAN消息到总线
//
// 使用示例:
//
//	./canpulse -list                                    # 列出串口设备
//	./canpulse                                          # 自动检测并开始监听
//	./canpulse -port /dev/cu.usbserial-xxx              # 指定串口
//	./canpulse -detect -binary -log can_data.csv        # 变化检测+记录
//	./canpulse -send "123:11223344"                     # 发送CAN消息
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"canpulse/internal/can"
	"canpulse/internal/logger"
	"canpulse/internal/monitor"
	"canpulse/internal/serial"
)

// 命令行参数
var (
	portName      = flag.String("port", "", "串口设备名称 (例如: /dev/cu.usbserial-xxx)")
	baudRate      = flag.Int("baud", 115200, "串口波特率")
	canBitrate    = flag.String("canbitrate", "500K", "CAN总线波特率 (125K, 250K, 500K, 1M)")
	listPorts     = flag.Bool("list", false, "列出所有可用串口")
	logFile       = flag.String("log", "", "保存CAN数据到文件 (例如: can_log.csv)")
	detectMode    = flag.Bool("detect", false, "变化检测模式：只显示数据发生变化的帧")
	showBinary    = flag.Bool("binary", false, "显示二进制格式")
	minCount      = flag.Int("mincount", 1, "统计时过滤：只显示帧数>=此值的ID")
	statsInterval = flag.Int("stats", 0, "每N秒显示一次统计信息 (0=不显示)")
	sendFrame     = flag.String("send", "", "发送CAN帧 (格式: ID:DATA, 例如: 123:11223344)")
	logDir        = flag.String("logdir", "logs", "日志文件目录")
)

func main() {
	flag.Parse()

	// 初始化日志系统
	if err := logger.Init(logger.Config{
		Level:      logger.LevelInfo,
		LogDir:     *logDir,
		EnableFile: true,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "初始化日志系统失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.DefaultLogger.Close()

	// 列出所有串口
	if *listPorts {
		if err := listSerialPorts(); err != nil {
			logger.Error("列出串口失败: %v", err)
			os.Exit(1)
		}
		return
	}

	// 发送CAN帧
	if *sendFrame != "" {
		if err := sendCANFrame(*sendFrame); err != nil {
			logger.Error("发送CAN帧失败: %v", err)
			os.Exit(1)
		}
		return
	}

	// 运行主监听循环
	if err := run(); err != nil {
		logger.Error("运行失败: %v", err)
		os.Exit(1)
	}
}

// listSerialPorts 列出所有可用的串口设备
func listSerialPorts() error {
	ports, err := serial.ListPorts()
	if err != nil {
		return fmt.Errorf("获取串口列表失败: %w", err)
	}

	logger.Println("可用串口设备:")
	for _, port := range ports {
		logger.Printf("   %s\n", port)
	}

	return nil
}

// sendCANFrame 发送CAN帧到总线
func sendCANFrame(frameStr string) error {
	logger.Info("准备发送CAN帧: %s", frameStr)

	// 解析帧字符串: ID:DATA
	parts := strings.Split(frameStr, ":")
	if len(parts) != 2 {
		return fmt.Errorf("无效的帧格式，应为 ID:DATA (例如: 123:11223344)")
	}

	// 解析ID
	idStr := parts[0]
	id, err := strconv.ParseUint(idStr, 16, 32)
	if err != nil {
		return fmt.Errorf("无效的ID: %s", idStr)
	}

	// 解析数据
	dataStr := parts[1]
	data, err := hex.DecodeString(dataStr)
	if err != nil {
		return fmt.Errorf("无效的数据: %s", dataStr)
	}

	// 创建CAN帧
	frame, err := can.NewFrame(uint32(id), data, false)
	if err != nil {
		return fmt.Errorf("创建CAN帧失败: %w", err)
	}

	logger.Info("CAN帧: %s", frame.String())

	// 解析CAN波特率
	bitrate := can.Bitrate(*canBitrate)

	// 打开串口设备
	logger.Info("正在连接CAN设备...")
	device, err := serial.Open(serial.Config{
		PortName: *portName,
		BaudRate: *baudRate,
		Bitrate:  bitrate,
	})
	if err != nil {
		return fmt.Errorf("打开串口失败: %w", err)
	}
	defer device.Close()

	logger.Info("已连接: %s (波特率: %d, CAN: %s)", device.PortName(), device.BaudRate(), bitrate)

	// 发送帧
	logger.Info("发送中...")
	if err := device.Send(frame); err != nil {
		return fmt.Errorf("发送失败: %w", err)
	}

	logger.Info("发送成功！")
	return nil
}

// run 主监听循环
func run() error {
	// 解析CAN波特率
	bitrate := can.Bitrate(*canBitrate)

	// 打开串口设备
	logger.Info("正在连接CAN设备...")
	device, err := serial.Open(serial.Config{
		PortName: *portName,
		BaudRate: *baudRate,
		Bitrate:  bitrate,
	})
	if err != nil {
		return fmt.Errorf("打开串口失败: %w", err)
	}
	defer device.Close()

	logger.Info("已连接: %s (波特率: %d, CAN: %s)", device.PortName(), device.BaudRate(), bitrate)

	// 创建监控器
	mon, err := monitor.NewMonitor(monitor.Options{
		LogFile:       *logFile,
		DetectMode:    *detectMode,
		ShowBinary:    *showBinary,
		MinCount:      *minCount,
		StatsInterval: *statsInterval,
	})
	if err != nil {
		return fmt.Errorf("创建监控器失败: %w", err)
	}
	defer mon.Close()

	// 输出配置信息
	printConfig(*logFile, *detectMode, *showBinary)

	// 启动统计定时器
	mon.StartStatsTicker()

	// 设置信号处理（Ctrl+C时显示统计）
	setupSignalHandler(mon)

	// 开始监听
	logger.Println("开始监听CAN总线数据...")
	logger.Println("提示: 按 Ctrl+C 退出并显示统计信息")
	logger.Println("========================================")

	// 创建CAN帧解析器
	parser := can.NewParser()

	// 读取缓冲区
	buffer := make([]byte, 0, 1024)
	readBuf := make([]byte, 256)

	// 主循环：读取串口数据并解析CAN帧
	for {
		n, err := device.Read(readBuf)
		if err != nil {
			// 超时是正常的，继续读取
			continue
		}

		if n > 0 {
			buffer = append(buffer, readBuf[:n]...)
			buffer = processBuffer(buffer, parser, mon)
		}
	}
}

// processBuffer 处理缓冲区中的数据，提取并解析完整的CAN帧
func processBuffer(buffer []byte, parser *can.Parser, mon *monitor.Monitor) []byte {
	for {
		// 查找完整的CAN帧（以\r或\n结尾）
		endIdx := -1
		for i, b := range buffer {
			if b == '\r' || b == '\n' {
				endIdx = i
				break
			}
		}

		// 没有完整帧，返回
		if endIdx == -1 {
			break
		}

		// 提取一帧数据
		frameData := buffer[:endIdx]
		buffer = buffer[endIdx+1:]

		if len(frameData) == 0 {
			continue
		}

		// 解析CAN帧
		frame, err := parser.Parse(frameData)
		if err != nil {
			// 解析失败，打印原始数据（可能是设备响应）
			if len(frameData) > 0 && frameData[0] != '\r' && frameData[0] != '\n' {
				logger.Debug("原始数据: %s", string(frameData))
			}
			continue
		}

		// 处理帧
		mon.ProcessFrame(frame)
	}

	return buffer
}

// printConfig 打印配置信息
func printConfig(logFile string, detectMode, showBinary bool) {
	if logFile != "" {
		logger.Info("CAN数据将保存到: %s", logFile)
	}
	if detectMode {
		logger.Println("【变化检测模式】：只显示数据发生变化的帧")
	}
	if showBinary {
		logger.Println("【二进制模式】：显示数据的二进制表示")
	}
}

// setupSignalHandler 设置信号处理器（捕获Ctrl+C）
func setupSignalHandler(mon *monitor.Monitor) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("\n正在退出...")
		mon.PrintStats()
		logger.DefaultLogger.Close()
		os.Exit(0)
	}()
}
