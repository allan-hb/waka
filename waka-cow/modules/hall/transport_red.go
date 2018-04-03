package hall

import (
	"sync/atomic"

	waka "github.com/liuhan907/waka/waka-cow/proto"
	"github.com/liuhan907/waka/waka/modules/supervisor/supervisor_message"
	"github.com/sirupsen/logrus"
)

func (my *actorT) playerTransportedRed(player *playerT, ev *supervisor_message.PlayerTransported) bool {
	switch evd := ev.Payload.(type) {
	case *waka.RedCreateBag:
		my.RedCreateBag(player, evd)
	case *waka.RedGrab:
		my.RedGrab(player, evd)
	case *waka.RedLeave:
		my.RedLeave(player, evd)
	default:
		return false
	}
	return true
}

func (my *actorT) RedCreateBag(player *playerT, ev *waka.RedCreateBag) {
	ev.GetOption().Money *= 100

	if ev.GetOption().GetNumber() != 7 && ev.GetOption().GetNumber() != 10 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
			"option": ev.GetOption().String(),
		}).Warnln("create red but option illegal")
		my.sendRedCreateBagFailed(player.Player, 0)
		return
	}

	if len(ev.GetOption().GetMantissa()) != 1 && len(ev.GetOption().GetMantissa()) != 2 && len(ev.GetOption().GetMantissa()) != 3 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
			"option": ev.GetOption().String(),
		}).Warnln("create red but option illegal")
		my.sendRedCreateBagFailed(player.Player, 0)
		return
	}

	if len(ev.GetOption().GetMantissa()) > 1 && ev.GetOption().GetNumber() != 7 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
			"option": ev.GetOption().String(),
		}).Warnln("create red but option illegal")
		my.sendRedCreateBagFailed(player.Player, 0)
		return
	}

	id := atomic.AddInt32(&my.redIdPool, 1)

	bag := new(redBagT)
	bag.Create(my, id, ev.GetOption(), player.Player)
}

func (my *actorT) RedGrab(player *playerT, ev *waka.RedGrab) {
	bag, being := my.redBags[ev.GetId()]
	if !being {
		log.WithFields(logrus.Fields{
			"player": player.Player,
			"id":     ev.GetId(),
		}).Warnln("grab red but not found")
		my.sendRedGrabFailed(player.Player, 1)
		return
	}

	bag.Grab(player)
}

func (my *actorT) RedLeave(player *playerT, ev *waka.RedLeave) {
	player.InsideRed = 0
}
