package hall

import (
	"errors"

	"github.com/golang/protobuf/proto"
	"github.com/liuhan907/waka/waka-cow/database"
	"github.com/liuhan907/waka/waka-cow/proto"
	"github.com/liuhan907/waka/waka/modules/supervisor/supervisor_message"
	"github.com/sirupsen/logrus"
)

func (my *actorT) playerFutureRequestedRed(player *playerT, ev *supervisor_message.PlayerFutureRequested) bool {
	switch evd := ev.Payload.(type) {
	case *waka.RedGetRedPaperBagResultRequest:
		my.RedGetRedPaperBagResultRequest(player, evd, ev.Respond)
	case *waka.RedGetHistoryRequest:
		my.RedGetHistoryRequest(player, evd, ev.Respond)
	default:
		return false
	}
	return true
}

func (my *actorT) RedGetRedPaperBagResultRequest(player *playerT,
	ev *waka.RedGetRedPaperBagResultRequest,
	respond func(proto.Message, error)) {
	if player.InsideRed == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("get red result but not in")
		respond(nil, errors.New("not in bag"))
		return
	}

	bag, being := my.redBags[player.InsideRed]
	if !being {
		log.WithFields(logrus.Fields{
			"player": player.Player,
			"id":     player.InsideRed,
		}).Warnln("get red result but not found")
		player.InsideRed = 0
		respond(nil, errors.New("not found bag"))
		return
	}

	if int32(len(bag.Players)) != bag.Option.Number {
		respond(nil, errors.New("not settled"))
		return
	}

	redPlayer, being := bag.Players[player.Player]
	if !being {
		respond(nil, errors.New("not grabbed"))
		return
	}

	redPlayer.Lookup = true
	player.InsideRed = 0

	respond(&waka.RedGetRedPaperBagResultResponse{bag.RedRedPaperBag3()}, nil)

	my.sendRedUpdateRedPaperBagList(player.Player, my.redBags)
}

func (my *actorT) RedGetHistoryRequest(player *playerT,
	ev *waka.RedGetHistoryRequest,
	respond func(proto.Message, error)) {

	grabs, err := database.RedQueryGrabWarHistory(player.Player, 10)
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Warnln("query red war history failed")
		respond(nil, err)
		return
	}

	hands, err := database.RedQueryHandWarHistory(player.Player, 10)
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Warnln("query red war history failed")
		respond(nil, err)
		return
	}

	respond(&waka.RedGetHistoryResponse{grabs, hands}, nil)
}
