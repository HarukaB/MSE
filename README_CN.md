# MSE (Muse Save Editor)

这是一个用 Go 语言编写的命令行工具，用于序列化和反序列化 Muse Dash 的存档文件 (`.sav`)。它支持将二进制存档转换为可读的 JSON 格式以便编辑，并能将其还原为游戏可识别的二进制格式。

## 功能特性

- **SAV 转 JSON**：将 Muse Dash 二进制存档转换为人类可读的 JSON 格式。
- **JSON 转 SAV**：将修改后的 JSON 文件转回游戏使用的二进制格式。

## 安装指南

### 从 Releases 下载（推荐）

直接从 [Releases](../../releases) 页面下载适用于 Windows、macOS 和 Linux 的预编译版本。下载后即可直接运行。

### 从源码编译

如需从源码编译，请确保已安装 [Go](https://golang.org/dl/)（版本 1.21 或更高）。

1. 克隆仓库或下载源代码。
2. 编译可执行文件：
   ```bash
   go build -o mse main.go
   ```

## 使用方法

本工具提供两个主要命令：`to-json` 和 `to-sav`。

### 将 SAV 转换为 JSON

若要查看或编辑存档，请将其转换为 JSON：

```bash
./mse to-json input.sav output.json
```

### 将 JSON 转换为 SAV

在 JSON 文件中完成修改后，将其转回二进制 `.sav` 格式：

```bash
./mse to-sav input.json output.sav
```

## 项目结构

- `main.go`：CLI 应用程序入口。
- `converter/`：高层转换逻辑。
- `odin/`：Odin 序列化格式的核心实现。
  - `reader.go`：用于解析 `.sav` 文件的二进制数据读取器。
  - `writer.go`：用于生成 `.sav` 文件的二进制数据写入器。
  - `node.go`：表示序列化树的数据结构。
  - `entry.go`：Odin 格式的常量和类型定义。

## 免责声明

本工具仅用于教育和 Mod 开发目的。在修改存档之前，请务必备份原始文件。作者对因使用本工具而导致的任何数据丢失或账号问题概不负责。
