package model

type contextKey string

const (
	// CtxKeyCmd              = contextKey("command")
	// CtxKeyHttpServerRunner = contextKey("HttpServerController")
	CtxKeyBuildInfo = contextKey("BuildInfo")
	// CtxKeyServerConfig     = contextKey("ServerConfig")
	// CtxKeyTestConfig       = contextKey("TestConfig")
)

type BuildInfo interface {
	Version() string
	BuildTime() string
	AppName() string
	ModulePath() string
}
