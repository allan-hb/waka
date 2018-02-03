package hall

import (
	"github.com/golang/protobuf/proto"

	"github.com/liuhan907/waka/waka-cow2/database"
	"github.com/liuhan907/waka/waka-cow2/proto"
	"github.com/liuhan907/waka/waka/modules/supervisor/supervisor_message"
)

func (my *actorT) playerFutureRequestedCow(player *playerT, ev *supervisor_message.PlayerFutureRequested) bool {
	switch evd := ev.Payload.(type) {
	case *cow_proto.NiuniuGetPayForAnotherRoomListRequest:
		my.NiuniuGetPayForAnotherRoomListRequest(player, evd, ev.Respond)
	case *cow_proto.NiuniuGetWarHistoryRequest:
		my.NiuniuGetWarHistoryRequest(player, evd, ev.Respond)
	default:
		return false
	}
	return true
}

func (my *actorT) NiuniuGetPayForAnotherRoomListRequest(player *playerT,
	ev *cow_proto.NiuniuGetPayForAnotherRoomListRequest,
	respond func(proto.Message, error)) {

	rooms := my.cowRooms.
		WherePayForAnother().
		WhereCreator(player.Player)
	respond(&cow_proto.NiuniuGetPayForAnotherRoomListResponse{rooms.NiuniuRoomData1()}, nil)
}

func (my *actorT) NiuniuGetWarHistoryRequest(player *playerT,
	ev *cow_proto.NiuniuGetWarHistoryRequest,
	respond func(proto.Message, error)) {

	records, err := database.CowQueryWarHistory(player.Player, 20)
	if err != nil {
		respond(nil, err)
		return
	}

	respond(&cow_proto.NiuniuGetWarHistoryResponse{records}, nil)
}
