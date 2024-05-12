package logger

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
)

type logger struct {
	*zap.SugaredLogger
	once    *sync.Once
	logFile string
}

var (
	log = &logger{once: &sync.Once{}, logFile: "logfile.log"}
)

func SetLogFile(logFilePath string) {
	log.logFile = logFilePath
}

func Logger() *logger {
	var err error
	log.once.Do(func() {
		cfg := zap.NewProductionConfig()
		cfg.OutputPaths = []string{log.logFile, "stdout"}

		var logger *zap.Logger
		logger, err = cfg.Build()
		if err != nil {
			return
		}

		log.SugaredLogger = logger.Sugar()
	})

	if err != nil {
		panic(fmt.Errorf("logger.Logger: %w", err))
	}

	return log
}

func (l *logger) Write(p []byte) (n int, err error) {
	l.Errorln(string(p))
	return len(p), nil
}
