package logger

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/liuhan907/waka/waka/modules/logger/logger_message"
	"github.com/sirupsen/logrus"
)

type LogHook struct {
	Target *actor.PID

	logrus.Hook
}

func (hook *LogHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}

func (hook *LogHook) Fire(entry *logrus.Entry) error {
	formatter := logrus.JSONFormatter{}
	d, err := formatter.Format(entry)
	if err != nil {
		return err
	}

	hook.Target.Tell(&logger_message.LogData{d})

	return nil
}
