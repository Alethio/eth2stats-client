package commands

import (
	"github.com/gin-gonic/gin"
	formatter "github.com/kwix/logrus-module-formatter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func initLogging() {
	logging := viper.GetString("logging")

	if verbose {
		logging = "*=debug"
	}

	if vverbose {
		logging = "*=trace"
	}

	if logging == "" {
		logging = "*=info"
	}

	gin.SetMode(gin.DebugMode)

	modules := formatter.NewModulesMap(logging)
	if level, exists := modules["gin"]; exists {
		if level < logrus.DebugLevel {
			gin.SetMode(gin.ReleaseMode)
		}
	} else {
		level := modules["*"]
		if level < logrus.DebugLevel {
			gin.SetMode(gin.ReleaseMode)
		}
	}
	logrusFormatter := &logrus.TextFormatter{
		FullTimestamp: fullTimestamps,
	}
	f, err := formatter.NewWithFormatter(modules, logrusFormatter)
	if err != nil {
		panic(err)
	}

	logrus.SetFormatter(f)

	log.Debug("Debug mode")
}
