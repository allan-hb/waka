package player

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/liuhan907/waka/waka-four/database"
	"github.com/liuhan907/waka/waka-four/proto"
	"github.com/liuhan907/waka/waka/modules/session/session_message"
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
		"conn": my.conn.String(),
	})
	my.pid = context.Self()

	my.conn.Tell(&session_message.Send{&four_proto.Welcome{
		Customers: database.GetCustomerServices(),
		Exts:      database.GetExt(),
		Notices:   database.GetNotices(),
		Urls:      database.GetUrls(),
	}})
}
