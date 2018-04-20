package hall

import (
	"github.com/golang/protobuf/proto"

	"github.com/liuhan907/waka/waka-cow/proto"
	"github.com/liuhan907/waka/waka/modules/supervisor/supervisor_message"
)

func (my *actorT) playerFutureRequestedPlayer(player *playerT, ev *supervisor_message.PlayerFutureRequested) bool {
	switch evd := ev.Payload.(type) {
	case *cow_proto.GetMyRequest:
		my.GetMyRequest(player, evd, ev.Respond)
	case *cow_proto.GetPlayerRequest:
		my.GetPlayerRequest(player, evd, ev.Respond)
	default:
		return false
	}
	return true
}

func (my *actorT) GetMyRequest(player *playerT,
	ev *cow_proto.GetMyRequest,
	respond func(proto.Message, error)) {
	playerData := my.ToPlayerSecret(player.Player)
	r := new(cow_proto.PlayerSecret)

	if ev.Mask&cow_proto.PlayerMask_ID > 0 {
		r.Id = playerData.Id
	}
	if ev.Mask&cow_proto.PlayerMask_CreatedAt > 0 {
		r.CreatedAt = playerData.CreatedAt
	}
	if ev.Mask&cow_proto.PlayerMask_WechatUID > 0 {
		r.WechatUid = playerData.WechatUid
	}
	if ev.Mask&cow_proto.PlayerMask_Nickname > 0 {
		r.Nickname = playerData.Nickname
	}
	if ev.Mask&cow_proto.PlayerMask_Head > 0 {
		r.Head = playerData.Head
	}
	if ev.Mask&cow_proto.PlayerMask_Money > 0 {
		r.Money = playerData.Money
	}
	if ev.Mask&cow_proto.PlayerMask_Vip > 0 {
		r.Vip = playerData.Vip
	}
	if ev.Mask&cow_proto.PlayerMask_Wechat > 0 {
		r.Wechat = playerData.Wechat
	}
	if ev.Mask&cow_proto.PlayerMask_Idcard > 0 {
		r.Idcard = playerData.Idcard
	}
	if ev.Mask&cow_proto.PlayerMask_Name > 0 {
		r.Name = playerData.Name
	}
	if ev.Mask&cow_proto.PlayerMask_Supervisor > 0 {
		r.Supervisor = playerData.Supervisor
	}
	if ev.Mask&cow_proto.PlayerMask_Ip > 0 {
		r.Ip = playerData.Ip
	}

	respond(&cow_proto.GetMyResponse{r}, nil)
}

func (my *actorT) GetPlayerRequest(player *playerT,
	ev *cow_proto.GetPlayerRequest,
	respond func(proto.Message, error)) {
	playerData := my.ToPlayer(player.Player)
	r := new(cow_proto.Player)

	if ev.Mask&cow_proto.PlayerMask_ID > 0 {
		r.Id = playerData.Id
	}
	if ev.Mask&cow_proto.PlayerMask_Nickname > 0 {
		r.Nickname = playerData.Nickname
	}
	if ev.Mask&cow_proto.PlayerMask_Head > 0 {
		r.Head = playerData.Head
	}
	if ev.Mask&cow_proto.PlayerMask_Money > 0 {
		r.Money = playerData.Money
	}
	if ev.Mask&cow_proto.PlayerMask_Vip > 0 {
		r.Vip = playerData.Vip
	}
	if ev.Mask&cow_proto.PlayerMask_Wechat > 0 {
		r.Wechat = playerData.Wechat
	}
	if ev.Mask&cow_proto.PlayerMask_Ip > 0 {
		r.Ip = playerData.Ip
	}

	respond(&cow_proto.GetPlayerResponse{r}, nil)
}
