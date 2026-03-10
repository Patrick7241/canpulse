# CANPulse

汽车CAN总线监听与分析工具。支持通过USB转CAN模块（智潜物联、周立功等）实时读取、解析和发送CAN总线数据。

## 功能特性

- **实时监听**: 支持标准帧(11位)和扩展帧(29位)
- **变化检测**: 高亮显示数据变化的字节，快速识别车身控制信号
- **二进制显示**: 查看每个bit的变化，分析位级信号
- **发送CAN消息**: 向总线发送自定义CAN帧
- **数据记录**: 保存CAN数据到CSV文件
- **统计分析**: 实时统计每个CAN ID的出现频率
- **OBD-II解析**: 自动识别并解析常见OBD-II参数（转速、车速、温度等）
- **日志系统**: 所有操作同时记录到控制台和按日期分类的日志文件

## 系统要求

- Go 1.25+
- macOS / Linux / Windows
- USB转CAN设备（智潜物联、周立功等兼容SLCAN协议的设备）

## 安装

```bash
# 克隆仓库
git clone https://github.com/Patrick7241/canpulse.git
cd canpulse

# 安装依赖
go mod download

# 编译
go build -o canpulse
```

## 快速开始

### 1. 硬件连接

```
电脑(Mac) --USB--> USB转CAN模块 --CAN--> OBD-II接口 ---> 汽车OBD-II口
```

连接步骤：
1. USB转CAN模块插入电脑USB口
2. CAN模块通过OBD-II转接线连接到汽车OBD-II接口（通常在方向盘下方）
3. 打开车辆点火开关或启动发动机

### 2. 列出可用串口

```bash
./canpulse -list
```

输出示例：
```
可用串口设备:
   /dev/cu.usbserial-14140
   /dev/cu.SLAB_USBtoUART
```

### 3. 开始监听

```bash
# 自动检测设备
./canpulse

# 指定串口设备
./canpulse -port /dev/cu.usbserial-14140

# 指定CAN波特率（常见：125K, 250K, 500K, 1M）
./canpulse -canbitrate 500K
```

## 使用说明

### 基本监听

最简单的使用方式：

```bash
./canpulse
```

输出示例：
```
[1] 14:23:45.123 | 标准帧 | ID: 0x123 | DLC: 8 | Hex: [00 11 22 33 44 55 66 77]
[2] 14:23:45.234 | 标准帧 | ID: 0x7E8 | DLC: 8 | Hex: [03 41 0C 1F A0 00 00 00]
     └─ OBD-II 响应: Mode=0x41, PID=0x0C | 发动机转速: 2024 RPM
```

### 变化检测模式（推荐用于信号识别）

**用途**：找到转向灯、车窗、空调等私有CAN信号

```bash
./canpulse -detect -binary
```

**操作步骤**：
1. 启动变化检测模式
2. 等待10秒，让程序建立基线
3. **只操作一个功能**（例如只打开左转向灯）
4. 观察哪个ID开始变化，变化的字节会用黄色高亮显示
5. 记录ID和字节位置
6. 关闭功能，再打开，重复验证

**输出示例**：
```
[123] 14:23:45.123 | 标准帧 | ID: 0x294 | DLC: 8 | Hex: [00 20 00 00 00 00 00 00] 🔄 [变化字节: [1]]
     二进制: [00000000 00100000 00000000 00000000 00000000 00000000 00000000 00000000]
                      ^^^^^^^^
                   第5位是1，可能是左转向灯信号
```

### 发送CAN消息

**格式**: `ID:DATA`（十六进制）

```bash
# 发送标准帧，ID=0x123，数据=11223344
./canpulse -send "123:11223344"

# 发送到指定设备
./canpulse -port /dev/cu.usbserial-14140 -send "456:AABBCCDD"

# 发送到不同波特率的总线
./canpulse -canbitrate 250K -send "789:12345678"
```

**应用场景**：
- 验证发现的控制信号
- 模拟CAN消息进行测试
- 开发车载应用

**⚠️ 安全警告**：
- CAN消息会直接发送到车辆总线，可能影响车辆行为
- 建议在车辆静止时进行发送测试
- 发送前确保了解消息的含义

