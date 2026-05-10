package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type TerminalSession struct {
	SSHClient  *ssh.Client
	SSHSession *ssh.Session
	Stdin      io.Writer
	Mu         sync.Mutex
	closed     atomic.Bool
}

type TerminalHandler struct{}

func NewTerminalHandler() *TerminalHandler {
	return &TerminalHandler{}
}

func (h *TerminalHandler) HandleTerminal(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to upgrade connection"})
		return
	}
	defer conn.Close()

	var termSession *TerminalSession
	var sessionMu sync.Mutex

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		if len(msg) < 1 {
			continue
		}

		msgType := string(msg[:1])
		content := string(msg[1:])

		switch msgType {
		case "C": // Connect
			parts := splitString(content, "\n")
			if len(parts) < 4 {
				conn.WriteMessage(websocket.TextMessage, []byte("Einvalid format"))
				continue
			}

			ip := strings.TrimSpace(parts[0])
			port := strings.TrimSpace(parts[1])
			username := strings.TrimSpace(parts[2])
			password := strings.TrimSpace(parts[3])

			session, err := h.connectSSH(ip, port, username, password)
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte("E"+err.Error()))
				continue
			}

			sessionMu.Lock()
			termSession = session
			sessionMu.Unlock()

			conn.WriteMessage(websocket.TextMessage, []byte("Cconnected"))

			go h.forwardOutput(termSession, conn, &sessionMu)

		case "I": // Input
			sessionMu.Lock()
			session := termSession
			sessionMu.Unlock()

			if session != nil && !session.closed.Load() {
				session.Mu.Lock()
				stdin := session.Stdin
				session.Mu.Unlock()

				if stdin != nil {
					stdin.Write([]byte(content))
				}
			}

		case "R": // Resize
			sessionMu.Lock()
			session := termSession
			sessionMu.Unlock()

			if session != nil && !session.closed.Load() {
				parts := splitString(content, "\n")
				if len(parts) == 2 {
					var cols, rows int
					fmt.Sscanf(parts[0], "%d", &cols)
					fmt.Sscanf(parts[1], "%d", &rows)
					session.SSHSession.WindowChange(rows, cols)
				}
			}

		case "D": // Disconnect
			sessionMu.Lock()
			session := termSession
			sessionMu.Unlock()

			if session != nil {
				session.closed.Store(true)
				session.SSHSession.Close()
				session.SSHClient.Close()
			}
		}
	}

	// Final cleanup
	sessionMu.Lock()
	session := termSession
	sessionMu.Unlock()

	if session != nil {
		session.closed.Store(true)
		session.SSHSession.Close()
		session.SSHClient.Close()
	}
}

func (h *TerminalHandler) connectSSH(ip, port, username, password string) (*TerminalSession, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%s", ip, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("连接失败: %v", err)
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("创建会话失败: %v", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	err = session.RequestPty("xterm", 80, 30, modes)
	if err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Errorf("请求pty失败: %v", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Errorf("获取stdin失败: %v", err)
	}

	err = session.Shell()
	if err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Errorf("启动shell失败: %v", err)
	}

	return &TerminalSession{
		SSHClient:  client,
		SSHSession: session,
		Stdin:      stdin,
	}, nil
}

func (h *TerminalHandler) forwardOutput(session *TerminalSession, conn *websocket.Conn, sessionMu *sync.Mutex) {
	// First check if session is already closed
	if session.closed.Load() {
		return
	}

	session.Mu.Lock()
	if session.closed.Load() {
		session.Mu.Unlock()
		return
	}
	stdout, err := session.SSHSession.StdoutPipe()
	if err != nil {
		session.Mu.Unlock()
		return
	}
	stderr, err := session.SSHSession.StderrPipe()
	if err != nil {
		session.Mu.Unlock()
		return
	}
	session.Mu.Unlock()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(connWriteWrapper{conn, session}, stdout)
	}()

	go func() {
		defer wg.Done()
		io.Copy(connWriteWrapper{conn, session}, stderr)
	}()

	wg.Wait()
}

type connWriteWrapper struct {
	conn    *websocket.Conn
	session *TerminalSession
}

func (w connWriteWrapper) Write(p []byte) (n int, err error) {
	if len(p) == 0 || w.session.closed.Load() {
		return 0, io.EOF
	}
	err = w.conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		w.session.closed.Store(true)
		return 0, err
	}
	return len(p), nil
}

func splitString(s string, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i = start - 1
		}
	}
	result = append(result, s[start:])
	return result
}