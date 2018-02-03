package supervisor

import (
	"os"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/sirupsen/logrus"
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
	my.log = logrus.WithFields(logrus.Fields{
		"pid":    os.Getpid(),
		"module": "waka.supervisor",
		"target": my.name,
	})
	my.pid = context.Self()
	my.target = my.option.TargetCreator(my.pid)
}
