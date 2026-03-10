# CANPulse 架构文档

## 概述

CANPulse 采用分层架构设计，将功能按职责清晰分离，确保代码的可维护性和可扩展性。

## 架构图

```
┌─────────────────────────────────────────────────────────┐
│                    main.go (应用层)                      │
│   - 命令行参数解析                                        │
│   - 协调各模块工作                                        │
│   - 信号处理                                             │
└─────────────────┬───────────────────────────────────────┘
                  │
    ┌─────────────┼─────────────┐
    │             │             │
    ▼             ▼             ▼
┌─────────┐  ┌──────────┐  ┌──────────┐
│  CAN    │  │ Serial   │  │ Monitor  │
│  协议层  │  │ 串口层   │  │ 监控层   │
└─────────┘  └──────────┘  └──────────┘
```

## 分层设计

### 第1层: 应用层 (`main.go`)

**职责**:
- 命令行接口
- 参数解析和验证
- 协调底层模块
- 信号处理（Ctrl+C）

**关键函数**:
- `main()`: 程序入口
- `run()`: 主监听循环
- `processBuffer()`: 缓冲区处理
- `setupSignalHandler()`: 信号处理

### 第2层: 功能模块 (`internal/`)

#### 2.1 CAN协议层 (`internal/can/`)

**职责**: CAN协议相关的所有逻辑

**文件结构**:
```
can/
├── frame.go      # CAN帧定义和解析
└── protocol.go   # 协议实现（SLCAN, OBD-II）
```

**核心类型**:

```go
// Frame - CAN帧结构
type Frame struct {
    Timestamp  time.Time
    ID         uint32
    DLC        uint8
    Data       []byte
    IsExtended bool
}

// Parser - CAN帧解析器
type Parser struct{}

// Bitrate - CAN波特率
type Bitrate string
```

**关键功能**:
- `Parse()`: 解析SLCAN格式的CAN帧
- `parseStandardFrame()`: 解析标准帧（11位ID）
- `parseExtendedFrame()`: 解析扩展帧（29位ID）
- `ParseOBDResponse()`: 解析OBD-II响应
- `SLCANCommand()`: 生成SLCAN初始化命令

#### 2.2 串口通信层 (`internal/serial/`)

**职责**: 封装串口设备的所有操作

**文件结构**:
```
serial/
└── device.go    # 串口设备管理
```

**核心类型**:

```go
// Device - 串口设备
type Device struct {
    port     serial.Port
    portName string
    baudRate int
}

// Config - 串口配置
type Config struct {
    PortName string
    BaudRate int
    Bitrate  can.Bitrate
}
```

**关键功能**:
- `ListPorts()`: 列出所有可用串口
- `AutoDetectPort()`: 自动检测USB转CAN设备
- `Open()`: 打开串口并初始化CAN接口
- `initCAN()`: 使用SLCAN协议初始化CAN设备
- `Read()`: 从串口读取数据

#### 2.3 监控分析层 (`internal/monitor/`)

**职责**: 数据监控、统计和记录

**文件结构**:
```
monitor/
├── monitor.go   # 监控器实现
└── stats.go     # 统计信息管理
```

**核心类型**:

```go
// Monitor - 监控器
type Monitor struct {
    stats      *Stats
    logger     *Logger
    options    Options
    frameCount int
}

// Stats - 统计信息管理器
type Stats struct {
    data     map[uint32]*IDStats
    mutex    sync.RWMutex
    minCount int
}

// IDStats - 单个CAN ID的统计
type IDStats struct {
    Count      int
    LastData   []byte
    LastUpdate time.Time
    FirstSeen  time.Time
}
```

**关键功能**:
- `ProcessFrame()`: 处理CAN帧
- `Update()`: 更新统计并检测变化
- `printFrame()`: 格式化输出CAN帧
- `Log()`: 记录到CSV文件
- `PrintStats()`: 打印统计信息

## 数据流

```
USB设备 → 串口数据
         ↓
    Serial.Read() - 读取原始字节
         ↓
    Buffer - 累积数据直到完整帧
         ↓
    Parser.Parse() - 解析SLCAN格式
         ↓
    Frame - CAN帧对象
         ↓
    Monitor.ProcessFrame() - 处理帧
         ↓
    ├─→ Stats.Update() - 更新统计和变化检测
    ├─→ Logger.Log() - 记录到CSV文件
    └─→ printFrame() - 终端显示
```

## 设计原则

### 1. 单一职责原则 (SRP)

每个模块只负责一个功能领域：
- `can`: 只处理CAN协议
- `serial`: 只处理串口通信
- `monitor`: 只处理监控和统计

### 2. 依赖倒置原则 (DIP)

高层模块不依赖低层模块的具体实现：
- `main.go` 通过接口使用各个模块
- 模块之间通过数据结构（`Frame`）通信

### 3. 开闭原则 (OCP)

对扩展开放，对修改封闭：
- 添加新的OBD-II PID只需修改 `protocol.go`
- 添加新的输出格式只需修改 `monitor.go`
- 支持新的设备类型只需修改 `device.go`

### 4. 接口隔离原则 (ISP)

