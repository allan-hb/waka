package hall

import (
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/liuhan907/waka/waka-cow/database"
	"github.com/liuhan907/waka/waka-cow/modules/hall/hall_message"
	"gopkg.in/ahmetb/go-linq.v3"
)

func (my *actorT) ReceiveBackend(context actor.Context) bool {
	switch evd := context.Message().(type) {
	case *hall_message.GetFlowingRoom:
		my.GetFlowingRoom(evd)
	case *hall_message.GetPlayerRoom:
		my.GetPlayerRoom(evd)
	case *hall_message.GetOnlinePlayer:
		my.GetOnlinePlayer(evd)
	default:
		return false
	}
	return true
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) GetFlowingRoom(evd *hall_message.GetFlowingRoom) {
	evd.Respond(my.cowRooms.WhereFlowing().NiuniuRoomData(), nil)
}

func (my *actorT) GetPlayerRoom(evd *hall_message.GetPlayerRoom) {
	r := my.cowRooms.WherePlayer()
	if evd.Player != 0 {
		r = r.WhereCreator(evd.Player)
	}
	evd.Respond(r.NiuniuRoomData(), nil)
}

func (my *actorT) GetOnlinePlayer(evd *hall_message.GetOnlinePlayer) {
	var r []int32
	linq.From(my.players.SelectOnline()).SelectT(func(in linq.KeyValue) int32 {
		return int32(in.Key.(database.Player))
	}).ToSlice(&r)
	evd.Respond(r, nil)
}
