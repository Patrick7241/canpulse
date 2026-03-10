# CANPulse

<div align="center">

**汽车CAN总线监听与分析工具**

[![Go Version](https://img.shields.io/badge/Go-1.25-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

一个用Go语言编写的专业CAN总线监听工具，支持通过USB转CAN模块实时读取和解析汽车CAN总线数据。

[功能特性](#功能特性) • [快速开始](#快速开始) • [使用指南](#使用指南) • [项目结构](#项目结构)

</div>

---

## 功能特性

### 核心功能
- 🔌 **自动设备检测** - 自动识别USB转CAN设备
- 📡 **SLCAN协议** - 兼容周立功、智潜物联等常见设备
- 📊 **实时解析** - 支持标准帧(11位)和扩展帧(29位)
- 🚗 **OBD-II解析** - 自动识别和解析常见的OBD-II PID

### 高级功能
- 🔍 **变化检测** - 只显示数据发生变化的帧，快速识别车身控制信号
- 🎯 **位级分析** - 二进制显示，查看每个bit的变化（转向灯、车窗等）
- 📝 **数据记录** - 保存所有CAN数据到CSV文件（含十六进制和二进制）
- 📈 **统计分析** - 实时统计每个CAN ID的出现频率
- 🎨 **高亮显示** - 变化的字节自动黄色高亮

## 快速开始

### 系统要求
- Go 1.25+
- macOS / Linux / Windows
- USB转CAN设备（智潜物联、周立功等）

### 安装

```bash
# 克隆仓库
git clone https://github.com/yourusername/canpulse.git
cd canpulse

# 安装依赖
go mod download

# 编译
go build -o canpulse
```

### 基础使用

```bash
# 1. 列出可用串口设备
./canpulse -list

# 2. 自动检测设备并开始监听
./canpulse

# 3. 指定串口设备
./canpulse -port /dev/cu.usbserial-14140
```

## 使用指南

### 基本监听

```bash
# 使用默认配置（500K波特率）
./canpulse

# 指定CAN波特率
./canpulse -canbitrate 250K
```

**支持的CAN波特率**: `10K`, `20K`, `50K`, `100K`, `125K`, `250K`, `500K`, `800K`, `1M`

### 识别车身控制信号

这是本工具的核心应用场景 - 找到转向灯、车窗、空调等私有CAN信号。

#### 方法1: 变化检测模式（推荐）

```bash
# 只显示数据变化的帧
./canpulse -detect -binary -log turn_signal.csv
```

**操作步骤**:
1. 启动程序，等待10秒建立基线
2. **只操作一个功能**（例如只打左转向灯）
3. 观察哪个ID开始变化，变化的字节会黄色高亮
4. 记录ID和字节位置
5. 重复验证

**输出示例**:
```
[123] 14:23:45.123 | 标准帧 | ID: 0x294 | DLC: 8 | Hex: [00 20 00 00 00 00 00 00] 🔄 [变化字节: [1]]
     二进制: [00000000 00100000 00000000 00000000 00000000 00000000 00000000 00000000]
                      ^^^^^^^^
                   第5位是1 - 可能是左转向灯
```

#### 方法2: 统计分析模式

```bash
# 每10秒显示一次统计，只显示出现≥100次的ID
./canpulse -stats 10 -mincount 100
```

**统计输出示例**:
```
========== CAN总线统计 ==========
总共监测到 45 个不同的CAN ID

ID       | 帧数  | 最后数据
---------|-------|--------------------------------
0x123    |  1523 | 00 11 22 33 44 55 66 77
0x294    |   892 | 00 20 00 00 00 00 00 00
...
```

### 数据记录与分析

```bash
# 记录所有数据到CSV文件
./canpulse -log my_car_data.csv

# 组合使用多个功能
./canpulse -detect -binary -log analysis.csv -stats 10
```

**CSV文件格式**:
```csv
Timestamp,ID,DLC,Data,Binary
2026-03-10 14:23:45.123,0x123,8,00 11 22 33 44 55 66 77,00000000 00010001 ...
```

### 命令行参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `-port` | string | 自动检测 | 串口设备名称 |
| `-baud` | int | 115200 | 串口波特率 |
| `-canbitrate` | string | 500K | CAN总线波特率 |
| `-list` | bool | false | 列出所有可用串口 |
| `-detect` | bool | false | 变化检测模式 |
| `-binary` | bool | false | 显示二进制格式 |
| `-log` | string | - | 保存数据到CSV文件 |
| `-stats` | int | 0 | 统计输出间隔（秒），0=不显示 |
| `-mincount` | int | 1 | 统计过滤：只显示帧数≥此值的ID |

## 项目结构

```
canpulse/
├── main.go                    # 主程序入口
├── go.mod                     # Go模块定义
├── go.sum                     # 依赖校验和
├── README.md                  # 项目文档
├── internal/                  # 内部包（不对外暴露）
│   ├── can/                   # CAN协议相关
│   │   ├── frame.go          # CAN帧定义和解析
│   │   └── protocol.go       # 协议实现（SLCAN, OBD-II）
│   ├── serial/               # 串口通信
│   │   └── device.go         # 串口设备管理
│   └── monitor/              # 监控和统计
│       ├── monitor.go        # 监控器实现
│       └── stats.go          # 统计信息管理
└── canpulse                  # 编译后的可执行文件
```

### 模块说明

#### `internal/can` - CAN协议层
负责CAN帧的定义、解析和协议实现。
- **frame.go**: 定义CAN帧结构，实现SLCAN协议解析
- **protocol.go**: OBD-II协议解析，CAN波特率定义

#### `internal/serial` - 串口通信层
封装串口设备的操作。
- **device.go**: 串口设备管理，自动检测，CAN初始化

#### `internal/monitor` - 监控分析层
提供数据监控、统计和记录功能。
- **monitor.go**: 监控器核心逻辑，数据记录
- **stats.go**: 统计信息管理，变化检测

#### `main.go` - 应用层
命令行接口，协调各模块工作。

## 应用场景

### 1. 汽车CAN总线分析
- 识别车身控制信号（转向灯、车窗、空调等）
- 逆向工程车辆CAN协议
- 开发自定义车载应用

### 2. OBD-II诊断
- 实时监控发动机参数
- 读取车辆传感器数据
- 故障诊断

### 3. CAN总线学习
- 理解CAN协议工作原理
- 学习汽车电子架构
- 教育和研究

## 常见车身信号CAN ID范围

以下是大众车系的参考范围（具体车型可能不同）：

| ID范围 | 功能 | 说明 |
|--------|------|------|
| 0x288-0x350 | 车身控制 | 转向灯、车门状态等 |
| 0x351-0x3D0 | 空调系统 | 温度、风速等 |
| 0x470-0x5A0 | 舒适系统 | 车窗、座椅、后视镜等 |
| 0x7E0-0x7EF | OBD-II | 标准诊断协议 |

> **注意**: 不同车型的CAN ID定义可能完全不同，需要通过变化检测来识别。

## 硬件连接

```
┌──────────┐   USB    ┌──────────────┐   CAN    ┌──────────┐
│  电脑    │ ────────> │ USB转CAN模块 │ ────────> │ OBD-II   │
│  (Mac)   │          │ (智潜/周立功) │          │ 接口     │
└──────────┘          └──────────────┘          └──────────┘
                                                      │
                                                      V
                                              ┌────────────┐
                                              │  汽车      │
                                              │  (CAN总线) │
                                              └────────────┘
```

**连接步骤**:
1. USB转CAN模块 → Mac电脑USB接口
2. CAN模块 → OBD-II转接线
3. OBD-II接口 → 汽车OBD-II母头（通常在方向盘下方）
4. 打开车辆点火开关或启动发动机

## 故障排查

### 问题: 设备未找到

```bash
# 检查设备是否连接
./canpulse -list

# 手动指定设备
./canpulse -port /dev/cu.usbserial-xxx
```

### 问题: 没有CAN数据

**可能原因**:
- 点火开关未打开
- CAN波特率不正确
- 物理连接问题

**解决方案**:
```bash
# 尝试不同的CAN波特率
./canpulse -canbitrate 250K  # 或 125K, 500K
```

### 问题: 数据太多看不清

**解决方案**:
```bash
# 使用变化检测模式
./canpulse -detect

# 过滤低频ID
./canpulse -mincount 100

# 保存到文件后用Excel分析
./canpulse -log data.csv
```

## 安全提示

- ⚠️ **只读操作**: 本工具只读取CAN数据，不会发送数据，不影响车辆运行
- ⚠️ **静止操作**: 请在车辆静止时操作，不要在行驶中使用
- ⚠️ **数据备份**: 如果记录敏感数据，请妥善保管日志文件
- ⚠️ **合法使用**: 仅用于自己的车辆，不要用于非法目的

## 开发

### 编译

```bash
# 标准编译
go build -o canpulse

# 跨平台编译
GOOS=linux GOARCH=amd64 go build -o canpulse-linux
GOOS=windows GOARCH=amd64 go build -o canpulse.exe
```

### 代码风格

项目遵循标准Go代码规范:
- 使用 `gofmt` 格式化代码
- 遵循 [Effective Go](https://golang.org/doc/effective_go) 指南
- 每个包都有清晰的职责分离

### 贡献

欢迎提交Issue和Pull Request！

## 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 致谢

- [go.bug.st/serial](https://github.com/bugst/go-serial) - 串口通信库
- SLCAN协议规范
- OBD-II标准协议

---

<div align="center">

**如果这个项目对你有帮助，请给个 ⭐ Star！**

Made with ❤️ by [Your Name]

</div>
