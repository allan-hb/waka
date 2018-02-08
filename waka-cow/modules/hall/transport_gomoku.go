package hall

import (
	"github.com/liuhan907/waka/waka-cow/proto"
	"github.com/liuhan907/waka/waka/modules/supervisor/supervisor_message"
	"github.com/sirupsen/logrus"
)

func (my *actorT) playerTransportedGomoku(player *playerT, ev *supervisor_message.PlayerTransported) bool {
	switch evd := ev.Payload.(type) {
	case *waka.GomokuCreateRoom:
		my.GomokuCreateRoom(player, evd)
	case *waka.GomokuJoinRoom:
		my.GomokuJoinRoom(player, evd)
	case *waka.GomokuSetCost:
		my.GomokuSetCost(player, evd)
	case *waka.GomokuLeave:
		my.GomokuLeave(player, evd)
	case *waka.GomokuDismiss:
		my.GomokuDismiss(player, evd)
	case *waka.GomokuStart:
		my.GomokuStart(player, evd)
	case *waka.GomokuPlay:
		my.GomokuPlay(player, evd)
	case *waka.GomokuSurrender:
		my.GomokuSurrender(player, evd)
	default:
		return false
	}
	return true
}

func (my *actorT) GomokuCreateRoom(player *playerT, ev *waka.GomokuCreateRoom) {
	if player.InsideGomoku != 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("create gomoku but already in room")
		my.sendGomokuCreateRoomFailed(player.Player, 2)
		return
	}

	id, ok := my.gomokuNumberPool.Acquire()
	if !ok {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("create gomoku but acquire id faield")
		my.sendGomokuCreateRoomFailed(player.Player, 0)
		return
	}

	room := new(gomokuRoomT)
	room.Create(my, player, id)
}

func (my *actorT) GomokuJoinRoom(player *playerT, ev *waka.GomokuJoinRoom) {
	if player.InsideGomoku != 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("join gomoku but already in room")
		my.sendGomokuJoinRoomFailed(player.Player, 3)
		return
	}

	room, being := my.gomokuRooms[ev.GetRoomId()]
	if !being {
		log.WithFields(logrus.Fields{
			"player": player.Player,
			"id":     ev.GetRoomId(),
		}).Warnln("join gomoku but room not found")
		my.sendGomokuJoinRoomFailed(player.Player, 1)
		return
	}

	if room.Gaming {
		my.sendGomokuJoinRoomFailed(player.Player, 0)
		return
	}

	if room.Creator != nil && room.Creator.Player == player.Player {
		log.WithFields(logrus.Fields{
			"player": player.Player,
			"id":     ev.GetRoomId(),
		}).Warnln("join gomoku but already in")
		my.sendGomokuJoinRoomFailed(player.Player, 0)
		return
	}

	if room.Student != nil && room.Student.Player == player.Player {
		log.WithFields(logrus.Fields{
			"player": player.Player,
			"id":     ev.GetRoomId(),
		}).Warnln("join gomoku but already in")
		my.sendGomokuJoinRoomFailed(player.Player, 0)
		return
	}

	if room.Creator != nil && room.Student != nil {
		my.sendGomokuJoinRoomFailed(player.Player, 4)
		return
	}

	if player.Player.PlayerData().Money < room.Cost*100 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
			"id":     ev.GetRoomId(),
		}).Warnln("join gomoku but money not enough")
		my.sendGomokuJoinRoomFailed(player.Player, 2)
		return
	}

	room.Join(player)
}

