package hall

import (
	"sort"

	"github.com/golang/protobuf/proto"
	"gopkg.in/ahmetb/go-linq.v3"

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

	case *cow_proto.NiuniuPullFriendsListRequest:
		my.NiuniuPullFriendsListRequest(player, evd, ev.Respond)
	case *cow_proto.NiuniuPullWantListRequest:
		my.NiuniuPullWantListRequest(player, evd, ev.Respond)
	case *cow_proto.NiuniuPullAskListRequest:
		my.NiuniuPullAskListRequest(player, evd, ev.Respond)
	case *cow_proto.NiuniuPullBanListRequest:
		my.NiuniuPullBanListRequest(player, evd, ev.Respond)
	case *cow_proto.NiuniuBanFriendRequest:
		my.NiuniuBanFriendRequest(player, evd, ev.Respond)
	case *cow_proto.NiuniuWantFriendRequest:
		my.NiuniuWantFriendRequest(player, evd, ev.Respond)
	case *cow_proto.NiuniuBecomeFriendRequest:
		my.NiuniuBecomeFriendRequest(player, evd, ev.Respond)
	case *cow_proto.NiuniuCancelBanFriendRequest:
		my.NiuniuCancelBanFriendRequest(player, evd, ev.Respond)

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

func (my *actorT) NiuniuPullFriendsListRequest(player *playerT,
	ev *cow_proto.NiuniuPullFriendsListRequest,
	respond func(proto.Message, error)) {

	friends, err := database.QueryFriendList(player.Player)
	if err != nil {
		respond(nil, err)
		return
	}

	var d []*cow_proto.NiuniuPullFriendsListResponse_NiuniuFriend
	linq.From(friends).SelectT(func(x *database.FriendData) *cow_proto.NiuniuPullFriendsListResponse_NiuniuFriend {
		online := true
		if playerData, being := my.players[x.Friend]; !being || playerData.Remote == "" {
			online = false
		}
		return &cow_proto.NiuniuPullFriendsListResponse_NiuniuFriend{
			PlayerId: int32(x.Friend),
			Nickname: x.Friend.PlayerData().Nickname,
			Online:   online,
		}
	}).ToSlice(&d)

	respond(&cow_proto.NiuniuPullFriendsListResponse{
		Friends: d,
	}, nil)
}

func (my *actorT) NiuniuPullWantListRequest(player *playerT,
	ev *cow_proto.NiuniuPullWantListRequest,
	respond func(proto.Message, error)) {

	wants, err := database.QueryWantListSend(player.Player)
	if err != nil {
		respond(nil, err)
		return
	}

	if len(wants)+10 < 20 {
		wantsDeal, err := database.QueryWantListDeal(player.Player, int32(20-len(wants)))
		if err != nil {
			respond(nil, err)
			return
		}
		wants = append(wants, wantsDeal...)
	}

	sort.Slice(wants, func(i, j int) bool {
		return wants[j].CreatedAt.Unix() < wants[i].CreatedAt.Unix()
	})

	var d []*cow_proto.NiuniuPullWantListResponse_NiuniuFriend
	linq.From(wants).SelectT(func(x *database.AskData) *cow_proto.NiuniuPullWantListResponse_NiuniuFriend {
		online := true
		if playerData, being := my.players[x.Sender]; !being || playerData.Remote == "" {
			online = false
		}
		return &cow_proto.NiuniuPullWantListResponse_NiuniuFriend{
			PlayerId: int32(x.Player),
			Nickname: x.Player.PlayerData().Nickname,
			Online:   online,
			Status:   x.Status,
		}
	}).ToSlice(&d)

	respond(&cow_proto.NiuniuPullWantListResponse{
		Friends: d,
	}, nil)
}

func (my *actorT) NiuniuPullAskListRequest(player *playerT,
	ev *cow_proto.NiuniuPullAskListRequest,
	respond func(proto.Message, error)) {

	asks, err := database.QueryAskListUndeal(player.Player)
	if err != nil {
		respond(nil, err)
		return
	}

	if len(asks)+10 < 20 {
		asksDeal, err := database.QueryAskListDeal(player.Player, int32(20-len(asks)))
		if err != nil {
			respond(nil, err)
			return
		}
		asks = append(asks, asksDeal...)
	}

	sort.Slice(asks, func(i, j int) bool {
		return asks[j].CreatedAt.Unix() < asks[i].CreatedAt.Unix()
	})

	var d []*cow_proto.NiuniuPullAskListResponse_NiuniuFriend
	linq.From(asks).SelectT(func(x *database.AskData) *cow_proto.NiuniuPullAskListResponse_NiuniuFriend {
		online := true
		if playerData, being := my.players[x.Sender]; !being || playerData.Remote == "" {
			online = false
		}
		return &cow_proto.NiuniuPullAskListResponse_NiuniuFriend{
			PlayerId: int32(x.Sender),
			Nickname: x.Sender.PlayerData().Nickname,
			Online:   online,
			Status:   x.Status,
			Number:   x.Id,
		}
	}).ToSlice(&d)

	respond(&cow_proto.NiuniuPullAskListResponse{
		Friends: d,
	}, nil)
}

func (my *actorT) NiuniuPullBanListRequest(player *playerT,
	ev *cow_proto.NiuniuPullBanListRequest,
	respond func(proto.Message, error)) {

	friends, err := database.QueryBanFriendList(player.Player)
	if err != nil {
		respond(nil, err)
		return
	}

	var d []*cow_proto.NiuniuPullBanListResponse_NiuniuFriend
	linq.From(friends).SelectT(func(x *database.FriendData) *cow_proto.NiuniuPullBanListResponse_NiuniuFriend {
		online := true
		if playerData, being := my.players[x.Friend]; !being || playerData.Remote == "" {
			online = false
		}
		return &cow_proto.NiuniuPullBanListResponse_NiuniuFriend{
			PlayerId: int32(x.Friend),
			Nickname: x.Friend.PlayerData().Nickname,
			Online:   online,
		}
	}).ToSlice(&d)

	respond(&cow_proto.NiuniuPullBanListResponse{
		Friends: d,
	}, nil)
}

func (my *actorT) NiuniuBanFriendRequest(player *playerT,
	ev *cow_proto.NiuniuBanFriendRequest,
	respond func(proto.Message, error)) {

	err := database.BanFriend(player.Player, database.Player(ev.GetPlayerId()))
	if err != nil {
		respond(nil, err)
	} else {
		respond(&cow_proto.NiuniuBanFriendResponse{}, nil)
	}
}

func (my *actorT) NiuniuCancelBanFriendRequest(player *playerT,
	ev *cow_proto.NiuniuCancelBanFriendRequest,
	respond func(proto.Message, error)) {

	err := database.CancelBanFriend(player.Player, database.Player(ev.GetPlayerId()))
	if err != nil {
		respond(nil, err)
	} else {
		respond(&cow_proto.NiuniuCancelBanFriendResponse{}, nil)
	}
}

func (my *actorT) NiuniuWantFriendRequest(player *playerT,
	ev *cow_proto.NiuniuWantFriendRequest,
	respond func(proto.Message, error)) {

	if err := database.WantFriend(player.Player, database.Player(ev.GetPlayerId())); err != nil {
		respond(nil, err)
	} else {
		respond(&cow_proto.NiuniuWantFriendResponse{}, nil)
	}
}

func (my *actorT) NiuniuBecomeFriendRequest(player *playerT,
	ev *cow_proto.NiuniuBecomeFriendRequest,
	respond func(proto.Message, error)) {

	if err := database.ReplayAskFriend(ev.GetNumber(), ev.GetOperate()); err != nil {
		respond(nil, err)
	} else {
		respond(&cow_proto.NiuniuBecomeFriendResponse{}, nil)
	}
}
