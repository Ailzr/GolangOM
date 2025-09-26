package ws

import (
	"GolangOM/constant"
	"GolangOM/logs"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 生产环境应配置具体域名
	},
}

type WsConnectionPool struct {
	mutex       sync.RWMutex
	Connections map[string]*websocket.Conn
}

var wsPool = &WsConnectionPool{
	Connections: make(map[string]*websocket.Conn),
	mutex:       sync.RWMutex{},
}

type Message struct {
	ServerID     uint                   `json:"server_id"`
	AppID        uint                   `json:"app_id"`
	ServerStatus constant.ConnectStatus `json:"server_status"`
	AppStatus    bool                   `json:"app_status"`
}

const (
	maxLiveTime       = 60 //允许ws连接最大的存活时间, 单位秒
	heartBeatInterval = 30 //心跳机制发送ping消息的时间, 单位秒
)

// 创建WebSocket路由处理
func WebsocketFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, ok := c.Get("username")
		if !ok {
			logs.Logger.Debug("user not login")
			return
		}
		handleWebSocketConnection(c, username.(string))
	}
}

// 处理WebSocket连接
func handleWebSocketConnection(c *gin.Context, username string) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logs.Logger.Error("WebSocket upgrade failed", zap.Error(err))
		return
	}
	defer conn.Close()

	// 注册用户连接
	storeConnection(username, conn)
	defer removeConnection(username)

	// 初始化连接设置
	initConnectionSettings(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// 开启心跳协程
	go StartHeartbeat(ctx, conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			// 打印详细的错误日志，方便调试
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				logs.Logger.Debug("client close websocket connection", zap.String("username", username))
			} else {
				logs.Logger.Warn("WebSocket read failed，connection may have been disconnected or timed out", zap.String("username", username), zap.Error(err))
			}
			break
		}
	}
}

// 存储连接信息
func storeConnection(username string, conn *websocket.Conn) {
	wsPool.mutex.Lock()
	defer wsPool.mutex.Unlock()
	if wsPool.Connections[username] != nil {
		return
	}
	wsPool.Connections[username] = conn
	logs.Logger.Debug("user connect", zap.String("username", username))
}

// 移除连接信息
func removeConnection(username string) {
	wsPool.mutex.Lock()
	defer wsPool.mutex.Unlock()
	delete(wsPool.Connections, username)
	logs.Logger.Debug("user disconnect", zap.String("username", username))
}

// 初始化连接设置
func initConnectionSettings(conn *websocket.Conn) {
	conn.SetReadLimit(512)
	if err := conn.SetReadDeadline(time.Now().Add(maxLiveTime * time.Second)); err != nil {
		logs.Logger.Error("reset ReadDeadline failed", zap.Error(err))
	}
	conn.SetPongHandler(func(appData string) error {
		logs.Logger.Debug("update heartbeat")
		// 每次收到 Pong 重置 ReadDeadline
		if err := conn.SetReadDeadline(time.Now().Add(maxLiveTime * time.Second)); err != nil {
			logs.Logger.Error("reset ReadDeadline failed", zap.Error(err))
			return err
		}
		return nil
	})
}

func StartHeartbeat(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(heartBeatInterval * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second))
			if err != nil {
				logs.Logger.Error("send heartbeat failed", zap.Error(err))
				return
			}
		}
	}
}

func SendMessage(msg Message) {
	wsPool.mutex.Lock()
	defer wsPool.mutex.Unlock()
	for _, conn := range wsPool.Connections {
		if err := conn.WriteJSON(msg); err != nil {
			logs.Logger.Error("message send failed", zap.Error(err))
		}
	}
}
