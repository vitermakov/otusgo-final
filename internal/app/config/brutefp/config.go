package brutefp

import (
	"fmt"
	"log"

	common "github.com/vitermakov/otusgo-final/internal/app/config"
	"github.com/vitermakov/otusgo-final/pkg/utils/jsonx"
)

const (
	defLimitsLoginPerMin    = 10
	defLimitsPasswordPerMin = 100
	defLimitsIPPerMin       = 1000
)

type Config struct {
	ServiceID   string         `json:"serviceId"`
	ServiceName string         `json:"serviceName"`
	Limits      Limits         `json:"limits"`
	Logger      common.Logger  `json:"logger"`
	API         common.Server  `json:"api"`
	Storage     common.Storage `json:"storage"`
}

type Limits struct {
	Method         string `json:"method"`
	Store          string `json:"store"`
	LoginPerMin    int    `json:"loginPerMin"`
	PasswordPerMin int    `json:"passwordPerMin"`
	IPPerMin       int    `json:"ipPerMin"`
	// BasePeriod настройка задаваемая только для тестов.
	BaseDuration jsonx.Duration `json:"baseDuration"`
}

func New(fileName string) (Config, error) {
	var cfg Config
	if err := common.New(fileName, &cfg); err != nil {
		return cfg, fmt.Errorf("error reading configuaration from '%s': %w", fileName, err)
	}
	if cfg.Limits.LoginPerMin <= 0 {
		log.Printf("wrong login per minute limit value, set default '%d'\n", defLimitsLoginPerMin)
		cfg.Limits.LoginPerMin = defLimitsLoginPerMin
	}
	if cfg.Limits.PasswordPerMin <= 0 {
		log.Printf("wrong password per minute limit value, set default '%d'\n", defLimitsPasswordPerMin)
		cfg.Limits.PasswordPerMin = defLimitsPasswordPerMin
	}
	if cfg.Limits.IPPerMin <= 0 {
		log.Printf("wrong ip per minute limit value, set default '%d'\n", defLimitsIPPerMin)
		cfg.Limits.IPPerMin = defLimitsIPPerMin
	}
	// for tests only
	baseDur, err := cfg.Limits.BaseDuration.AsDuration()
	if err != nil || baseDur.Nanoseconds() <= 0 {
		log.Printf("wrong base duration value, set default 1 minute\n")
		cfg.Limits.BaseDuration = jsonx.NewDuration(1, 'm')
	}
	return cfg, nil
}
