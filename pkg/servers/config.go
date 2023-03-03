package servers

const (
	defaultHost = "localhost"
	defaultPort = 8080
)

type CtxKey struct{}

type Config struct {
	host  string
	port  int
	debug bool
}

func (cfg Config) GetHost() string {
	if len(cfg.host) > 0 {
		return cfg.host
	}
	return defaultHost
}

func (cfg Config) GetPort() int {
	if cfg.port > 0 {
		return cfg.port
	}
	return defaultPort
}

func (cfg Config) IsDebug() bool {
	return cfg.debug
}

func NewConfig(host string, port int, debug bool) Config {
	return Config{
		host:  host,
		port:  port,
		debug: debug,
	}
}
