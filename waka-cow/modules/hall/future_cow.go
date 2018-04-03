package hall

import (
	"sort"

	"github.com/golang/protobuf/proto"

	"github.com/liuhan907/waka/waka-cow/database"
	"github.com/liuhan907/waka/waka-cow/proto"
	"github.com/liuhan907/waka/waka/modules/supervisor/supervisor_message"
)

func (my *actorT) playerFutureRequestedCow(player *playerT, ev *supervisor_message.PlayerFutureRequested) bool {
	switch evd := ev.Payload.(type) {
	case *cow_proto.NiuniuQueryPayForAnotherRoomListRequest:
		my.NiuniuQueryPayForAnotherRoomListRequest(player, evd, ev.Respond)
	case *cow_proto.NiuniuQueryFlowingRoomListRequest:
		my.NiuniuQueryFlowingRoomListRequest(player, evd, ev.Respond)
	case *cow_proto.NiuniuQueryHistoryRequest:
		my.NiuniuQueryHistoryRequest(player, evd, ev.Respond)
	default:
		return false
	}
	return true
}

func (my *actorT) NiuniuQueryPayForAnotherRoomListRequest(player *playerT,
	ev *cow_proto.NiuniuQueryPayForAnotherRoomListRequest,
	respond func(proto.Message, error)) {

	rooms := my.cowRooms.
		WherePayForAnother().
		WhereCreator(player.Player)

	pb := rooms.NiuniuRoomData()
	sort.Slice(pb, func(i, j int) bool {
		return pb[i].GetRoomId() < pb[j].GetRoomId()
	})

	respond(&cow_proto.NiuniuQueryPayForAnotherRoomListResponse{pb}, nil)
}

func (my *actorT) NiuniuQueryFlowingRoomListRequest(player *playerT,
	ev *cow_proto.NiuniuQueryFlowingRoomListRequest,
	respond func(proto.Message, error)) {

	rooms := my.cowRooms.
		WhereFlowing().
		WhereReady()

	pb := rooms.NiuniuRoomData()
	sort.Slice(pb, func(i, j int) bool {
		return pb[i].GetRoomId() < pb[j].GetRoomId()
	})

	respond(&cow_proto.NiuniuQueryFlowingRoomListResponse{pb}, nil)
}

func (my *actorT) NiuniuQueryHistoryRequest(player *playerT,
	ev *cow_proto.NiuniuQueryHistoryRequest,
	respond func(proto.Message, error)) {

	records, err := database.CowQueryHistory(player.Player, 20)
	if err != nil {
		respond(nil, err)
		return
	}

	respond(&cow_proto.NiuniuQueryHistoryResponse{records}, nil)
}
