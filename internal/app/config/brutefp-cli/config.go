package brutefp

import (
	"fmt"

	common "github.com/vitermakov/otusgo-final/internal/app/config"
)

type Config struct {
	ServiceID   string        `json:"serviceId"`
	ServiceName string        `json:"serviceName"`
	Logger      common.Logger `json:"logger"`
	GrpcClient  common.Server `json:"grpcClient"`
}

func New(fileName string) (Config, error) {
	var cfg Config
	if err := common.New(fileName, &cfg); err != nil {
		return cfg, fmt.Errorf("error reading configuaration from '%s': %w", fileName, err)
	}
	return cfg, nil
}
