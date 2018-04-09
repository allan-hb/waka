package player

import (
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka-cow2/database"
	"github.com/liuhan907/waka/waka-cow2/proto"
	"github.com/liuhan907/waka/waka/modules/session/session_message"
	"github.com/liuhan907/waka/waka/modules/supervisor/supervisor_message"
)

func (my *actorT) WechatLogin(ev *cow_proto.WechatLogin) {
	my.log.WithFields(logrus.Fields{
		"union_id": ev.GetWechatUid(),
		"nickname": ev.GetNickname(),
		"head":     ev.GetHead(),
	}).Debugln("wechat login")

	token := buildToken(ev.GetWechatUid())

	player, being, err := database.QueryPlayerByWechatUid(ev.GetWechatUid())
	if err != nil {
		my.log.WithFields(logrus.Fields{
			"union_id": ev.GetWechatUid(),
			"err":      err,
		}).Warnln("query player by union_id failed")
		my.conn.Tell(&session_message.Send{&cow_proto.LoginFailed{0}})
		return
	}

	if !being {
		player, err := database.RegisterPlayer(ev.GetWechatUid(), ev.GetNickname(), ev.GetHead(), token)
		if err != nil {
			my.log.WithFields(logrus.Fields{
				"union_id": ev.GetWechatUid(),
				"nickname": ev.GetNickname(),
				"head":     ev.GetHead(),
				"token":    token,
				"err":      err,
			}).Warnln("register player failed")
			my.conn.Tell(&session_message.Send{&cow_proto.LoginFailed{0}})
			return
		}

		my.player = player.Id
	} else {
		if player.Ban != 0 {
			my.conn.Tell(&session_message.Send{&cow_proto.LoginFailed{0}})
			return
		}

		err = database.UpdatePlayerLogin(player.Id, ev.GetNickname(), ev.GetHead(), token)
		if err != nil {
			my.log.WithFields(logrus.Fields{
				"union_id": ev.GetWechatUid(),
				"nickname": ev.GetNickname(),
				"head":     ev.GetHead(),
				"token":    token,
				"err":      err,
			}).Warnln("update player failed")
			my.conn.Tell(&session_message.Send{&cow_proto.LoginFailed{0}})
			return
		}

		my.player = player.Id
	}

	my.hall.Tell(&supervisor_message.PlayerEnter{my.pid, uint64(my.player), my.remote})
	my.conn.Tell(&session_message.Send{&cow_proto.LoginSuccess{token}})
	my.log.WithFields(logrus.Fields{
		"union_id": ev.GetWechatUid(),
		"nickname": ev.GetNickname(),
		"head":     ev.GetHead(),
		"token":    token,
	}).Debugln("wechat login success")
}

func (my *actorT) TokenLogin(ev *cow_proto.TokenLogin) {
	my.log.WithFields(logrus.Fields{
		"token": ev.GetToken(),
	}).Debugln("token login")

	player, being, err := database.QueryPlayerByToken(ev.GetToken())
	if err != nil {
		my.log.WithFields(logrus.Fields{
			"token": ev.GetToken(),
			"err":   err,
		}).Warnln("query player by token failed")
		my.conn.Tell(&session_message.Send{&cow_proto.LoginFailed{0}})
		return
	}
	if !being {
		my.log.WithFields(logrus.Fields{
			"token": ev.GetToken(),
		}).Debugln("token not found")
		my.conn.Tell(&session_message.Send{&cow_proto.LoginFailed{1}})
	} else {
		if player.Ban != 0 {
			my.conn.Tell(&session_message.Send{&cow_proto.LoginFailed{0}})
			return
		}

		my.player = player.Id
		my.hall.Tell(&supervisor_message.PlayerEnter{my.pid, uint64(my.player), my.remote})
		my.conn.Tell(&session_message.Send{&cow_proto.LoginSuccess{ev.GetToken()}})
		my.log.WithFields(logrus.Fields{
			"union_id": player.UnionId,
			"nickname": player.Nickname,
			"head":     player.Head,
			"player":   player.Id,
			"token":    ev.GetToken(),
		}).Debugln("token login success")
	}
}
