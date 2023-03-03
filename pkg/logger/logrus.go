package logger

import (
	"io"
	"log"
	"os"

	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

type Logrus struct {
	logger *logrus.Logger
}

func NewLogrus(cfg Config) (Logger, error) {
	inst := logrus.New()
	if !cfg.Level.Valid() {
		return nil, ErrorUnknownLevel
	}
	var level logrus.Level
	switch cfg.Level {
	case LevelNone, LevelOut:
		level = logrus.PanicLevel
	case LevelInfo:
		level = logrus.InfoLevel
	case LevelWarn:
		level = logrus.WarnLevel
	case LevelError:
		level = logrus.ErrorLevel
	case LevelFatal:
		level = logrus.FatalLevel
	case LevelDebug:
		level = logrus.DebugLevel
	}
	if len(cfg.FileName) > 0 {
		f, err := os.OpenFile(cfg.FileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
		if err != nil {
			return nil, ErrorOpenLogFile
		}
		if !cfg.IsTesting {
			inst.SetOutput(io.MultiWriter(f, os.Stdout))
		} else {
			inst.SetOutput(f)
		}
	}
	inst.SetLevel(level)
	inst.SetFormatter(&easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       "[%lvl%]: %time% - %msg%\n",
	})
	if cfg.IsTesting {
		inst.ExitFunc = func(i int) {
			log.Printf("logrus fatal exit(%d)\n", i)
		}
	}
	return &Logrus{logger: inst}, nil
}

func (l Logrus) Debug(msg string, args ...interface{}) {
	l.logger.Debugf(msg, args...)
}

func (l Logrus) Info(msg string, args ...interface{}) {
	l.logger.Infof(msg, args...)
}

func (l Logrus) Warn(msg string, args ...interface{}) {
	l.logger.Warnf(msg, args...)
}

func (l Logrus) Error(msg string, args ...interface{}) {
	l.logger.Errorf(msg, args...)
}

func (l Logrus) Fatal(msg string, args ...interface{}) {
	l.logger.Fatalf(msg, args...)
}
