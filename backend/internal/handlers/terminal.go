package handlers

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize:  1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type TerminalSession struct {
	SSHClient  *ssh.Client
	SSHSession *ssh.Session
	Stdin      io.Writer
	Mu         sync.Mutex
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

			ip := parts[0]
			port := parts[1]
			username := parts[2]
			password := parts[3]

			session, err := h.connectSSH(ip, port, username, password)
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte("E"+err.Error()))
				continue
			}

			termSession = session
			conn.WriteMessage(websocket.TextMessage, []byte("Cconnected"))

			go h.forwardOutput(termSession, conn)

		case "I": // Input
			if termSession != nil {
				termSession.Mu.Lock()
				stdin := termSession.Stdin
				termSession.Mu.Unlock()

				if stdin != nil {
					stdin.Write([]byte(content))
				}
			}

		case "R": // Resize
			if termSession != nil {
				parts := splitString(content, "\n")
				if len(parts) == 2 {
					var cols, rows int
					fmt.Sscanf(parts[0], "%d", &cols)
					fmt.Sscanf(parts[1], "%d", &rows)
					termSession.SSHSession.WindowChange(rows, cols)
				}
			}

		case "D": // Disconnect
			if termSession != nil {
				termSession.SSHSession.Close()
				termSession.SSHClient.Close()
			}
		}
	}

	if termSession != nil {
		termSession.SSHSession.Close()
		termSession.SSHClient.Close()
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

func (h *TerminalHandler) forwardOutput(session *TerminalSession, conn *websocket.Conn) {
	session.Mu.Lock()
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
		io.Copy(connWriteWrapper{conn}, stdout)
	}()

	go func() {
		defer wg.Done()
		io.Copy(connWriteWrapper{conn}, stderr)
	}()

	wg.Wait()
}

type connWriteWrapper struct {
	conn *websocket.Conn
}

func (w connWriteWrapper) Write(p []byte) (n int, err error) {
	if len(p) > 0 {
		w.conn.WriteMessage(websocket.TextMessage, p)
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