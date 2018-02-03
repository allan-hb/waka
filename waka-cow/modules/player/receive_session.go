package player

import (
	"errors"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/liuhan907/waka/waka-cow/proto"

	"github.com/liuhan907/waka/waka/modules/session/session_message"
	"github.com/liuhan907/waka/waka/modules/supervisor/supervisor_message"
)

func (my *actorT) ReceiveSession(context actor.Context) bool {
	switch ev := context.Message().(type) {
	case *session_message.Closed:
		my.closed(ev)
	case *session_message.Transport:
		my.transport(ev)
	case *session_message.FutureRequest:
		my.futureRequest(ev)
	default:
		return false
	}
	return true
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) closed(ev *session_message.Closed) {
	if my.player != 0 {
		my.hall.Tell(&supervisor_message.PlayerLeave{uint64(my.player)})
	}
}

func (my *actorT) transport(ev *session_message.Transport) {
	if my.player == 0 {
		switch evd := ev.Payload.(type) {
		case *waka.WechatLogin:
			my.wechatLogin(evd)
		case *waka.TokenLogin:
			my.tokenLogin(evd)
		}
	} else {
		if my.hall != nil {
			my.hall.Tell(&supervisor_message.PlayerTransport{uint64(my.player), ev.Payload})
		}
	}
}

func (my *actorT) futureRequest(ev *session_message.FutureRequest) {
	if my.player == 0 {
		ev.Respond(nil, errors.New("unauthorized"))
	} else {
		switch evd := ev.Payload.(type) {
		case *waka.SetPlayerExtRequest:
			my.setPlayerExt(evd, ev.Respond)
		case *waka.SetPlayerAgentRequest:
			my.setPlayerSupervisor(evd, ev.Respond)
		default:
			if my.hall != nil {
				my.hall.Tell(&supervisor_message.PlayerFutureRequest{uint64(my.player), ev.Payload, ev.Respond})
			}
		}
	}
}
