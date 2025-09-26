package pkg

import (
	"GolangOM/constant"
	"GolangOM/logs"
	"GolangOM/ws"
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

var LocalServer = &Server{
	ID:        constant.LocalServerID,
	IP:        "127.0.0.1",
	Status:    constant.Connected,
	SSHClient: nil,
}

// 对外暴露的用于创建链接的结构体
type ServerConfig struct {
	ID         uint
	IP         string
	Port       int
	User       string
	AuthMethod constant.AuthMethod // password 或 key
	Credential string              // 密钥路径
	Password   string              // 密码 或 密钥的密码
}

// 服务器结构体
type Server struct {
	ID            uint
	IP            string
	Port          int
	User          string
	AuthMethod    constant.AuthMethod    // password 或 key
	Credential    string                 // 密钥路径
	Password      string                 // 密码 或 密钥的密码
	Status        constant.ConnectStatus // connected, disconnected, connecting
	LastCheckTime time.Time
	SSHClient     *ssh.Client
}

type ConnectionPool struct {
	maxConnectionNumber int
	servers             map[uint]*Server
	connectionNumber    int
	mutex               sync.RWMutex
	// 连接池配置
}

var connectionPool = ConnectionPool{
	maxConnectionNumber: viper.GetInt("Server.MaxConnectionNumber"),
	servers:             make(map[uint]*Server),
	connectionNumber:    0,
	mutex:               sync.RWMutex{},
}

func init() {
	go connectionPool.StartSSHConnectionCheckTicker(viper.GetInt("Server.CheckInterval"))
}

func GetConnectionPool() *ConnectionPool {
	return &connectionPool
}

func (c *ConnectionPool) GetServerIDs() []uint {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	ids := make([]uint, 0, len(c.servers))
	ids = append(ids, constant.LocalServerID)
	for id := range c.servers {
		ids = append(ids, id)
	}
	return ids
}

func (c *ConnectionPool) GetServerByID(serverID uint) *Server {
	if _, ok := c.servers[serverID]; !ok {
		return nil
	}
	if serverID == constant.LocalServerID {
		return LocalServer
	}
	return c.servers[serverID]
}

func (c *ConnectionPool) GetServers() map[uint]*Server {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	servers := make(map[uint]*Server)
	servers[constant.LocalServerID] = LocalServer
	for id, server := range c.servers {
		servers[id] = server
	}
	return servers
}

func (c *ConnectionPool) AddServerToConnectionPool(server *Server) error {
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

func (c *ConnectionPool) RemoveServerFromConnectionPoolByID(serverID uint) {
	connectionPool.mutex.Lock()
	defer connectionPool.mutex.Unlock()
	if _, ok := connectionPool.servers[serverID]; !ok {
		return
	}
	if server := connectionPool.servers[serverID]; server.Status == "connected" && server.SSHClient != nil {
		server.SSHClient.Close()
	}
	delete(connectionPool.servers, serverID)
	connectionPool.connectionNumber--
	return
}

// 新建连接
func (c *ConnectionPool) NewConnection(config *ServerConfig) error {
	logs.Logger.Debug("NewConnection")

	server := &Server{
		ID:            config.ID,
		IP:            config.IP,
		Port:          config.Port,
		User:          config.User,
		AuthMethod:    config.AuthMethod,
		Credential:    config.Credential,
		Password:      config.Password,
		Status:        constant.Disconnected,
		LastCheckTime: time.Now(),
		SSHClient:     nil,
	}

	err := c.AddServerToConnectionPool(server)
	if err != nil {
		return err
	}

	client, err := sshConnect(config)
	if err != nil {
		return err
	}

	connectionPool.mutex.Lock()
	connectionPool.servers[server.ID].Status = constant.Connected
	connectionPool.servers[server.ID].SSHClient = client
	connectionPool.mutex.Unlock()

	logs.Logger.Info("SSH connect successfully", zap.String("ip", server.IP), zap.String("id", strconv.Itoa(int(server.ID))))

	return nil
}

func sshConnect(config *ServerConfig) (*ssh.Client, error) {
	authMethod, err := getAuthMethods(config)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return client, nil
}

// 获取认证方式（密码或密钥）
func getAuthMethods(config *ServerConfig) ([]ssh.AuthMethod, error) {
	var authMethods []ssh.AuthMethod

	if config.AuthMethod == constant.AuthMethodPassword {
		// 密码认证
		authMethods = append(authMethods, ssh.Password(config.Password))
		return authMethods, nil
	} else if config.AuthMethod == constant.AuthMethodKey {
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
	if s.ID == constant.LocalServerID {
		// 执行本地命令（捕获 stdout 和 stderr 合并输出）
		startTime := time.Now()
		output, err := exec.Command(cmd).CombinedOutput()
		result := string(output)

		if err != nil {
			// 本地命令执行失败：返回错误和输出详情
			errMsg := fmt.Errorf("local command failed: %v, output: %s", err, result)
			return "", errMsg
		}

		// 本地命令执行成功
		logs.Logger.Debug("local command execute success",
			zap.String("server_id", strconv.Itoa(int(s.ID))),
			zap.String("cmd", cmd),
			zap.Duration("time used", time.Since(startTime)),
			zap.Int("out length", len(result)))
		return result, nil
	}

	// 检查SSH客户端是否有效
	if !s.CheckSSHConnection() {
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
					zap.String("server_id", strconv.Itoa(int(s.ID))),
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
	logs.Logger.Debug("command execute successfully",
		zap.String("server_id", strconv.Itoa(int(s.ID))),
		zap.String("cmd", cmd),
		zap.Duration("time used", time.Since(startTime)),
		zap.Int("out string length", len(result)))

	return result, nil
}

// CheckSSHConnection 检测 SSH 连接状态
// 返回 true 表示连接有效，false 表示连接已断开
func (s *Server) CheckSSHConnection() bool {
	// 1. 先快速判断：客户端对象是否为 nil 或状态标记为断开
	if s.SSHClient == nil || s.Status != constant.Connected {
		logs.Logger.Debug("SSH connection invalid (client nil or status disconnected)",
			zap.String("server_id", strconv.Itoa(int(s.ID))))
		return false
	}

	// 2. 发送 SSH 心跳请求（keepalive）
	// 第三个参数为请求数据（nil 即可），true 表示等待服务器响应
	_, _, err := s.SSHClient.SendRequest("keepalive@openssh.com", true, nil)
	if err != nil {
		connectionPool.mutex.Lock()
		// 心跳失败：更新状态为断开，清理客户端
		s.Status = constant.Disconnected
		s.SSHClient.Close()
		s.SSHClient = nil
		connectionPool.mutex.Unlock()
		logs.Logger.Warn("SSH connection keepalive failed",
			zap.String("server_id", strconv.Itoa(int(s.ID))),
			zap.Error(err))
		return false
	}

	// 3. 心跳成功：更新最后检查时间
	connectionPool.mutex.Lock()
	s.LastCheckTime = time.Now()
	connectionPool.mutex.Unlock()

	logs.Logger.Debug("SSH connection keepalive success",
		zap.String("server_id", strconv.Itoa(int(s.ID))))
	return true
}

// StartSSHConnectionCheckTicker 启动定时批量检测（间隔可配置）
// interval: 检测间隔，单位秒
func (c *ConnectionPool) StartSSHConnectionCheckTicker(interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.RLock() // 读锁：批量读取服务器，不阻塞其他读操作
		servers := make(map[uint]*Server, len(c.servers))
		for k, v := range c.servers {
			// 判断是否已超过检测间隔
			if v.LastCheckTime.Add(time.Duration(interval) * time.Second).Before(time.Now()) {
				servers[k] = v
			}
		}
		c.mutex.RUnlock()

		// 批量检测每个服务器的 SSH 连接
		for _, s := range servers {
			// 检测失败时尝试重连
			if !s.CheckSSHConnection() {
				// websocket广播服务器状态
				ws.SendMessage(ws.Message{
					ServerID:     s.ID,
					ServerStatus: s.Status,
				})

				logs.Logger.Info("try to reconnect SSH server",
					zap.String("server_id", strconv.Itoa(int(s.ID))))

				client, err := sshConnect(&ServerConfig{
					AuthMethod: s.AuthMethod,
					Credential: s.Credential,
					IP:         s.IP,
					Password:   s.Password,
					Port:       s.Port,
					User:       s.User,
				})
				if err != nil {
					logs.Logger.Error("reconnect SSH server failed",
						zap.String("server_id", strconv.Itoa(int(s.ID))),
						zap.String("server_ip", s.IP),
						zap.Error(err),
					)
					continue
				}
				// 如果重连成功，更新服务器状态
				c.mutex.Lock()
				if existingServer, exists := c.servers[s.ID]; exists {
					existingServer.SSHClient = client
					existingServer.LastCheckTime = time.Now()
					existingServer.Status = constant.Connected
					ws.SendMessage(ws.Message{
						ServerID:     s.ID,
						ServerStatus: constant.Connected,
					})
				}
				c.mutex.Unlock()
			}
		}

	}
}