没有大而全的接口，每个功能独立：
- `Parser` 专注于解析
- `Device` 专注于设备IO
- `Monitor` 专注于数据处理

## 错误处理

### 分层错误处理

```go
// 底层: 返回详细错误
func Open(config Config) (*Device, error) {
    if err != nil {
        return nil, fmt.Errorf("failed to open port: %w", err)
    }
}

// 中层: 包装错误添加上下文
func run() error {
    device, err := serial.Open(...)
    if err != nil {
        return fmt.Errorf("打开串口失败: %w", err)
    }
}

// 顶层: 处理错误（日志或退出）
func main() {
    if err := run(); err != nil {
        log.Fatal(err)
    }
}
```

## 并发设计

### 线程安全

1. **统计信息保护**:
```go
type Stats struct {
    data  map[uint32]*IDStats
    mutex sync.RWMutex  // 读写锁保护
}
```

2. **定时器协程**:
```go
func (m *Monitor) StartStatsTicker() {
    go func() {
        ticker := time.NewTicker(...)
        for range ticker.C {
            m.PrintStats()
        }
    }()
}
```

3. **信号处理协程**:
```go
func setupSignalHandler(mon *monitor.Monitor) {
    go func() {
        <-sigChan
        mon.PrintStats()
        os.Exit(0)
    }()
}
```

## 扩展点

### 1. 添加新的CAN协议支持

修改 `internal/can/frame.go`:
```go
func (p *Parser) Parse(data []byte) (*Frame, error) {
    switch dataStr[0] {
    case 't': // 标准帧
        return p.parseStandardFrame(dataStr)
    case 'T': // 扩展帧
        return p.parseExtendedFrame(dataStr)
    case 'x': // 你的新协议
        return p.parseCustomFrame(dataStr)
    }
}
```

### 2. 添加新的输出格式

修改 `internal/monitor/monitor.go`:
```go
func (m *Monitor) printFrame(frame *can.Frame, changeInfo ChangeInfo) {
    // 添加JSON输出
    if m.options.OutputJSON {
        m.printFrameJSON(frame)
    }
}
```

### 3. 添加新的设备类型

修改 `internal/serial/device.go`:
```go
func AutoDetectPort() (string, error) {
    patterns := []string{
        "cu.usbserial",
        "cu.SLAB_USBtoUART",
        "cu.wchusbserial",
        "cu.YourNewDevice",  // 添加新设备
    }
}
```

## 测试策略

### 单元测试建议

```bash
# 测试CAN帧解析
go test ./internal/can -v

# 测试统计功能
go test ./internal/monitor -v

# 测试串口设备检测
go test ./internal/serial -v
```

### 测试文件示例

```go
// internal/can/frame_test.go
func TestParse_StandardFrame(t *testing.T) {
    parser := NewParser()
    frame, err := parser.Parse([]byte("t1230811223344"))

    assert.NoError(t, err)
    assert.Equal(t, uint32(0x123), frame.ID)
    assert.Equal(t, uint8(8), frame.DLC)
}
```

## 性能考虑

### 1. 内存管理

- 使用缓冲区复用，减少内存分配
- 统计信息使用map，O(1)查找

### 2. IO优化

- 批量读取串口数据（256字节）
- 使用缓冲区累积完整帧

### 3. 并发优化

- 读写锁 (`sync.RWMutex`) 允许多读
- 统计定时器独立协程，不阻塞主循环

## 代码规范

### 命名约定

- **包名**: 小写单数名词 (`can`, `serial`, `monitor`)
- **类型**: 大驼峰 (`Frame`, `Device`, `Monitor`)
- **函数**: 大驼峰（导出）或小驼峰（内部）
- **常量**: 大驼峰 (`Bitrate500K`)

### 注释规范

```go
// Package can 提供CAN总线协议相关的功能
package can

// Frame 表示一个CAN帧
type Frame struct { ... }

// Parse 从原始字节数据解析CAN帧
// 支持SLCAN协议格式
func (p *Parser) Parse(data []byte) (*Frame, error) { ... }
```

### 文件组织

- 每个文件专注一个主题
- 相关功能放在同一个文件
- 文件名清晰表达用途

## 未来改进方向

### 短期
- [ ] 添加单元测试
- [ ] 支持更多OBD-II PID
- [ ] JSON输出格式
- [ ] 性能监控（帧率统计）

### 中期
- [ ] 支持CAN FD
- [ ] Web UI界面
- [ ] 数据回放功能
- [ ] 自定义过滤规则

### 长期
- [ ] 支持多CAN接口
- [ ] 实时图表显示
- [ ] 机器学习识别信号
- [ ] 云端数据分析

## 贡献指南

添加新功能时请遵循：
1. 保持模块职责单一
2. 添加充分的注释
3. 遵循Go代码规范
4. 编写单元测试
5. 更新文档

## 参考资料

- [SLCAN协议规范](http://www.can232.com/docs/canusb_manual.pdf)
- [OBD-II标准](https://en.wikipedia.org/wiki/OBD-II_PIDs)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go项目布局](https://github.com/golang-standards/project-layout)
