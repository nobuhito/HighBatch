package highbatch

import (
	// "github.com/mattn/go-colorable"
	"github.com/Sirupsen/logrus"
	"os"
)

var logger logrus.Logger

func LogInit(debugLevel string) {

	if _, err := os.Stat("log"); err != nil {
		if err := os.Mkdir("log", 0666); err != nil {
			Le(err)
			Lw("can't create log directory")
		}
	}

	f, err := os.OpenFile("log/HighBatch.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		Le(err)
		Lw("can't create log file")
	}

	var level = logrus.WarnLevel
	switch debugLevel{
	case "debug":
		level = logrus.DebugLevel
	case "info":
		level = logrus.InfoLevel
	case "error":
		level = logrus.ErrorLevel
	}

	logger = logrus.Logger{
		Formatter: &logrus.TextFormatter{DisableColors: true},
		Level: level,
		Out: f,
	}
}


func Ld(msg string) {
	logger.Debug(msg)
}


func Li(msg string) {
	logger.Info(msg)
}

func Lw(msg string) {
	logger.Warn(msg)
}

func Le(err error) {
	logger.Error(err)
}
