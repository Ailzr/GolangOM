package constant

type AuthMethod string

const (
	AuthMethodPassword AuthMethod = "password" // password authentication
	AuthMethodKey      AuthMethod = "key"      // key authentication
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

// local server ID
const LocalServerID = 1

type ServiceErrorCode int

const (
	UnknownError       ServiceErrorCode = 10000 // unknown error
	ParameterError     ServiceErrorCode = 10001 // parameter error
	AuthError          ServiceErrorCode = 10002 // authentication error
	SessionError       ServiceErrorCode = 10003 // session error
	ServerConnectError ServiceErrorCode = 10004 // server connection error
	TargetNotFound     ServiceErrorCode = 10005 // target not found
)
