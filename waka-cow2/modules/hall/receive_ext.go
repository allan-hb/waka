package hall

import (
	"github.com/AsynkronIT/protoactor-go/actor"

	"github.com/liuhan907/waka/waka-cow2/modules/hall/hall_message"
)

func (my *actorT) ReceiveExt(context actor.Context) bool {
	switch ev := context.Message().(type) {
	case *hall_message.UpdatePlayerSecret:
		my.UpdatePlayerSecret(ev)
	default:
		return false
	}
	return true
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) UpdatePlayerSecret(ev *hall_message.UpdatePlayerSecret) {
	my.sendPlayerSecret(ev.Player)
}
