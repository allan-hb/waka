package hall

import (
	"encoding/json"

	"github.com/AsynkronIT/protoactor-go/actor"

	"github.com/liuhan907/waka/waka-cow2/modules/hall/hall_message"
)

func (my *actorT) ReceiveBackend(context actor.Context) bool {
	switch ev := context.Message().(type) {
	case *hall_message.GetTotalOnlineNumber:
		my.GetTotalOnlineNumber(ev)
	case *hall_message.GetTotalOnline:
		my.GetTotalOnline(ev)
	case *hall_message.GetTotalRoom:
		my.GetTotalRoom(ev)
	default:
		return false
	}
	return true
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) GetTotalOnlineNumber(ev *hall_message.GetTotalOnlineNumber) {
	d, err := json.Marshal(map[string]interface{}{
		"number": len(my.players.SelectOnline().ToSlice()),
	})
	if err != nil {
		ev.Respond("", err)
	} else {
		ev.Respond(string(d), err)
	}
}

func (my *actorT) GetTotalOnline(ev *hall_message.GetTotalOnline) {
	d, err := json.Marshal(map[string]interface{}{
		"players": my.players.SelectOnline().ToSlice(),
	})
	if err != nil {
		ev.Respond("", err)
	} else {
		ev.Respond(string(d), err)
	}
}

func (my *actorT) GetTotalRoom(ev *hall_message.GetTotalRoom) {
	var rooms []map[string]interface{}
	for _, room := range my.cowRooms {
		rooms = append(rooms, room.BackendRoom())
	}
	d, err := json.Marshal(map[string]interface{}{
		"rooms": rooms,
	})
	if err != nil {
		ev.Respond("", err)
	} else {
		ev.Respond(string(d), err)
	}
}
