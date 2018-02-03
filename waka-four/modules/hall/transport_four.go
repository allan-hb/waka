package hall

import (
	"github.com/liuhan907/waka/waka-four/proto"
	"github.com/liuhan907/waka/waka/modules/supervisor/supervisor_message"
	"github.com/sirupsen/logrus"
)

func (my *actorT) playerTransportedFour(player *playerT, ev *supervisor_message.PlayerTransported) bool {
	switch evd := ev.Payload.(type) {
	case *four_proto.FourCreateRoom:
		my.FourCreateRoom(player, evd)
	case *four_proto.FourJoinRoom:
		my.FourJoinRoom(player, evd)
	case *four_proto.FourSwitchReady:
		my.FourSwitchReady(player, evd)
	case *four_proto.FourLeaveRoom:
		my.FourLeaveRoom(player, evd)
	case *four_proto.FourDismiss:
		my.FourDismiss(player, evd)
	case *four_proto.FourDismissVote:
		my.FourDismissVote(player, evd)
	case *four_proto.FourStart:
		my.FourStart(player, evd)
	case *four_proto.FourCut:
		my.FourCut(player, evd)
	case *four_proto.FourCommitPokers:
		my.FourCommitPokers(player, evd)
	case *four_proto.FourSendMessage:
		my.FourSendMessage(player, evd)
	case *four_proto.FourContinueWith:
		my.FourContinueWith(player, evd)
	default:
		return false
	}
	return true
}

func (my *actorT) FourCreateRoom(player *playerT, ev *four_proto.FourCreateRoom) {
	if player.InsideFour != 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("create four room but already in room")
		my.sendFourCreateRoomFailed(player.Player, 2)
		return
	}

	option := ev.Option
	if (option.GetRounds() != 8 && option.GetRounds() != 16 && option.GetRounds() != 24) ||
		(option.GetRate() != 1 && option.GetRate() != 2 && option.GetRate() != 3) ||
		option.GetRuleMode() != 1 ||
		(option.GetPayMode() != 1 && option.GetPayMode() != 2 && option.GetPayMode() != 3) {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("create four room but option illegal")
		my.sendFourCreateRoomFailed(player.Player, 0)
		return
	}

	id, ok := my.fourPlayerNumberPool.Acquire()
	if !ok {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("create four room but acquire room id failed")
		my.sendFourCreateRoomFailed(player.Player, 0)
		return
	}

	var room fourRoomT

	if option.GetPayMode() == 1 || option.GetPayMode() == 2 {
		room = new(fourOrderRoomT)
	} else if option.GetPayMode() == 3 {
		room = new(fourPayForAnotherRoomT)
	} else {
		panic("this code should not be executed")
	}

	room.CreateRoom(my, id, ev.GetOption(), player.Player)
}

func (my *actorT) FourJoinRoom(player *playerT, ev *four_proto.FourJoinRoom) {
	if player.InsideFour != 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("join four room but already in room")
		my.sendFourJoinRoomFailed(player.Player, 4)
		return
	}

	room, being := my.fourRooms[ev.GetRoomId()]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"room_id": ev.GetRoomId(),
		}).Warnln("join four room but not found")
		my.sendFourJoinRoomFailed(player.Player, 1)
		return
	}

	room.JoinRoom(player)
}

func (my *actorT) FourSwitchReady(player *playerT, ev *four_proto.FourSwitchReady) {
	if player.InsideFour == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("switch ready but not in room")
		return
	}

	room, being := my.fourRooms[player.InsideFour]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"room_id": player.InsideFour,
		}).Warnln("switch ready but not room not found")
		player.InsideFour = 0
		return
	}

	room.SwitchReady(player)
}

func (my *actorT) FourLeaveRoom(player *playerT, ev *four_proto.FourLeaveRoom) {
	if player.InsideFour == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("leave four room but not had")
		return
	}

	room, being := my.fourRooms[player.InsideFour]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"room_id": player.InsideFour,
		}).Warnln("leave four room but not found")
		player.InsideFour = 0
		return
	}

	room.LeaveRoom(player)
}

func (my *actorT) FourDismiss(player *playerT, ev *four_proto.FourDismiss) {
	roomId := int32(0)

	if ev.GetRoomId() != 0 {
		roomId = ev.GetRoomId()
	} else if player.InsideFour != 0 {
		roomId = player.InsideFour
	}

	if roomId == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("dismiss room but not found")
		return
	}

	room, being := my.fourRooms[roomId]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"room_id": player.InsideFour,
		}).Warnln("dismiss room but not found")
		return
	}

	room.Dismiss(player)
}

func (my *actorT) FourDismissVote(player *playerT, ev *four_proto.FourDismissVote) {
	if player.InsideFour == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("dismiss vote but not in room")
		return
	}

	room, being := my.fourRooms[player.InsideFour]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"room_id": player.InsideFour,
		}).Warnln("dismiss vote but room not found")
		player.InsideFour = 0
		return
	}

	room.DismissVote(player, ev.GetPassing())
}

func (my *actorT) FourStart(player *playerT, ev *four_proto.FourStart) {
	if player.InsideFour == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("start but not in room")
		return
	}

	room, being := my.fourRooms[player.InsideFour]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"room_id": player.InsideFour,
		}).Warnln("start but room not found")
		player.InsideFour = 0
		return
	}

	room.Start(player)
}

func (my *actorT) FourCut(player *playerT, ev *four_proto.FourCut) {
	if player.InsideFour == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("cut but not in room")
		return
	}

	room, being := my.fourRooms[player.InsideFour]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"room_id": player.InsideFour,
		}).Warnln("cut but room not found")
		player.InsideFour = 0
		return
	}

	room.Cut(player, ev.GetPos())
}

func (my *actorT) FourCommitPokers(player *playerT, ev *four_proto.FourCommitPokers) {
	if player.InsideFour == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("commit pokers but not in room")
		return
	}

	room, being := my.fourRooms[player.InsideFour]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"room_id": player.InsideFour,
		}).Warnln("commit pokers but room not found")
		player.InsideFour = 0
		return
	}

	room.CommitPokers(player, ev.GetFront(), ev.GetBehind())
}

func (my *actorT) FourSendMessage(player *playerT, ev *four_proto.FourSendMessage) {
	if player.InsideFour == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("send message but not in room")
		return
	}

	room, being := my.fourRooms[player.InsideFour]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"room_id": player.InsideFour,
		}).Warnln("send message but room not found")
		player.InsideFour = 0
		return
	}

	room.SendMessage(player, ev.GetMessage().GetType(), ev.GetMessage().GetText())
}

func (my *actorT) FourContinueWith(player *playerT, ev *four_proto.FourContinueWith) {
	if player.InsideFour == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("continue but not in room")
		return
	}

	room, being := my.fourRooms[player.InsideFour]
	if !being {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"room_id": player.InsideFour,
		}).Warnln("continue but room not found")
		player.InsideFour = 0
		return
	}

	room.ContinueWith(player)
}