func (my *actorT) GomokuSetCost(player *playerT, ev *waka.GomokuSetCost) {
	if player.InsideGomoku == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("set gomoku cost but not in room")
		my.sendGomokuSetRoomCostFailed(player.Player, 3)
		return
	}

	room, being := my.gomokuRooms[player.InsideGomoku]
	if !being {
		log.WithFields(logrus.Fields{
			"player": player.Player,
			"id":     player.InsideGomoku,
		}).Warnln("set gomoku cost but room not found")
		player.InsideGomoku = 0
		my.sendGomokuSetRoomCostFailed(player.Player, 3)
		return
	}

	if room.Gaming {
		my.sendGomokuSetRoomCostFailed(player.Player, 0)
		return
	}

	if room.Creator.Player != player.Player {
		my.sendGomokuSetRoomCostFailed(player.Player, 4)
		return
	}

	if ev.GetCost() < 550 || ev.GetCost() > 110000 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
			"id":     player.InsideGomoku,
			"cost":   ev.GetCost(),
		}).Warnln("set gomoku cost but number illegal")
		my.sendGomokuSetRoomCostFailed(player.Player, 1)
		return
	}

	if room.Student != nil && room.Student.Player.PlayerData().Money < ev.GetCost()*100 {
		log.WithFields(logrus.Fields{
			"player":  player.Player,
			"student": room.Student.Player,
			"id":      player.InsideGomoku,
			"cost":    ev.GetCost(),
		}).Warnln("set gomoku cost but student money not enough")
		my.sendGomokuSetRoomCostFailed(player.Player, 2)
		return
	}

	room.SetCost(player, ev.Cost)
}

func (my *actorT) GomokuLeave(player *playerT, ev *waka.GomokuLeave) {
	if player.InsideGomoku == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("leave gomoku room but not in room")
		return
	}

	room, being := my.gomokuRooms[player.InsideGomoku]
	if !being {
		log.WithFields(logrus.Fields{
			"player": player.Player,
			"id":     player.InsideGomoku,
		}).Warnln("leave gomoku room but room not found")
		player.InsideGomoku = 0
		return
	}

	if room.Gaming {
		return
	}

	room.Leave(player)
}

func (my *actorT) GomokuDismiss(player *playerT, ev *waka.GomokuDismiss) {
	if player.InsideGomoku == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("dismiss gomoku room but not in room")
		return
	}

	room, being := my.gomokuRooms[player.InsideGomoku]
	if !being {
		log.WithFields(logrus.Fields{
			"player": player.Player,
			"id":     player.InsideGomoku,
		}).Warnln("dismiss gomoku room but room not found")
		player.InsideGomoku = 0
		return
	}

	room.Dismiss(player)
}

func (my *actorT) GomokuStart(player *playerT, ev *waka.GomokuStart) {
	if player.InsideGomoku == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("start gomoku but not in room")
		return
	}

	room, being := my.gomokuRooms[player.InsideGomoku]
	if !being {
		log.WithFields(logrus.Fields{
			"player": player.Player,
			"id":     player.InsideGomoku,
		}).Warnln("start gomoku but room not found")
		player.InsideGomoku = 0
		return
	}

	if room.Creator.Player != player.Player {
		return
	}

	if room.Gaming {
		return
	}

	room.Start(player)
}

func (my *actorT) GomokuPlay(player *playerT, ev *waka.GomokuPlay) {
	if player.InsideGomoku == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("play gomoku but not in room")
		return
	}

	room, being := my.gomokuRooms[player.InsideGomoku]
	if !being {
		log.WithFields(logrus.Fields{
			"player": player.Player,
			"id":     player.InsideGomoku,
		}).Warnln("play gomoku but room not found")
		player.InsideGomoku = 0
		return
	}

	if !room.Gaming {
		return
	}

	room.Play(player, ev.GetX(), ev.GetY())
}

func (my *actorT) GomokuSurrender(player *playerT, ev *waka.GomokuSurrender) {
	if player.InsideGomoku == 0 {
		log.WithFields(logrus.Fields{
			"player": player.Player,
		}).Warnln("surrender gomoku but not in room")
		return
	}

	room, being := my.gomokuRooms[player.InsideGomoku]
	if !being {
		log.WithFields(logrus.Fields{
			"player": player.Player,
			"id":     player.InsideGomoku,
		}).Warnln("surrender gomoku but room not found")
		player.InsideGomoku = 0
		return
	}

	if !room.Gaming {
		return
	}

	room.Surrender(player)
}
