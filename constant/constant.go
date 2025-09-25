package constant

type AuthMethod string

const (
	AuthMethodPassword AuthMethod = "password" // 密码验证
	AuthMethodKey      AuthMethod = "key"      // 密钥验证
)

type ConnectStatus string

const (
	Connected    ConnectStatus = "connected"
	Disconnected ConnectStatus = "disconnected"
	Connecting   ConnectStatus = "connecting"
)

type AppCheckType string

const (
	AppCheckTypePid  AppCheckType = "pid"
	AppCheckTypePort AppCheckType = "port"
	AppCheckTypeHttp AppCheckType = "http"
)

// 本地服务器ID
const LocalServerID = "LocalHostServer"
