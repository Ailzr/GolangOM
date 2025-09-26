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
const LocalServerID = 1

type ServiceErrorCode int

const (
	UnknownError       ServiceErrorCode = 10000 // 未知错误
	ParameterError     ServiceErrorCode = 10001 // 参数错误
	AuthError          ServiceErrorCode = 10002 // 认证错误
	SessionError       ServiceErrorCode = 10003 // Session 错误
	ServerConnectError ServiceErrorCode = 10004 // 服务器连接错误
	TargetNotFound     ServiceErrorCode = 10005 // 目标未找到
)