### 数据记录

将CAN数据保存到CSV文件：

```bash
# 保存到CSV
./canpulse -log can_data.csv

# 组合使用多个功能
./canpulse -detect -binary -log data.csv -stats 10
```

CSV文件格式：
```csv
Timestamp,ID,DLC,Data,Binary
2026-03-10 14:23:45.123,0x123,8,00 11 22 33 44 55 66 77,00000000 00010001 ...
```

### 统计分析

定期显示CAN ID统计信息：

```bash
# 每10秒显示一次统计
./canpulse -stats 10

# 只显示出现次数>=100的ID
./canpulse -stats 10 -mincount 100
```

输出示例：
```
========== CAN总线统计 ==========
总共监测到 45 个不同的CAN ID

ID       | 帧数  | 最后数据
---------|-------|--------------------------------
0x123    |  1523 | 00 11 22 33 44 55 66 77
0x294    |   892 | 00 20 00 00 00 00 00 00
0x7E8    |   456 | 03 41 0C 1F A0 00 00 00
=========================================
```

### 日志系统

所有操作会自动记录到日志文件：

```bash
# 默认日志目录: logs/
./canpulse

# 指定日志目录
./canpulse -logdir custom_logs
```

日志文件按日期自动分类：
```
logs/
├── canpulse-2026-03-10.log
├── canpulse-2026-03-11.log
└── canpulse-2026-03-12.log
```

查看日志：
```bash
# 查看今天的日志
cat logs/canpulse-$(date +%Y-%m-%d).log

# 搜索特定ID
grep "ID: 0x123" logs/*.log

# 查看错误日志
grep "\[ERROR\]" logs/*.log
```

## 命令行参数

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `-port` | string | 自动检测 | 串口设备名称 (例如: /dev/cu.usbserial-xxx) |
| `-baud` | int | 115200 | 串口波特率 |
| `-canbitrate` | string | 500K | CAN总线波特率 (10K/20K/50K/100K/125K/250K/500K/800K/1M) |
| `-list` | bool | false | 列出所有可用串口 |
| `-send` | string | - | 发送CAN帧 (格式: ID:DATA，例如: 123:11223344) |
| `-detect` | bool | false | 变化检测模式：只显示数据发生变化的帧 |
| `-binary` | bool | false | 显示二进制格式 |
| `-log` | string | - | 保存CAN数据到CSV文件 |
| `-logdir` | string | logs | 日志文件目录 |
| `-stats` | int | 0 | 每N秒显示一次统计信息 (0=不显示) |
| `-mincount` | int | 1 | 统计时过滤：只显示帧数>=此值的ID |

## 实战案例

### 案例1：识别转向灯信号

```bash
# 1. 启动变化检测模式
./canpulse -detect -binary -log turn_signal.csv

# 2. 等待10秒建立基线

# 3. 打开左转向灯
#    观察输出，记录变化的ID（假设是0x294）和字节位置（字节1的第5位）

# 4. 关闭转向灯，再打开右转向灯
#    对比数据，确认哪一位对应左/右转向

# 5. 验证发现的信号
./canpulse -send "294:00200000"  # 应该会触发左转向灯
```

### 案例2：监控车辆状态

```bash
# 持续监听并记录，每30秒显示统计
./canpulse -stats 30 -log vehicle_data.csv

# 查看实时的OBD-II数据（转速、车速等）
# 所有数据会保存到CSV文件和日志文件
```

### 案例3：批量发送测试

创建测试脚本：

```bash
#!/bin/bash
# test_sequence.sh

echo "开始测试序列..."

for i in {0..255}; do
  hex=$(printf "%02X" $i)
  echo "发送: 123:${hex}000000"
  ./canpulse -send "123:${hex}000000"
  sleep 0.5
done

echo "测试完成"
```

运行：
```bash
chmod +x test_sequence.sh
./test_sequence.sh
```

### 案例4：多CAN波特率扫描

如果不确定车辆的CAN波特率：

```bash
# 依次尝试常见波特率
for rate in 125K 250K 500K 1M; do
  echo "尝试 $rate..."
  timeout 5 ./canpulse -canbitrate $rate
done
```

