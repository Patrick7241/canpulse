# 新功能更新说明

## 版本信息
- 更新日期: 2026-03-10
- Git仓库: https://github.com/Patrick7241/canpulse.git
- 分支: main

## 🎉 新增功能

### 1. CAN消息发送功能

**功能说明**: 支持向CAN总线发送自定义消息，用于测试和控制

**使用方法**:
```bash
# 发送格式: ID:DATA (十六进制)
./canpulse -send "123:11223344"

# 指定串口和波特率
./canpulse -port /dev/cu.usbserial-xxx -send "123:11223344"

# 发送到不同CAN波特率的总线
./canpulse -canbitrate 250K -send "456:AABBCCDD"
```

**应用场景**:
- 测试车辆控制信号
- 模拟CAN消息
- 验证信号识别结果
- 开发车载应用

**实现细节**:
- `internal/can/frame.go`: 添加 `ToSLCAN()` 和 `NewFrame()` 方法
- `internal/serial/device.go`: 添加 `Send()` 方法
- `main.go`: 添加 `sendCANFrame()` 函数和 `-send` 参数

### 2. 完整的日志系统

**功能说明**: 双输出日志系统，同时记录到控制台和文件

**特性**:
- ✅ 同时输出到控制台和文件
- ✅ 按日期自动分类日志文件
- ✅ 支持多个日志级别 (DEBUG, INFO, WARN, ERROR)
- ✅ 自动创建日志目录
- ✅ 线程安全

**日志文件格式**:
```
logs/
└── canpulse-2026-03-10.log   # 按日期命名
```

**日志内容示例**:
```
[2026-03-10 10:58:23] [INFO] 正在连接CAN设备...
[2026-03-10 10:58:23] [INFO] 已连接: /dev/cu.usbserial-14140 (波特率: 115200, CAN: 500K)
[2026-03-10 10:58:24] [INFO] 开始监听CAN总线数据...
[2026-03-10 10:58:25] [123] 14:23:45.123 | 标准帧 | ID: 0x123 | DLC: 8 | Hex: [00 11 22 33 44 55 66 77]
```

**配置选项**:
```bash
# 指定日志目录
./canpulse -logdir custom_logs

# 默认日志目录为 logs/
./canpulse
```

**实现细节**:
- `internal/logger/logger.go`: 完整的日志系统实现
- 集成到所有模块: `main.go`, `monitor.go`, `stats.go`

## 📝 命令行参数更新

新增参数:
- `-send string`: 发送CAN帧 (格式: ID:DATA, 例如: 123:11223344)
- `-logdir string`: 日志文件目录 (默认: logs)

完整参数列表:
```
Usage of ./canpulse:
  -baud int
        串口波特率 (default 115200)
  -binary
        显示二进制格式
  -canbitrate string
        CAN总线波特率 (125K, 250K, 500K, 1M) (default "500K")
  -detect
        变化检测模式：只显示数据发生变化的帧
  -list
        列出所有可用串口
  -log string
        保存CAN数据到文件 (例如: can_log.csv)
  -logdir string
        日志文件目录 (default "logs")
  -mincount int
        统计时过滤：只显示帧数>=此值的ID (default 1)
  -port string
        串口设备名称 (例如: /dev/cu.usbserial-xxx)
  -send string
        发送CAN帧 (格式: ID:DATA, 例如: 123:11223344)
  -stats int
        每N秒显示一次统计信息 (0=不显示)
```

## 🔧 使用示例

### 示例1: 发送CAN消息测试转向灯

```bash
# 1. 先监听，找到转向灯的CAN ID (假设是0x294)
./canpulse -detect -binary

# 2. 发送测试消息，模拟左转向灯
./canpulse -send "294:00200000"

# 3. 观察车辆是否响应
```

### 示例2: 完整的调试会话

