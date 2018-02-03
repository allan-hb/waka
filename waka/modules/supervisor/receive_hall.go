package supervisor

import (
	"reflect"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka/modules/supervisor/supervisor_message"
)

func (my *actorT) ReceiveHall(context actor.Context) bool {
	switch ev := context.Message().(type) {
	case *supervisor_message.SendFromHall:
		my.send(ev)
	default:
		return false
	}
	return true
}

func (my *actorT) send(ev *supervisor_message.SendFromHall) {
	player, being := my.players[ev.Player]
	if !being {
		if my.option.EnableLog {
			my.log.WithFields(logrus.Fields{
				"player":  ev.Player,
				"type":    reflect.TypeOf(ev.Payload).Elem().Name(),
				"payload": ev.Payload.String(),
			}).Warnln("redirect transport from hall to player but player not found")
		}
		return
	}

	if my.option.EnableLog {
		my.log.WithFields(logrus.Fields{
			"player":  ev.Player,
			"type":    reflect.TypeOf(ev.Payload).Elem().Name(),
			"payload": ev.Payload.String(),
		}).Debugln("redirect transport from hall to player")
	}

	player.Tell(&supervisor_message.SendFromSupervisor{ev.Payload})
}
