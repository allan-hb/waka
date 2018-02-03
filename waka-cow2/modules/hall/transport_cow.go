package hall

import (
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka-cow2/database"
	"github.com/liuhan907/waka/waka-cow2/proto"
	"github.com/liuhan907/waka/waka/modules/supervisor/supervisor_message"
)

func (my *actorT) playerTransportedCow(player *playerT, ev *supervisor_message.PlayerTransported) bool {
	switch evd := ev.Payload.(type) {
	case *cow_proto.NiuniuCreateRoom:
		my.NiuniuCreateRoom(player, evd)
	case *cow_proto.NiuniuJoinRoom:
		my.NiuniuJoinRoom(player, evd)
	case *cow_proto.NiuniuLeaveRoom:
		my.NiuniuLeaveRoom(player, evd)
	case *cow_proto.NiuniuSwitchReady:
		my.NiuniuSwitchReady(player, evd)
	case *cow_proto.NiuniuDismiss:
		my.NiuniuDismiss(player, evd)
	case *cow_proto.NiuniuKickPlayer:
		my.NiuniuKickPlayer(player, evd)
	case *cow_proto.NiuniuStart:
		my.NiuniuStart(player, evd)
	case *cow_proto.NiuniuSpecifyBanker:
		my.NiuniuSpecifyBanker(player, evd)
	case *cow_proto.NiuniuGrab:
		my.NiuniuGrab(player, evd)
	case *cow_proto.NiuniuSpecifyRate:
		my.NiuniuSpecifyRate(player, evd)
	case *cow_proto.NiuniuContinueWith:
		my.NiuniuContinueWith(player, evd)
	default:
		return false
	}
	return true
}

func (my *actorT) NiuniuCreateRoom(player *playerT, ev *cow_proto.NiuniuCreateRoom) {
	if player.InsideCow != 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("create cow room but already in room")
		my.sendNiuniuCreateRoomFailed(player.Player, 2)
		return
	}

	if (ev.GetOption().GetBankerMode() != 0 && ev.GetOption().GetBankerMode() != 1 && ev.GetOption().GetBankerMode() != 2) ||
		(ev.GetOption().GetScore() != 1 && ev.GetOption().GetScore() != 2 && ev.GetOption().GetScore() != 3 && ev.GetOption().GetScore() != 5 &&
			ev.GetOption().GetScore() != 10 && ev.GetOption().GetScore() != 20 && ev.GetOption().GetScore() != 30 && ev.GetOption().GetScore() != 50) ||
		(ev.Option.GetRoundNumber() != 12 || ev.Option.GetRoundNumber() != 20) ||
		(ev.Option.GetPayMode() != 1 && ev.Option.GetPayMode() != 2) ||
		(ev.Option.GetMode() != 0 && ev.Option.GetMode() != 1) ||
		(ev.Option.GetAdditionalPokers() != 0 && ev.Option.GetAdditionalPokers() != 1) {
		log.WithFields(logrus.Fields{
			"player": player.Player,
			"option": ev.GetOption().String(),
		}).Warnln("create cow room but has illegal option")
		my.sendNiuniuCreateRoomFailed(player.Player, 3)
		return
	}

	id, ok := my.cowNumberPool.Acquire()
	if !ok {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("create cow room but acquire room id failed")
		my.sendNiuniuCreateRoomFailed(player.Player, 0)
		return
	}

	var room cowRoomT

	switch ev.GetOption().GetPayMode() {
	case 1:
		room = new(payForAnotherRoomT)
	case 2:
		room = new(aaRoomT)
	default:
		panic("this code should not be executed")
	}

	room.CreateRoom(my, id, ev.GetOption(), player.Player)
}

func (my *actorT) NiuniuJoinRoom(player *playerT, ev *cow_proto.NiuniuJoinRoom) {
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

func (my *actorT) NiuniuLeaveRoom(player *playerT, ev *cow_proto.NiuniuLeaveRoom) {
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

func (my *actorT) NiuniuSwitchReady(player *playerT, ev *cow_proto.NiuniuSwitchReady) {
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

func (my *actorT) NiuniuDismiss(player *playerT, ev *cow_proto.NiuniuDismiss) {
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

func (my *actorT) NiuniuKickPlayer(player *playerT, ev *cow_proto.NiuniuKickPlayer) {
	roomId := int32(0)

	if ev.GetRoomId() != 0 {
		roomId = ev.GetRoomId()
	} else if player.InsideCow != 0 {
		roomId = player.InsideCow
	}

	if roomId == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("kick player but room not found")
		return
	}

	room, being := my.cowRooms[roomId]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"room_id": player.InsideCow,
		}).Warnln("kick player but room not found")
		return
	}

	room.KickPlayer(player, database.Player(ev.GetPlayerId()))
}

func (my *actorT) NiuniuStart(player *playerT, ev *cow_proto.NiuniuStart) {
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

func (my *actorT) NiuniuSpecifyBanker(player *playerT, ev *cow_proto.NiuniuSpecifyBanker) {
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

func (my *actorT) NiuniuGrab(player *playerT, ev *cow_proto.NiuniuGrab) {
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

func (my *actorT) NiuniuSpecifyRate(player *playerT, ev *cow_proto.NiuniuSpecifyRate) {
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

func (my *actorT) NiuniuContinueWith(player *playerT, ev *cow_proto.NiuniuContinueWith) {
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