```bash
# 启动监听并记录所有日志
./canpulse -detect -binary -log test.csv -stats 10 -logdir debug_logs

# 操作车辆功能（如开关空调）
# 观察变化的CAN ID

# 测试发送消息验证
./canpulse -send "ABC:11223344"

# 查看日志文件
cat debug_logs/canpulse-2026-03-10.log
```

### 示例3: 批量测试发送

```bash
# 创建测试脚本
cat > test_send.sh << 'EOF'
#!/bin/bash
./canpulse -send "100:0000000000000000"
sleep 1
./canpulse -send "100:FFFFFFFFFFFFFFFF"
sleep 1
./canpulse -send "100:0000000000000000"
EOF

chmod +x test_send.sh
./test_send.sh
```

## 🎯 应用场景

### 1. 车身信号逆向工程

```bash
# 步骤1: 监听并记录
./canpulse -detect -binary -log signals.csv > console.log

# 步骤2: 操作车辆功能，观察日志
# 找到对应的CAN ID和数据位

# 步骤3: 验证发现的信号
./canpulse -send "ID:DATA"
```

### 2. 自动化测试

```bash
# 持续监听 + 定期统计
./canpulse -stats 30 -mincount 50 &

# 发送测试序列
for i in {1..10}; do
  ./canpulse -send "123:000000000000000$i"
  sleep 2
done
```

### 3. 故障排查

所有操作都会记录到日志文件，便于事后分析：
```bash
# 查看今天的所有日志
cat logs/canpulse-$(date +%Y-%m-%d).log

# 搜索特定ID的日志
grep "ID: 0x123" logs/canpulse-2026-03-10.log

# 查看错误日志
grep "\[ERROR\]" logs/canpulse-2026-03-10.log
```

## 📊 项目统计

**代码行数**:
- 总计: ~1200行Go代码
- CAN协议层: 324行
- 串口通信层: 179行
- 日志系统: 271行
- 监控分析层: 322行
- 主程序: 303行

**Git提交记录**:
```
f1e696f feat: add main application with CLI interface
4db80d4 feat: add CAN monitoring and statistics module
7f6da28 feat: add comprehensive logging system
8776bec feat: add serial communication layer
6d4bc22 feat: add CAN protocol layer with SLCAN and OBD-II support
83e4289 Initial commit: project setup and documentation
```

## 🔐 安全注意事项

**发送功能的使用警告**:
- ⚠️ **谨慎发送**: CAN消息会直接发送到车辆总线，可能影响车辆行为
- ⚠️ **静止测试**: 建议在车辆静止时进行发送测试
- ⚠️ **了解协议**: 发送前确保了解消息的含义
- ⚠️ **小范围测试**: 先用无害的ID进行测试
- ⚠️ **备份重要数据**: 测试前备份车辆设置

**日志系统注意事项**:
- 📁 日志文件可能包含敏感的车辆数据
- 📁 定期清理旧日志文件以节省空间
- 📁 logs/ 目录已添加到 .gitignore

## 🚀 后续改进计划

### 短期 (已完成)
- ✅ CAN消息发送功能
- ✅ 完整的日志系统
- ✅ 按日期分类日志文件

### 中期
- [ ] 批量发送脚本支持
- [ ] 消息重放功能
- [ ] 日志分析工具
- [ ] Web UI界面

### 长期
- [ ] CAN FD支持
- [ ] 多CAN接口支持
- [ ] 云端数据分析
- [ ] 机器学习信号识别

## 📚 相关文档

- [README.md](README.md) - 完整使用指南
- [ARCHITECTURE.md](ARCHITECTURE.md) - 架构设计文档
- [.gitignore](.gitignore) - Git配置

## 🔗 链接

- GitHub仓库: https://github.com/Patrick7241/canpulse
- Issue报告: https://github.com/Patrick7241/canpulse/issues

---

**最后更新**: 2026-03-10
**版本**: v1.0.0 with CAN send & logging
**作者**: Patrick & Claude Sonnet 4.5
