package supervisor

import (
	"reflect"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka/modules/supervisor/supervisor_message"
)

func (my *actorT) ReceivePlayer(context actor.Context) bool {
	switch ev := context.Message().(type) {
	case *supervisor_message.PlayerEnter:
		my.playerEnter(context, ev)
	case *supervisor_message.PlayerLeave:
		my.playerLeave(context, ev)
	case *supervisor_message.PlayerTransport:
		my.playerTransport(context, ev)
	case *supervisor_message.PlayerFutureRequest:
		my.futureRequest(context, ev)
	default:
		return false
	}
	return true
}

func (my *actorT) playerEnter(context actor.Context, ev *supervisor_message.PlayerEnter) {
	if my.option.EnableLog {
		my.log.WithFields(logrus.Fields{
			"player": ev.Player,
		}).Debugln("player enter")
	}

	exchanged := false
	if player, being := my.players[ev.Player]; being {
		if my.option.EnableLog {
			my.log.WithFields(logrus.Fields{
				"player": ev.Player,
			}).Debugln("player exchange")
		}

		player.Tell(&supervisor_message.Close{})

		exchanged = true
	}

	my.players[ev.Player] = ev.Conn

	if !exchanged {
		my.target.Tell(&supervisor_message.PlayerEntered{ev.Player, ev.Remote})
	} else {
		my.target.Tell(&supervisor_message.PlayerExchanged{ev.Player, ev.Remote})
	}
}

func (my *actorT) playerLeave(context actor.Context, ev *supervisor_message.PlayerLeave) {
	if my.option.EnableLog {
		my.log.WithFields(logrus.Fields{
			"player": ev.Player,
		}).Debugln("player leave")
	}

	player, being := my.players[ev.Player]
	if !being {
		if my.option.EnableLog {
			my.log.WithFields(logrus.Fields{
				"player": ev.Player,
			}).Warnln("player leave but not found")
		}
		return
	}

	player.Tell(&supervisor_message.Close{})

	delete(my.players, ev.Player)

	my.target.Tell(&supervisor_message.PlayerLeft{ev.Player})
}

func (my *actorT) playerTransport(context actor.Context, ev *supervisor_message.PlayerTransport) {
	_, being := my.players[ev.Player]
	if !being {
		if my.option.EnableLog {
			my.log.WithFields(logrus.Fields{
				"player":  ev.Player,
				"type":    reflect.TypeOf(ev.Payload).Elem().Name(),
				"payload": ev.Payload.String(),
			}).Warnln("redirect transport from player to hall but player not found")
		}
		return
	}

	if my.option.EnableLog {
		my.log.WithFields(logrus.Fields{
			"player":  ev.Player,
			"type":    reflect.TypeOf(ev.Payload).Elem().Name(),
			"payload": ev.Payload.String(),
		}).Debugln("redirect transport from player to hall")
	}

	my.target.Tell(&supervisor_message.PlayerTransported{ev.Player, ev.Payload})
}

func (my *actorT) futureRequest(context actor.Context, ev *supervisor_message.PlayerFutureRequest) {
	_, being := my.players[ev.Player]
	if !being {
		if my.option.EnableLog {
			my.log.WithFields(logrus.Fields{
				"player":  ev.Player,
				"type":    reflect.TypeOf(ev.Payload).Elem().Name(),
				"payload": ev.Payload.String(),
			}).Warnln("redirect future request from player to hall but player not found")
		}
		return
	}

	if my.option.EnableLog {
		my.log.WithFields(logrus.Fields{
			"player":  ev.Player,
			"type":    reflect.TypeOf(ev.Payload).Elem().Name(),
			"payload": ev.Payload.String(),
		}).Debugln("redirect future request from player to hall")
	}

	my.target.Tell(&supervisor_message.PlayerFutureRequested{ev.Player, ev.Payload, ev.Respond})
}
