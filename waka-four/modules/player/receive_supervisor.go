package player

import (
	"github.com/AsynkronIT/protoactor-go/actor"

	"github.com/liuhan907/waka/waka/modules/session/session_message"
	"github.com/liuhan907/waka/waka/modules/supervisor/supervisor_message"
)

func (my *actorT) ReceiveSupervisor(context actor.Context) bool {
	switch evd := context.Message().(type) {
	case *supervisor_message.Close:
		my.close(evd)
	case *supervisor_message.SendFromSupervisor:
		my.sendFromSupervisor(evd)
	default:
		return false
	}
	return true
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) close(ev *supervisor_message.Close) {
	my.conn.Tell(&session_message.Close{})
}

func (my *actorT) sendFromSupervisor(ev *supervisor_message.SendFromSupervisor) {
	my.conn.Tell(&session_message.Send{ev.Payload})
}
