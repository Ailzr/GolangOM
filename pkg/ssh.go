package pkg

import (
	"GolangOM/logs"
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
	return cmd + "execute successfully", nil
}

// 上传脚本并执行
func (s *Server) UploadAndExecuteScript(scriptContent string, scriptName string) (string, error) {
	return "", nil
}
