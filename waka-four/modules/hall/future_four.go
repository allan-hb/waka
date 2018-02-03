package hall

import (
	"sort"

	"github.com/golang/protobuf/proto"
	"github.com/liuhan907/waka/waka-four/database"
	"github.com/liuhan907/waka/waka-four/proto"
	"github.com/liuhan907/waka/waka/modules/supervisor/supervisor_message"
	"gopkg.in/ahmetb/go-linq.v3"
)

func (my *actorT) playerFutureRequestedFour(player *playerT, ev *supervisor_message.PlayerFutureRequested) bool {
	switch evd := ev.Payload.(type) {
	case *four_proto.PullPlayerRequest:
		my.PullPlayerRequest(player, evd, ev.Respond)
	case *four_proto.PullPlayerSecretRequest:
		my.PullPlayerSecretRequest(player, evd, ev.Respond)
	case *four_proto.FourPullFriendsListRequest:
		my.FourPullFriendsListRequest(player, evd, ev.Respond)
	case *four_proto.FourPullWantListRequest:
		my.FourPullWantListRequest(player, evd, ev.Respond)
	case *four_proto.FourPullAskListRequest:
		my.FourPullAskListRequest(player, evd, ev.Respond)
	case *four_proto.FourPullBanListRequest:
		my.FourPullBanListRequest(player, evd, ev.Respond)
	case *four_proto.FourBanFriendRequest:
		my.FourBanFriendRequest(player, evd, ev.Respond)
	case *four_proto.FourWantFriendRequest:
		my.FourWantFriendRequest(player, evd, ev.Respond)
	case *four_proto.FourBecomeFriendRequest:
		my.FourBecomeFriendRequest(player, evd, ev.Respond)
	case *four_proto.FourCancelBanFriendRequest:
		my.FourCancelBanFriendRequest(player, evd, ev.Respond)
	case *four_proto.FourPullPayForAnotherRoomListRequest:
		my.FourPullPayForAnotherRoomListRequest(player, evd, ev.Respond)
	case *four_proto.FourPullWarHistoryListRequest:
		my.FourPullWarHistoryListRequest(player, evd, ev.Respond)
	default:
		return false
	}
	return true
}

func (my *actorT) PullPlayerRequest(player *playerT,
	ev *four_proto.PullPlayerRequest,
	respond func(proto.Message, error)) {
	respond(&four_proto.PullPlayerResponse{my.ToPlayer(database.Player(ev.GetPlayerId()))}, nil)
}

func (my *actorT) PullPlayerSecretRequest(player *playerT,
	ev *four_proto.PullPlayerSecretRequest,
	respond func(proto.Message, error)) {
	respond(&four_proto.PullPlayerSecretResponse{my.ToPlayerSecret(database.Player(player.Player))}, nil)
}

func (my *actorT) FourPullFriendsListRequest(player *playerT,
	ev *four_proto.FourPullFriendsListRequest,
	respond func(proto.Message, error)) {

	friends, err := database.QueryFriendList(player.Player)
	if err != nil {
		respond(nil, err)
		return
	}

	var d []*four_proto.FourPullFriendsListResponse_FourFriend
	linq.From(friends).SelectT(func(x *database.FriendData) *four_proto.FourPullFriendsListResponse_FourFriend {
		online := true
		if playerData, being := my.players[x.Friend]; !being || playerData.Remote == "" {
			online = false
		}
		return &four_proto.FourPullFriendsListResponse_FourFriend{
			PlayerId: int32(x.Friend),
			Nickname: x.Friend.PlayerData().Nickname,
			Online:   online,
		}
	}).ToSlice(&d)

	respond(&four_proto.FourPullFriendsListResponse{
		Friends: d,
	}, nil)
}

func (my *actorT) FourPullWantListRequest(player *playerT,
	ev *four_proto.FourPullWantListRequest,
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

	var d []*four_proto.FourPullWantListResponse_FourFriend
	linq.From(wants).SelectT(func(x *database.AskData) *four_proto.FourPullWantListResponse_FourFriend {
		online := true
		if playerData, being := my.players[x.Sender]; !being || playerData.Remote == "" {
			online = false
		}
		return &four_proto.FourPullWantListResponse_FourFriend{
			PlayerId: int32(x.Player),
			Nickname: x.Player.PlayerData().Nickname,
			Online:   online,
			Status:   x.Status,
		}
	}).ToSlice(&d)

	respond(&four_proto.FourPullWantListResponse{
		Friends: d,
	}, nil)
}

