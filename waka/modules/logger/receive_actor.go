package logger

import (
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (my *actorT) ReceiveActor(context actor.Context) bool {
	switch context.Message().(type) {
	case *actor.Started:
		my.started(context)
	default:
		return false
	}
	return true
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) started(context actor.Context) {
	my.name = time.Now().Format("2006-01-02 15.04.05.log")
	my.pid = context.Self()
	my.startClock()
}
