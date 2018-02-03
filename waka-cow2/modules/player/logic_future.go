package player

import (
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka-cow2/database"
	"github.com/liuhan907/waka/waka-cow2/proto"
)

func (my *actorT) SetPlayerExtRequest(ev *cow_proto.SetPlayerExtRequest, respond func(proto.Message, error)) {
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
		respond(&cow_proto.SetPlayerExtResponse{}, nil)
	}
}

func (my *actorT) GetPlayerHeadRequest(ev *cow_proto.GetPlayerHeadRequest, respond func(proto.Message, error)) {
	respond(&cow_proto.GetPlayerHeadResponse{database.Player(ev.PlayerId).PlayerData().Head}, nil)
}
