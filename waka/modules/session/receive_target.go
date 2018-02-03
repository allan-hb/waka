package session

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka/codec"
	"github.com/liuhan907/waka/waka/modules/session/session_message"
	"github.com/liuhan907/waka/waka/proto"
)

func (my *actorT) ReceiveTarget(context actor.Context) bool {
	switch ev := context.Message().(type) {
	case *session_message.Close:
		my.close()
	case *session_message.Send:
		my.send(ev)
	default:
		return false
	}
	return true
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) close() {
	if my.option.EnableLog {
		log.WithFields(logrus.Fields{
			"pid": my.pid.String(),
		}).Debugln("target request close session")
	}

	my.conn.Close()
}

func (my *actorT) send(ev *session_message.Send) {
	d, id, name, err := codec.Encode(ev.Payload)
	if err != nil {
		if my.option.EnableLog {
			log.WithFields(logrus.Fields{
				"pid":     my.pid.String(),
				"payload": ev.Payload.String(),
				"err":     err,
			}).Warnln("transport encode failed")
		}
	} else {
		if my.option.EnableLog {
			log.WithFields(logrus.Fields{
				"pid":     my.pid.String(),
				"id":      id,
				"name":    name,
				"payload": ev.Payload.String(),
			}).Debugln("redirect transport from target to gateway")
		}

		my.conn.Send(&waka_proto.Transport{
			Id:      id,
			Payload: d,
		})
	}
}
