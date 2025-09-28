package ws

import (
	"GolangOM/constant"
	"GolangOM/logs"
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // production environment should configure specific domain
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
	maxLiveTime       = 60 // maximum survival time for ws connection, in seconds
	heartBeatInterval = 30 // heartbeat mechanism ping message time, in seconds
)

// create WebSocket route handler
func WebsocketFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, ok := c.Get("username")
		if !ok {
			logs.Logger.Debug("user not login")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not logged in"})
			return
		}
		handleWebSocketConnection(c, username.(string))
	}
}

// handle WebSocket connection
func handleWebSocketConnection(c *gin.Context, username string) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logs.Logger.Error("WebSocket upgrade failed", zap.Error(err))
		return
	}
	defer conn.Close()

	// register user connection
	storeConnection(username, conn)
	defer removeConnection(username)

	// initialize connection settings
	initConnectionSettings(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// start heartbeat goroutine
	go StartHeartbeat(ctx, conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			// print detailed error logs for debugging
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				logs.Logger.Debug("client close websocket connection", zap.String("username", username))
			} else {
				logs.Logger.Warn("WebSocket read failed，connection may have been disconnected or timed out", zap.String("username", username), zap.Error(err))
			}
			break
		}
	}
}

// store connection information
func storeConnection(username string, conn *websocket.Conn) {
	wsPool.mutex.Lock()
	defer wsPool.mutex.Unlock()
	if wsPool.Connections[username] != nil {
		return
	}
	wsPool.Connections[username] = conn
	logs.Logger.Debug("user connect", zap.String("username", username))
}

// remove connection information
func removeConnection(username string) {
	wsPool.mutex.Lock()
	defer wsPool.mutex.Unlock()
	delete(wsPool.Connections, username)
	logs.Logger.Debug("user disconnect", zap.String("username", username))
}

// initialize connection settings
func initConnectionSettings(conn *websocket.Conn) {
	conn.SetReadLimit(512)
	if err := conn.SetReadDeadline(time.Now().Add(maxLiveTime * time.Second)); err != nil {
		logs.Logger.Error("reset ReadDeadline failed", zap.Error(err))
	}
	conn.SetPongHandler(func(appData string) error {
		logs.Logger.Debug("update heartbeat")
		// reset ReadDeadline every time Pong is received
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