## 常见车身信号CAN ID参考

以下是大众车系的常见ID范围（仅供参考，不同车型可能不同）：

| ID范围 | 可能的功能 | 说明 |
|--------|-----------|------|
| 0x288-0x350 | 车身控制 | 转向灯、车门状态、后视镜等 |
| 0x351-0x3D0 | 空调系统 | 温度、风速、AC开关等 |
| 0x470-0x5A0 | 舒适系统 | 车窗、座椅、灯光等 |
| 0x7E0-0x7EF | OBD-II | 标准诊断协议（通用） |

**注意**: 具体ID需要通过变化检测来识别。

## 故障排查

### 问题1: 找不到USB设备

**症状**: 运行 `-list` 没有看到 usbserial 设备

**解决方案**:
1. 确认USB转CAN模块已插入电脑
2. 检查设备是否需要安装驱动（CH340/CP2102等）
3. macOS可能需要授予串口访问权限
4. 重新插拔USB设备

### 问题2: 没有CAN数据

**症状**: 程序运行但没有输出CAN帧

**可能原因**:
- 车辆点火开关未打开
- CAN波特率不正确
- 物理连接问题

**解决方案**:
```bash
# 尝试不同的CAN波特率
./canpulse -canbitrate 125K
./canpulse -canbitrate 250K
./canpulse -canbitrate 500K
./canpulse -canbitrate 1M

# 确认点火开关已打开或发动机运行
# 检查OBD-II接口连接是否牢固
```

### 问题3: 数据太多看不清

**解决方案**:
```bash
# 使用变化检测模式，只看变化的数据
./canpulse -detect

# 过滤低频ID
./canpulse -mincount 100

# 保存到文件后用Excel分析
./canpulse -log data.csv
```

### 问题4: 发送失败

**症状**: 发送命令报错或无响应

**检查项**:
1. 确认格式正确：`ID:DATA` （十六进制，不要有0x前缀）
2. 数据长度不超过8字节（16个十六进制字符）
3. 确认设备支持发送功能
4. 检查串口是否被其他程序占用

## 注意事项

### 安全警告

- ⚠️ **只读为主**: 建议优先使用监听功能，谨慎使用发送功能
- ⚠️ **了解风险**: CAN消息可能影响车辆行为，仅在了解风险后使用
- ⚠️ **静止测试**: 发送测试请在车辆静止、熄火状态下进行
- ⚠️ **合法使用**: 仅用于自己的车辆，不要用于非法目的

### 数据隐私

- 日志文件可能包含车辆识别信息和行驶数据
- 请妥善保管日志文件，不要随意分享
- `logs/` 目录已自动添加到 `.gitignore`

### 存储管理

- 日志文件会随时间累积，建议定期清理
- CSV文件可能很大，注意磁盘空间
- 可以使用 logrotate 或类似工具管理日志

## 项目结构

```
canpulse/
├── internal/
│   ├── can/          # CAN协议层（帧解析、SLCAN、OBD-II）
│   ├── serial/       # 串口通信层（设备管理、收发）
│   ├── logger/       # 日志系统（双输出、按日期分类）
│   └── monitor/      # 监控分析层（变化检测、统计）
├── logs/             # 日志目录（自动创建）
├── main.go           # 主程序
├── go.mod            # Go模块定义
├── go.sum            # 依赖校验和
├── README.md         # 本文件
└── .gitignore        # Git配置
```

## 技术细节

### 支持的协议

- **SLCAN**: 串口CAN适配器标准协议
- **OBD-II**: ISO 15765-4 (CAN)

### 支持的设备

理论上支持所有兼容SLCAN协议的USB转CAN设备，已测试：
- 智潜物联 USB转CAN模块
- 周立功 USBCAN系列
- CANable (slcan模式)

### CAN波特率支持

- 10K, 20K, 50K, 100K, 125K, 250K, 500K, 800K, 1M
- 常见车辆: 500K (高速CAN), 125K (低速CAN)

## 许可证

MIT License

## 联系方式

- GitHub: https://github.com/Patrick7241/canpulse
- Issues: https://github.com/Patrick7241/canpulse/issues
