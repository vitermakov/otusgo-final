package config

import (
	"encoding/json"
	"os"
)

type Logger struct {
	FileName string `json:"fileName"`
	Level    string `json:"level"`
}

type Server struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type API struct {
	Type    string `json:"type"`
	Address string `json:"address"`
}

type Storage struct {
	Type   string  `json:"type"`
	PGConn SQLConn `json:"pgsql"`
}

type Queue struct {
	Type     string `json:"type"`
	RabbitMQ Conn   `json:"rabbitMq"`
}

type Conn struct {
	Host              string `json:"host"`
	Port              int    `json:"port"`
	User              string `json:"user"`
	Password          string `json:"password"`
	ConnAttemptsCount int    `json:"connAttemptsCount"`
	ConnAttemptsWait  int    `json:"connAttemptsWait"` // сек.
}

type SQLConn struct {
	Conn
	DBName          string `json:"dbName"`
	ConnMaxLifetime int    `json:"maxLifetime"`     // сек.
	ConnMaxIdleTime int    `json:"connMaxIdleTime"` // сек.
	MaxOpenCons     int    `json:"maxOpenCons"`
	MaxIdleCons     int    `json:"maxIdleCons"`
}

type Mailer struct {
	Type         string `json:"type"`
	DefaultFrom  string `json:"defaultFrom"`
	TemplatePath string `json:"templatePath"`
}

// New используем обычный encode/json.
func New(fileName string, config interface{}) error {
	bs, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, config)
}
