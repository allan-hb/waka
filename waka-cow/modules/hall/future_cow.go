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
	case *waka.NiuniuQueryPayForAnotherRoomListRequest:
		my.NiuniuQueryPayForAnotherRoomListRequest(player, evd, ev.Respond)
	case *waka.NiuniuQueryAgentRoomCountRequest:
		my.NiuniuQueryAgentRoomCountRequest(player, evd, ev.Respond)
	case *waka.NiuniuQueryAgentRoomListRequest:
		my.NiuniuQueryAgentRoomListRequest(player, evd, ev.Respond)
	case *waka.NiuniuQueryRecordRequest:
		my.NiuniuQueryRecordRequest(player, evd, ev.Respond)
	default:
		return false
	}
	return true
}

func (my *actorT) NiuniuQueryPayForAnotherRoomListRequest(player *playerT,
	ev *waka.NiuniuQueryPayForAnotherRoomListRequest,
	respond func(proto.Message, error)) {

	rooms := my.cowRooms.
		WherePayForAnother().
		WhereCreator(player.Player)
	respond(&waka.NiuniuQueryPayForAnotherRoomListResponse{rooms.NiuniuRoomData1()}, nil)
}

func (my *actorT) NiuniuQueryAgentRoomCountRequest(player *playerT,
	ev *waka.NiuniuQueryAgentRoomCountRequest,
	respond func(proto.Message, error)) {

	rooms := my.cowRooms.WhereSupervisor()
	if ev.GetPlayerId() != 0 {
		rooms.WhereCreator(database.Player(ev.GetPlayerId()))
	}

	respond(&waka.NiuniuQueryAgentRoomCountResponse{int32(len(rooms))}, nil)
}

func (my *actorT) NiuniuQueryAgentRoomListRequest(player *playerT,
	ev *waka.NiuniuQueryAgentRoomListRequest,
	respond func(proto.Message, error)) {

	rooms := my.cowRooms.WhereSupervisor()
	if ev.GetPlayerId() != 0 {
		rooms.WhereCreator(database.Player(ev.GetPlayerId()))
	}

	pb := rooms.NiuniuRoomData1()
	sort.Slice(pb, func(i, j int) bool {
		return pb[i].Id < pb[j].Id
	})

	if ev.Range != nil {
		if ev.Range.Start < int32(len(pb)) &&
			ev.Range.Start+ev.Range.Number <= int32(len(pb)) {
			pb = pb[int(ev.Range.Start):int(ev.Range.Start+ev.Range.Number)]
		}
	}

	sort.Slice(pb, func(i, j int) bool {
		return pb[i].Id < pb[j].Id
	})

	respond(&waka.NiuniuQueryAgentRoomListResponse{pb}, nil)
}

func (my *actorT) NiuniuQueryRecordRequest(player *playerT,
	ev *waka.NiuniuQueryRecordRequest,
	respond func(proto.Message, error)) {

	records, err := database.CowQueryWarHistory(player.Player, 20)
	if err != nil {
		respond(nil, err)
		return
	}

	respond(&waka.NiuniuQueryRecordResponse{records}, nil)
}
