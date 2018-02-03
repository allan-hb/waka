package hall

import (
	"errors"

	"github.com/golang/protobuf/proto"
	"github.com/liuhan907/waka/waka-cow/database"
	"github.com/liuhan907/waka/waka-cow/proto"
	"github.com/liuhan907/waka/waka/modules/supervisor/supervisor_message"
	"github.com/sirupsen/logrus"
)

func (my *actorT) playerFutureRequestedLever28(player *playerT, ev *supervisor_message.PlayerFutureRequested) bool {
	switch evd := ev.Payload.(type) {
	case *waka.Lever28GetRedPaperBagResultRequest:
		my.Lever28GetRedPaperBagResultRequest(player, evd, ev.Respond)
	case *waka.Lever28GetHistoryRequest:
		my.Lever28GetHistoryRequest(player, evd, ev.Respond)
	default:
		return false
	}
	return true
}

func (my *actorT) Lever28GetRedPaperBagResultRequest(player *playerT,
	ev *waka.Lever28GetRedPaperBagResultRequest,
	respond func(proto.Message, error)) {

	if player.InsideLever28 == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("get lever28 result but not in")
		respond(nil, errors.New("not in bag"))
		return
	}

	bag, being := my.lever28Bags[player.InsideLever28]
	if !being {
		log.WithFields(logrus.Fields{
			"player": player.Player,
			"id":     player.InsideLever28,
		}).Warnln("get lever28 result but not found")
		player.InsideLever28 = 0
		respond(nil, errors.New("not found bag"))
		return
	}

	if int32(len(bag.Players)) != 4 {
		respond(nil, errors.New("not settled"))
		return
	}

	lever28Player, being := bag.Players[player.Player]
	if !being {
		respond(nil, errors.New("not grabbed"))
		return
	}

	lever28Player.Lookup = true
	player.InsideLever28 = 0

	respond(&waka.Lever28GetRedPaperBagResultResponse{bag.Lever28RedPaperBag3()}, nil)

	my.sendLever28UpdateRedPaperBagList(player.Player, my.lever28Bags)

}

func (my *actorT) Lever28GetHistoryRequest(player *playerT,
	ev *waka.Lever28GetHistoryRequest,
	respond func(proto.Message, error)) {

	grabs, err := database.Lever28QueryGrabWarHistory(player.Player, 10)
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Warnln("query red war history failed")
		respond(nil, err)
		return
	}

	hands, err := database.Lever28QueryHandWarHistory(player.Player, 10)
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Warnln("query red war history failed")
		respond(nil, err)
		return
	}

	respond(&waka.Lever28GetHistoryResponse{grabs, hands}, nil)
}
