package pkg

import (
	"GolangOM/logs"
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	"os"
	"sync"
	"time"
)

type AuthMethod string

const (
	AuthMethodPassword AuthMethod = "password" // 密码验证
	AuthMethodKey      AuthMethod = "key"      // 密钥验证
)

// 对外暴露的用于创建链接的结构体
type ServerConfig struct {
	IP         string
	Port       int
	User       string
	AuthMethod AuthMethod // password 或 key
	Credential string     // 密钥路径
	Password   string     // 密码 或 密钥的密码
}

// 服务器结构体
type Server struct {
	ID            string
	IP            string
	Port          int
	User          string
	AuthMethod    AuthMethod // password 或 key
	Credential    string     // 密钥路径
	Password      string     // 密码 或 密钥的密码
	Status        string     // connected, disconnected, connecting
	LastCheckTime time.Time
	SSHClient     *ssh.Client
}

type ConnectionPool struct {
	maxConnectionNumber int
	servers             map[string]*Server
	connectionNumber    int
	mutex               sync.RWMutex
	// 连接池配置
}

var connectionPool = ConnectionPool{
	maxConnectionNumber: viper.GetInt("Server.MaxConnectionNumber"),
	servers:             make(map[string]*Server),
	connectionNumber:    0,
	mutex:               sync.RWMutex{},
}

func GetConnectionPool() *ConnectionPool {
	return &connectionPool
}

func (c *ConnectionPool) GetServerIDs() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	ids := make([]string, 0, len(c.servers))
	for id := range c.servers {
		ids = append(ids, id)
	}
	return ids
}

func (c *ConnectionPool) GetServerByID(serverID string) (*Server, error) {
	return c.servers[serverID], nil
}

func addServerToConnectionPool(server *Server) error {
	connectionPool.mutex.Lock()
	defer connectionPool.mutex.Unlock()
	if connectionPool.connectionNumber >= connectionPool.maxConnectionNumber {
		return fmt.Errorf("over max connection number")
	}
	if _, ok := connectionPool.servers[server.ID]; ok {
		return fmt.Errorf("server already exists")
	}
	connectionPool.servers[server.ID] = server
	connectionPool.connectionNumber++
	return nil
}

// 暂时不提供删除服务器的函数
//func removeServerFromConnectionPoolByID(serverID string) error {
//	connectionPool.mutex.Lock()
//	defer connectionPool.mutex.Unlock()
//	if _, ok := connectionPool.servers[serverID]; !ok {
//		return fmt.Errorf("server not exists")
//	}
//	if server := connectionPool.servers[serverID]; server.Status == "connected" {
//		server.SSHClient.Close()
//	}
//	delete(connectionPool.servers, serverID)
//	connectionPool.connectionNumber--
//	return nil
//}

// 新建连接
func NewConnection(config *ServerConfig) error {
	logs.Logger.Debug("NewConnection")

	authMethod, err := getAuthMethods(config)
	if err != nil {
		return err
	}

	clientConfig := &ssh.ClientConfig{
		User:            config.User,
		Auth:            authMethod,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// 连接地址格式：ip:port
	addr := fmt.Sprintf("%s:%d", config.IP, config.Port)

	// 建立SSH连接
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return err
	}

	server := &Server{
		IP:            config.IP,
		Port:          config.Port,
		User:          config.User,
		AuthMethod:    config.AuthMethod,
		Credential:    config.Credential,
		Password:      config.Password,
		Status:        "connected",
		LastCheckTime: time.Now(),
		SSHClient:     client,
	}

	// 保存连接并生成UUID
	server.SSHClient = client
	server.ID = uuid.New().String()
	err = addServerToConnectionPool(server)
	if err != nil {
		return err
	}

	logs.Logger.Info("SSH connect successfully", zap.String("address", addr), zap.String("id", server.ID))

	return nil
}

// 获取认证方式（密码或密钥）
func getAuthMethods(config *ServerConfig) ([]ssh.AuthMethod, error) {
	var authMethods []ssh.AuthMethod

	if config.AuthMethod == AuthMethodPassword {
		// 密码认证
		authMethods = append(authMethods, ssh.Password(config.Password))
		return authMethods, nil
	} else if config.AuthMethod == AuthMethodKey {
		// 密钥认证
		key, err := os.ReadFile(config.Credential)
		if err != nil {
			return nil, fmt.Errorf("load key file failed: %v", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			// 尝试处理带密码的密钥
			if config.Password != "" {
				signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(config.Password))
			}
			if err != nil {
				return nil, fmt.Errorf("parse key failed: %v", err)
			}
		}

		authMethods = append(authMethods, ssh.PublicKeys(signer))
		return authMethods, nil

	}

	return nil, fmt.Errorf("not exists auth method")
}

// 执行命令
func (s *Server) ExecuteCommand(cmd string) (string, error) {
	// 检查SSH客户端是否有效
	if s.SSHClient == nil {
		return "", fmt.Errorf("SSH not init")
	}

	// 创建新的会话
	session, err := s.SSHClient.NewSession()
	if err != nil {
		return "", fmt.Errorf("create session failed: %v", err)
	}
	defer session.Close() // 确保会话关闭

	// 设置命令执行超时
	done := make(chan struct{})
	go func() {
		select {
		case <-done:
			return
		case <-time.After(30 * time.Second): // 30秒超时
			if err := session.Signal(ssh.SIGKILL); err != nil {
				logs.Logger.Warn("execute command failed: overtime ",
					zap.String("server_id", s.ID),
					zap.String("cmd", cmd),
					zap.Error(err))
			}
		}
	}()
	defer close(done)

	// 捕获命令输出
	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	// 执行命令
	startTime := time.Now()
	if err := session.Run(cmd); err != nil {
		// 命令执行出错时，返回错误信息
		return "", fmt.Errorf("execute command failed: %v, err out: %s", err, stderrBuf.String())
	}

	// 命令执行成功
	result := stdoutBuf.String()
	logs.Logger.Info("command execute successfully",
		zap.String("server_id", s.ID),
		zap.String("cmd", cmd),
		zap.Duration("time used", time.Since(startTime)),
		zap.Int("out string length", len(result)))

	return result, nil
}

// 上传脚本并执行
func (s *Server) UploadAndExecuteScript(scriptContent string, scriptName string) (string, error) {
	return "", nil
}
