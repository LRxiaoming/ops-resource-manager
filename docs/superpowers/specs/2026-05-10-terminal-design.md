# 远程终端功能设计

## 背景

为运维系统新增远程终端功能，允许用户通过浏览器直接连接 Linux 服务器进行操作。

## 需求

- 用户输入 IP、用户名、密码连接 SSH
- 后端代理 SSH 连接，前端通过 WebSocket 交互
- 完整终端界面（全屏、多标签页、复制粘贴、命令历史）

## 技术方案

### 架构

```
浏览器 <-> 后端(8080) <-> Linux服务器(22)
         WebSocket     SSH
```

### 后端实现

**依赖库**: `golang.org/x/crypto/ssh`

**新增 API**:

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/terminal/connect` | 验证资产是否存在，建立 SSH 连接 |
| GET | `/ws/terminal` | WebSocket 端点，实时交互 |

**TerminalSession 模型**:
```go
type TerminalSession struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    AssetID   uint      `json:"asset_id"`
    UserID    uint      `json:"user_id"`
    IP        string    `json:"ip"`
    Username  string    `json:"username"`
    Status    string    `json:"status"` // connected, disconnected
    ConnectedAt  time.Time `json:"connected_at"`
    DisconnectedAt *time.Time `json:"disconnected_at,omitempty"`
}
```

**SSH 连接流程**:
1. 接收前端 WebSocket 连接
2. 接收连接参数 (IP, port, username, password)
3. 建立 SSH 客户端
4. 开启交互式 shell session
5. 双向转发数据：WebSocket <-> SSH

### 前端实现

**依赖库**: `xterm`, `xterm-addon-fit`, `xterm-addon-web-links`

**页面**: `Terminal.vue`
- 顶部：连接信息（IP、用户名）+ 断开按钮
- 主体：xterm.js 终端
- 自动连接：基于路由参数自动连接指定资产

**组件**:
```vue
<template>
  <div class="terminal-page">
    <div class="terminal-header">
      <span>{{ assetIp }} - {{ username }}</span>
      <el-button @click="disconnect">断开</el-button>
    </div>
    <div ref="terminalRef" class="terminal-container"></div>
  </div>
</template>
```

### 界面入口

资产列表增加"终端"按钮，点击跳转 `/terminal/:assetId?ip=&port=22`

## 文件变更

### 后端
- `internal/models/models.go` - 新增 TerminalSession
- `internal/handlers/terminal.go` - 新增终端 handler
- `cmd/server/main.go` - 添加终端路由

### 前端
- `src/views/Terminal.vue` - 新增终端页面
- `src/router/index.js` - 添加终端路由

## 验证

1. 启动后端和前端
2. 进入资产列表，点击某资产的"终端"按钮
3. 输入用户名密码（使用资产记录的 IP，默认端口 22）
4. 成功连接后执行命令验证