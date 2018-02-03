package hall

import (
	"github.com/liuhan907/waka/waka-cow/database"
	"github.com/liuhan907/waka/waka-cow/modules/hall/tools/cow"
	"github.com/liuhan907/waka/waka-cow/proto"
	"github.com/liuhan907/waka/waka/modules/supervisor/supervisor_message"
	"github.com/sirupsen/logrus"
)

func (my *actorT) playerTransportedCow(player *playerT, ev *supervisor_message.PlayerTransported) bool {
	switch evd := ev.Payload.(type) {
	case *waka.NiuniuCreateRoom:
		my.NiuniuCreateRoom(player, evd)
	case *waka.NiuniuJoinRoom:
		my.NiuniuJoinRoom(player, evd)
	case *waka.NiuniuLeaveRoom:
		my.NiuniuLeaveRoom(player, evd)
	case *waka.NiuniuSwitchReady:
		my.NiuniuSwitchReady(player, evd)
	case *waka.NiuniuSwitchRole:
		my.NiuniuSwitchRole(player, evd)
	case *waka.NiuniuDismiss:
		my.NiuniuDismiss(player, evd)
	case *waka.NiuniuStart:
		my.NiuniuStart(player, evd)
	case *waka.NiuniuSpecifyBanker:
		my.NiuniuSpecifyBanker(player, evd)
	case *waka.NiuniuGrab:
		my.NiuniuGrab(player, evd)
	case *waka.NiuniuSpecifyRate:
		my.NiuniuSpecifyRate(player, evd)
	case *waka.NiuniuCommitPokers:
		my.NiuniuCommitPokers(player, evd)
	case *waka.NiuniuContinueWith:
		my.NiuniuContinueWith(player, evd)
	default:
		return false
	}
	return true
}

func (my *actorT) NiuniuCreateRoom(player *playerT, ev *waka.NiuniuCreateRoom) {
	if player.InsideCow != 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("create cow room but already in room")
		my.sendNiuniuCreateRoomFailed(player.Player, 2)
		return
	}

	if (ev.GetType() != waka.NiuniuRoomType_Order && ev.GetType() != waka.NiuniuRoomType_PayForAnother) ||
		(ev.Option.GetBanker() < 0 || ev.Option.GetBanker() > 2) ||
		(ev.Option.GetGames() != 20 && ev.Option.GetGames() != 30 && ev.Option.GetGames() != 40 && ev.Option.GetGames() != 5) ||
		(ev.Option.GetMode() != 0 && ev.Option.GetMode() != 1) {
		log.WithFields(logrus.Fields{
			"player": player.Player,
			"type":   ev.GetType().String(),
			"option": ev.GetOption().String(),
		}).Warnln("create cow room but has illegal option")
		my.sendNiuniuCreateRoomFailed(player.Player, 3)
		return
	}

	id, ok := my.cowPlayerNumberPool.Acquire()
	if !ok {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("create cow room but acquire room id failed")
		my.sendNiuniuCreateRoomFailed(player.Player, 0)
		return
	}

	var room cowRoom

	switch ev.GetType() {
	case waka.NiuniuRoomType_Order:
		room = new(orderRoomT)
	case waka.NiuniuRoomType_PayForAnother:
		room = new(payForAnotherRoomT)
	default:
		panic("this code should not be executed")
	}

	room.CreateRoom(my, id, ev.GetOption(), player.Player)
}

func (my *actorT) NiuniuJoinRoom(player *playerT, ev *waka.NiuniuJoinRoom) {
	if player.InsideCow != 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("join cow room but already in room")
		my.sendNiuniuJoinRoomFailed(player.Player, 2)
		return
	}

	room, being := my.cowRooms[ev.GetRoomId()]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"room_id": ev.GetRoomId(),
		}).Warnln("join cow room but not found")
		my.sendNiuniuJoinRoomFailed(player.Player, 3)
		return
	}

	room.JoinRoom(player)
}

func (my *actorT) NiuniuLeaveRoom(player *playerT, ev *waka.NiuniuLeaveRoom) {
	if player.InsideCow == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("leave cow room but not had")
		return
	}

	room, being := my.cowRooms[player.InsideCow]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"room_id": player.InsideCow,
		}).Warnln("leave cow room but not found")
		player.InsideCow = 0
		return
	}

	room.LeaveRoom(player)
}

