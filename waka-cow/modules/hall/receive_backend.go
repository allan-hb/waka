package hall

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/liuhan907/waka/waka-cow/modules/hall/hall_message"
	"github.com/liuhan907/waka/waka-cow/proto"
	"github.com/sirupsen/logrus"
)

func (my *actorT) ReceiveBackend(context actor.Context) bool {
	switch ev := context.Message().(type) {
	case *hall_message.GetSupervisorRoom:
		my.GetSupervisorRoom(ev)
	case *hall_message.GetPlayerRoom:
		my.GetPlayerRoom(ev)
	case *hall_message.GetOnlinePlayer:
		my.GetOnlinePlayer(ev)
	case *hall_message.KickPlayer:
		my.KickPlayer(ev)
	case *hall_message.KickRoom:
		my.KickRoom(ev)
	default:
		return false
	}
	return true
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) GetSupervisorRoom(ev *hall_message.GetSupervisorRoom) {
	if ev.Player != 0 {
		log.WithFields(logrus.Fields{
			"player": ev.Player,
		}).Debugln("get supervisor room list")
	} else {
		log.Debugln("get supervisor room list all")
	}

	var response cowRoomMapT
	if ev.Player != 0 {
		response = my.cowRooms.WhereSupervisor().WhereCreator(ev.Player)
	} else {
		response = my.cowRooms.WhereSupervisor()
	}

	ev.Respond(&waka.GetRoomResponse{
		response.NiuniuRoomData2(),
	}, nil)
}

func (my *actorT) GetPlayerRoom(ev *hall_message.GetPlayerRoom) {
	log.Debugln("get player room list all")

	ev.Respond(&waka.GetRoomResponse{
		my.cowRooms.WherePlayer().NiuniuRoomData2(),
	}, nil)
}

func (my *actorT) GetOnlinePlayer(ev *hall_message.GetOnlinePlayer) {
	log.Debugln("get online player list all")

	ev.Respond(&waka.GetOnlinePlayerResponse{
		my.players.SelectOnline().ToSlice(),
	}, nil)
}

func (my *actorT) KickPlayer(ev *hall_message.KickPlayer) {
	log.WithField("player", ev.Player).Debugln("kick player")

	playerData, being := my.players[ev.Player]
	if !being {
		return
	}

	playerData.InsideCow = 0
}

func (my *actorT) KickRoom(ev *hall_message.KickRoom) {
	log.WithField("room_id", ev.Room).Debugln("kick room")

	room, being := my.cowRooms[ev.Room]
	if !being {
		return
	}

	delete(my.cowRooms, ev.Room)

	for _, player := range room.GetPlayers() {
		playerData, being := my.players[player]
		if !being {
			continue
		}
		playerData.InsideCow = 0
	}
	for _, observer := range room.GetObservers() {
		playerData, being := my.players[observer]
		if !being {
			continue
		}
		playerData.InsideCow = 0
	}
}
