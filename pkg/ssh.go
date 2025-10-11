package pkg

import (
	"GolangOM/constant"
	"GolangOM/logs"
	"GolangOM/ws"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

var LocalServer = &Server{
	ID:        constant.LocalServerID,
	IP:        "127.0.0.1",
	Status:    constant.Connected,
	SSHClient: nil,
}

// struct exposed for creating connections
type ServerConfig struct {
	ID         uint
	IP         string
	Port       int
	User       string
	AuthMethod constant.AuthMethod // password or key
	Credential string              // key path
	Password   string              // password or key password
}

// server struct
type Server struct {
	ID            uint
	IP            string
	Port          int
	User          string
	AuthMethod    constant.AuthMethod    // password or key
	Credential    string                 // key path
	Password      string                 // password or key password
	Status        constant.ConnectStatus // connected, disconnected, connecting
	LastCheckTime time.Time
	SSHClient     *ssh.Client
}

type ConnectionPool struct {
	maxConnectionNumber int
	servers             map[uint]*Server
	connectionNumber    int
	mutex               sync.RWMutex
	// connection pool configuration
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

func (c *ConnectionPool) GetServers() []*Server {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	servers := make([]*Server, 0, len(c.servers)+1)
	servers = append(servers, LocalServer)
	for _, server := range c.servers {
		servers = append(servers, server)
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
}

// create new connection
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

	// connection address format: ip:port
	addr := fmt.Sprintf("%s:%d", config.IP, config.Port)

	// establish SSH connection
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// get authentication method (password or key)
func getAuthMethods(config *ServerConfig) ([]ssh.AuthMethod, error) {
	var authMethods []ssh.AuthMethod

	if config.AuthMethod == constant.AuthMethodPassword {
		// password authentication
		authMethods = append(authMethods, ssh.Password(config.Password))
		return authMethods, nil
	} else if config.AuthMethod == constant.AuthMethodKey {
		// key authentication
		key, err := os.ReadFile(config.Credential)
		if err != nil {
			return nil, fmt.Errorf("load key file failed: %v", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			// try to handle key with password
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

// execute command
func (s *Server) ExecuteCommand(cmd string) (string, error) {
	if s.ID == constant.LocalServerID {
		// execute local command (capture stdout and stderr combined output)
		startTime := time.Now()
		output, err := exec.Command(cmd).CombinedOutput()
		result := string(output)

		if err != nil {
			// local command execution failed: return error and output details
			errMsg := fmt.Errorf("local command failed: %v, output: %s", err, result)
			return "", errMsg
		}

		// local command execution successful
		logs.Logger.Debug("local command execute success",
			zap.String("server_id", strconv.Itoa(int(s.ID))),
			zap.String("cmd", cmd),
			zap.Duration("time used", time.Since(startTime)),
			zap.Int("out length", len(result)))
		return result, nil
	}

	// check if SSH client is valid
	if !s.CheckSSHConnection() {
		return "", fmt.Errorf("SSH not init")
	}

	// create new session
	session, err := s.SSHClient.NewSession()
	if err != nil {
		return "", fmt.Errorf("create session failed: %v", err)
	}
	defer session.Close() // ensure session is closed

	// set command execution timeout
	done := make(chan struct{})
	go func() {
		select {
		case <-done:
			return
		case <-time.After(30 * time.Second): // 30 second timeout
			if err := session.Signal(ssh.SIGKILL); err != nil {
				logs.Logger.Warn("execute command failed: overtime ",
					zap.String("server_id", strconv.Itoa(int(s.ID))),
					zap.String("cmd", cmd),
					zap.Error(err))
			}
		}
	}()
	defer close(done)

	// capture command output
	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	// execute command
	startTime := time.Now()
	if err := session.Run(cmd); err != nil {
		// return error information when command execution fails
		return "", fmt.Errorf("execute command failed: %v, err out: %s", err, stderrBuf.String())
	}

	if stderrBuf.String() != "" {
		logs.Logger.Warn("command execute warning info:  ", zap.String("warn out", stderrBuf.String()))
	}

	// command execution successful
	result := stdoutBuf.String()
	logs.Logger.Debug("command execute error info:  ", zap.String("err out", stderrBuf.String()))
	logs.Logger.Debug("command execute successfully",
		zap.String("server_id", strconv.Itoa(int(s.ID))),
		zap.String("cmd", cmd),
		zap.Duration("time used", time.Since(startTime)),
		zap.Int("out string length", len(result)))

	return result, nil
}

// CheckSSHConnection check SSH connection status
// returns true if connection is valid, false if connection is disconnected
func (s *Server) CheckSSHConnection() bool {
	// 1. quick check first: whether client object is nil or status is disconnected
	if s.SSHClient == nil || s.Status != constant.Connected {
		logs.Logger.Debug("SSH connection invalid (client nil or status disconnected)",
			zap.String("server_id", strconv.Itoa(int(s.ID))))
		return false
	}

	// 2. send SSH heartbeat request (keepalive)
	// third parameter is request data (nil is fine), true means wait for server response
	_, _, err := s.SSHClient.SendRequest("keepalive@openssh.com", true, nil)
	if err != nil {
		connectionPool.mutex.Lock()
		// heartbeat failed: update status to disconnected, cleanup client
		s.Status = constant.Disconnected
		s.SSHClient.Close()
		s.SSHClient = nil
		connectionPool.mutex.Unlock()
		logs.Logger.Warn("SSH connection keepalive failed",
			zap.String("server_id", strconv.Itoa(int(s.ID))),
			zap.Error(err))
		return false
	}

	// 3. heartbeat successful: update last check time
	connectionPool.mutex.Lock()
	s.LastCheckTime = time.Now()
	connectionPool.mutex.Unlock()

	logs.Logger.Debug("SSH connection keepalive success",
		zap.String("server_id", strconv.Itoa(int(s.ID))))
	return true
}

// StartSSHConnectionCheckTicker start scheduled batch detection (interval configurable)
// interval: detection interval, in seconds
func (c *ConnectionPool) StartSSHConnectionCheckTicker(interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.RLock() // read lock: batch read servers, don't block other read operations
		servers := make(map[uint]*Server, len(c.servers))
		for k, v := range c.servers {
			// check if detection interval has been exceeded
			if v.LastCheckTime.Add(time.Duration(interval) * time.Second).Before(time.Now()) {
				servers[k] = v
			}
		}
		c.mutex.RUnlock()

		// batch detect SSH connection for each server
		for _, s := range servers {
			// try to reconnect when detection fails
			if !s.CheckSSHConnection() {
				// websocket broadcast server status
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
				// if reconnection successful, update server status
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
			} else {
				// update server check time in UI
				ws.SendMessage(ws.Message{
					ServerID:     s.ID,
					ServerStatus: s.Status,
				})
			}
		}

	}
}