func (my *actorT) NiuniuSwitchReady(player *playerT, ev *waka.NiuniuSwitchReady) {
	if player.InsideCow == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("switch ready but not in room")
		return
	}

	room, being := my.cowRooms[player.InsideCow]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"room_id": player.InsideCow,
		}).Warnln("switch ready but not room not found")
		player.InsideCow = 0
		return
	}

	room.SwitchReady(player)
}

func (my *actorT) NiuniuSwitchRole(player *playerT, ev *waka.NiuniuSwitchRole) {
	if player.InsideCow == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("switch role but not in room")
		return
	}

	room, being := my.cowRooms[player.InsideCow]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"room_id": player.InsideCow,
		}).Warnln("switch role but not room not found")
		player.InsideCow = 0
		return
	}

	room.SwitchRole(player)
}

func (my *actorT) NiuniuDismiss(player *playerT, ev *waka.NiuniuDismiss) {
	roomId := int32(0)

	if ev.GetRoomId() != 0 {
		roomId = ev.GetRoomId()
	} else if player.InsideCow != 0 {
		roomId = player.InsideCow
	}

	if roomId == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("dismiss room but not found")
		return
	}

	room, being := my.cowRooms[roomId]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"room_id": player.InsideCow,
		}).Warnln("dismiss room but not found")
		return
	}

	room.Dismiss(player)
}

func (my *actorT) NiuniuStart(player *playerT, ev *waka.NiuniuStart) {
	if player.InsideCow == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("start but not in room")
		return
	}

	room, being := my.cowRooms[player.InsideCow]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"room_id": player.InsideCow,
		}).Warnln("start but room not found")
		player.InsideCow = 0
		return
	}

	room.Start(player)
}

func (my *actorT) NiuniuSpecifyBanker(player *playerT, ev *waka.NiuniuSpecifyBanker) {
	if player.InsideCow == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("specify banker but not in room")
		return
	}

	room, being := my.cowRooms[player.InsideCow]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"room_id": player.InsideCow,
		}).Warnln("specify banker but room not found")
		player.InsideCow = 0
		return
	}

	room.SpecifyBanker(player, database.Player(ev.GetBanker()))
}

func (my *actorT) NiuniuGrab(player *playerT, ev *waka.NiuniuGrab) {
	if player.InsideCow == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("grab but not in room")
		return
	}

	room, being := my.cowRooms[player.InsideCow]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"room_id": player.InsideCow,
		}).Warnln("grab but room not found")
		player.InsideCow = 0
		return
	}

	room.Grab(player, ev.GetDoing())
}

func (my *actorT) NiuniuSpecifyRate(player *playerT, ev *waka.NiuniuSpecifyRate) {
	if player.InsideCow == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("specify rate but not in room")
		return
	}

	if ev.GetRate() < 1 || ev.GetRate() > 3 {
		return
	}

	room, being := my.cowRooms[player.InsideCow]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"room_id": player.InsideCow,
		}).Warnln("specify rate but room not found")
		player.InsideCow = 0
		return
	}

	room.SpecifyRate(player, ev.GetRate())
}

func (my *actorT) NiuniuCommitPokers(player *playerT, ev *waka.NiuniuCommitPokers) {
	if player.InsideCow == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("commit pokers but not in room")
		return
	}

	room, being := my.cowRooms[player.InsideCow]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"room_id": player.InsideCow,
		}).Warnln("commit pokers but room not found")
		player.InsideCow = 0
		return
	}

	_, _, _, err := cow.GetPokersPattern(ev.GetPokers(), room.GetOption().GetMode())
	if err != nil {
		return
	}

	room.CommitPokers(player, ev.GetPokers())
}

func (my *actorT) NiuniuContinueWith(player *playerT, ev *waka.NiuniuContinueWith) {
	if player.InsideCow == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("continue but not in room")
		return
	}

	room, being := my.cowRooms[player.InsideCow]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"room_id": player.InsideCow,
		}).Warnln("continue but room not found")
		player.InsideCow = 0
		return
	}

	room.ContinueWith(player)
}