func (my *actorT) FourPullAskListRequest(player *playerT,
	ev *four_proto.FourPullAskListRequest,
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

	var d []*four_proto.FourPullAskListResponse_FourFriend
	linq.From(asks).SelectT(func(x *database.AskData) *four_proto.FourPullAskListResponse_FourFriend {
		online := true
		if playerData, being := my.players[x.Sender]; !being || playerData.Remote == "" {
			online = false
		}
		return &four_proto.FourPullAskListResponse_FourFriend{
			PlayerId: int32(x.Sender),
			Nickname: x.Sender.PlayerData().Nickname,
			Online:   online,
			Status:   x.Status,
			Number:   x.Id,
		}
	}).ToSlice(&d)

	respond(&four_proto.FourPullAskListResponse{
		Friends: d,
	}, nil)
}

func (my *actorT) FourPullBanListRequest(player *playerT,
	ev *four_proto.FourPullBanListRequest,
	respond func(proto.Message, error)) {

	friends, err := database.QueryBanFriendList(player.Player)
	if err != nil {
		respond(nil, err)
		return
	}

	var d []*four_proto.FourPullBanListResponse_FourFriend
	linq.From(friends).SelectT(func(x *database.FriendData) *four_proto.FourPullBanListResponse_FourFriend {
		online := true
		if playerData, being := my.players[x.Friend]; !being || playerData.Remote == "" {
			online = false
		}
		return &four_proto.FourPullBanListResponse_FourFriend{
			PlayerId: int32(x.Friend),
			Nickname: x.Friend.PlayerData().Nickname,
			Online:   online,
		}
	}).ToSlice(&d)

	respond(&four_proto.FourPullBanListResponse{
		Friends: d,
	}, nil)
}

func (my *actorT) FourBanFriendRequest(player *playerT,
	ev *four_proto.FourBanFriendRequest,
	respond func(proto.Message, error)) {

	err := database.BanFriend(player.Player, database.Player(ev.GetPlayerId()))
	if err != nil {
		respond(nil, err)
	} else {
		respond(&four_proto.FourBanFriendResponse{}, nil)
	}
}

func (my *actorT) FourCancelBanFriendRequest(player *playerT,
	ev *four_proto.FourCancelBanFriendRequest,
	respond func(proto.Message, error)) {

	err := database.CancelBanFriend(player.Player, database.Player(ev.GetPlayerId()))
	if err != nil {
		respond(nil, err)
	} else {
		respond(&four_proto.FourCancelBanFriendResponse{}, nil)
	}
}

func (my *actorT) FourWantFriendRequest(player *playerT,
	ev *four_proto.FourWantFriendRequest,
	respond func(proto.Message, error)) {

	if err := database.WantFriend(player.Player, database.Player(ev.GetPlayerId())); err != nil {
		respond(nil, err)
	} else {
		respond(&four_proto.FourWantFriendResponse{}, nil)
	}
}

func (my *actorT) FourBecomeFriendRequest(player *playerT,
	ev *four_proto.FourBecomeFriendRequest,
	respond func(proto.Message, error)) {

	if err := database.ReplayAskFriend(ev.GetNumber(), ev.GetOperate()); err != nil {
		respond(nil, err)
	} else {
		respond(&four_proto.FourBecomeFriendResponse{}, nil)
	}
}

func (my *actorT) FourPullPayForAnotherRoomListRequest(player *playerT,
	ev *four_proto.FourPullPayForAnotherRoomListRequest,
	respond func(proto.Message, error)) {

	respond(&four_proto.FourPullPayForAnotherRoomListResponse{my.fourRooms.WherePayForAnother().WhereCreator(player.Player).FourRoom1()}, nil)
}

func (my *actorT) FourPullWarHistoryListRequest(player *playerT,
	ev *four_proto.FourPullWarHistoryListRequest,
	respond func(proto.Message, error)) {

	histories, err := database.FourQueryWarHistory(player.Player, 20)
	if err != nil {
		respond(nil, err)
	} else {
		respond(&four_proto.FourPullWarHistoryListResponse{histories}, nil)
	}
}
