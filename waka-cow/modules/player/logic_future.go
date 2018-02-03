package player

import (
	"errors"

	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka-cow/database"
	"github.com/liuhan907/waka/waka-cow/proto"
)

func (my *actorT) setPlayerExt(ev *waka.SetPlayerExtRequest, respond func(proto.Message, error)) {
	my.log.WithFields(logrus.Fields{
		"player": my.player,
		"name":   ev.GetName(),
		"idcard": ev.GetIdcard(),
		"wechat": ev.GetWechat(),
	}).Debugln("set player ext")

	err := database.UpdatePlayerExt(my.player, ev.GetName(), ev.GetIdcard(), ev.GetWechat())
	if err != nil {
		my.log.WithFields(logrus.Fields{
			"player": my.player,
			"err":    err,
		}).Warnln("set player ext failed")

		respond(nil, err)
	} else {
		respond(&waka.SetPlayerExtResponse{}, nil)
	}
}

func (my *actorT) setPlayerSupervisor(ev *waka.SetPlayerAgentRequest, respond func(proto.Message, error)) {
	my.log.WithFields(logrus.Fields{
		"player":     my.player,
		"supervisor": ev.GetPlayerId(),
	}).Debugln("set player supervisor")

	_, being, err := database.QueryPlayerByRef(database.Player(ev.GetPlayerId()))
	if err != nil {
		my.log.WithFields(logrus.Fields{
			"player":     my.player,
			"supervisor": ev.GetPlayerId(),
			"err":        err,
		}).Warnln("set player supervisor failed")

		respond(nil, err)
	} else {
		if !being {
			my.log.WithFields(logrus.Fields{
				"player":     my.player,
				"supervisor": ev.GetPlayerId(),
			}).Warnln("set player supervisor but supervisor not found")

			respond(nil, errors.New("supervisor not found"))
		} else {
			err := database.UpdatePlayerSupervisor(my.player, database.Player(ev.GetPlayerId()))
			if err != nil {
				my.log.WithFields(logrus.Fields{
					"player":     my.player,
					"supervisor": ev.GetPlayerId(),
					"err":        err,
				}).Warnln("set player supervisor failed")

				respond(nil, err)
			} else {
				respond(&waka.SetPlayerAgentResponse{}, nil)
			}
		}
	}
}
