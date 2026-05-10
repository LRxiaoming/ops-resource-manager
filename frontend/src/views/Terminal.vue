<template>
  <div class="terminal-page">
    <div class="terminal-header">
      <div class="connection-info">
        <span v-if="assetIp">{{ assetIp }}</span>
        <span v-else>未连接</span>
        <span v-if="username" class="username"> - {{ username }}</span>
      </div>
      <div class="actions">
        <el-button v-if="!connected" type="primary" @click="showConnectDialog = true">
          连接
        </el-button>
        <el-button v-else type="danger" @click="disconnect">
          断开
        </el-button>
      </div>
    </div>

    <div v-if="!connected && !showConnectDialog" class="connect-prompt">
      <p>点击"连接"按钮输入SSH信息</p>
    </div>

    <div v-if="showConnectDialog" class="connect-dialog">
      <h3>SSH 连接</h3>
      <el-form :model="connectForm" label-width="80px">
        <el-form-item label="IP地址">
          <el-input v-model="connectForm.ip" placeholder="192.168.1.100" />
        </el-form-item>
        <el-form-item label="端口">
          <el-input v-model="connectForm.port" placeholder="22" type="number" />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="connectForm.username" placeholder="root" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="connectForm.password" type="password" placeholder="密码" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="connect">连接</el-button>
          <el-button @click="showConnectDialog = false">取消</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div ref="terminalRef" class="terminal-container"></div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useRoute } from 'vue-router'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'

const route = useRoute()

const terminalRef = ref(null)
const connected = ref(false)
const showConnectDialog = ref(true)
const assetIp = ref(route.query.ip || '')
const username = ref('')

const connectForm = ref({
  ip: route.query.ip || '',
  port: route.query.port || '22',
  username: '',
  password: ''
})

let term = null
let fitAddon = null
let ws = null

onMounted(() => {
  term = new Terminal({
    cursorBlink: true,
    fontSize: 14,
    fontFamily: 'Consolas, "Courier New", monospace',
    theme: {
      background: '#1e1e1e',
      foreground: '#ffffff'
    }
  })

  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)

  term.open(terminalRef.value)
  fitAddon.fit()

  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  if (ws) ws.close()
  if (term) term.dispose()
})

const handleResize = () => {
  if (fitAddon) fitAddon.fit()
}

const connect = () => {
  if (!connectForm.value.ip || !connectForm.value.username || !connectForm.value.password) {
    return
  }

  const wsUrl = `ws://${window.location.hostname}:8080/ws/terminal`
  ws = new WebSocket(wsUrl)

  ws.onopen = () => {
    const msg = `C\n${connectForm.value.ip}\n${connectForm.value.port}\n${connectForm.value.username}\n${connectForm.value.password}`
    ws.send(msg)
  }

  ws.onmessage = (event) => {
    const data = event.data
    if (data === 'Cconnected') {
      connected.value = true
      showConnectDialog.value = false
      assetIp.value = connectForm.value.ip
      username.value = connectForm.value.username
      term.write('\r\n\x1b[32m连接成功\x1b[0m\r\n\r\n')
    } else if (data.startsWith('E')) {
      term.write(`\r\n\x1b[31m错误: ${data.substring(1)}\x1b[0m\r\n`)
    } else {
      term.write(data)
    }
  }

  ws.onclose = () => {
    connected.value = false
    showConnectDialog.value = false
    term.write('\r\n\x1b[33m连接已断开\x1b[0m\r\n')
  }

  ws.onerror = (err) => {
    term.write(`\r\n\x1b[31m连接失败\x1b[0m\r\n`)
  }

  term.onData((data) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send('I' + data)
    }
  })

  term.onResize(({ cols, rows }) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(`R\n${cols}\n${rows}`)
    }
  })
}

const disconnect = () => {
  if (ws) {
    ws.send('D')
    ws.close()
    ws = null
  }
  connected.value = false
  showConnectDialog.value = true
  connectForm.value.password = ''
}
</script>

<style scoped>
.terminal-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 120px);
  background: #1e1e1e;
  border-radius: 4px;
  overflow: hidden;
}

.terminal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 20px;
  background: #2d2d2d;
  color: #fff;
}

.connection-info {
  font-size: 14px;
}

.connection-info .username {
  color: #4fc3f7;
}

.connect-prompt {
  display: flex;
  align-items: center;
  justify-content: center;
  flex: 1;
  color: #888;
}

.connect-dialog {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: #2d2d2d;
  padding: 30px;
  border-radius: 8px;
  z-index: 100;
  color: #fff;
}

.connect-dialog h3 {
  margin-bottom: 20px;
  color: #fff;
}

.connect-dialog :deep(.el-form-item__label) {
  color: #fff;
}

.connect-dialog :deep(.el-input) {
  width: 250px;
}

.terminal-container {
  flex: 1;
  padding: 10px;
  overflow: hidden;
}

.terminal-container :deep(.xterm) {
  height: 100%;
}

.terminal-container :deep(.xterm-viewport) {
  overflow-y: auto;
}
</style>