package log

import (
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka/modules/logger"
)

func init() {
	loggerOption := logger.Option{
		Prefix: "cow",
	}
	viewer := logger.Spawn(loggerOption)

	logrus.AddHook(&logger.LogHook{
		Target: viewer,
	})
}
