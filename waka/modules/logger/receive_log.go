package logger

import (
	"log"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/liuhan907/waka/waka/modules/logger/logger_message"
)

func (my *actorT) ReceiveLog(context actor.Context) bool {
	switch evd := context.Message().(type) {
	case *logger_message.LogData:
		my.logData(evd)
	default:
		return false
	}
	return true
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) logData(ev *logger_message.LogData) {
	_, err := my.w.Write(ev.Payload)
	if err != nil {
		log.Println("write log cache failed: ", err)
	}
}
