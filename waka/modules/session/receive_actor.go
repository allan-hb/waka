package session

import (
	"net"

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
		"session_id": my.conn.ID(),
	})
	my.pid = context.Self()
	my.target = my.option.TargetCreator(my.conn.RawConn().(*net.TCPConn).RemoteAddr().String(), my.pid)

	if my.option.EnableHeart {
		my.startHeartbeat()
	}

	if my.option.EnableLog {
		log.WithFields(logrus.Fields{
			"pid": my.pid.String(),
		}).Debugln("session started")
	}
}
