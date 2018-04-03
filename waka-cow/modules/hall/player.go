package hall

import (
	"time"

	"github.com/liuhan907/waka/waka-cow/database"
	"github.com/liuhan907/waka/waka-cow/proto"
)

type playerT struct {
	Player database.Player

	Remote string

	InsideCow     int32
	InsideRed     int32
	InsideLever28 int32
	InsideGomoku  int32
}

type playerMap map[database.Player]*playerT

func (my playerMap) SelectOnline() playerMap {
	r := make(playerMap, len(my))
	for _, player := range my {
		if player.Remote != "" {
			r[player.Player] = player
		}
	}
	return r
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) ToPlayer(player database.Player) (pb *cow_proto.Player) {
	pb = &cow_proto.Player{}

	playerData := player.PlayerData()
	pb.Id = int32(playerData.Id)
	pb.Nickname = playerData.Nickname
	pb.Head = playerData.Head
	pb.Money = playerData.Money / 100
	pb.Vip = int64(playerData.Vip.Sub(time.Now()).Seconds() / (24 * 60 * 60))
	pb.Wechat = playerData.Wechat

	localPlayer, being := my.players[player]
	if being {
		pb.Ip = localPlayer.Remote
	}
	if pb.Vip <= 0 {
		pb.Vip = 0
	}

	return pb
}

func (my *actorT) ToPlayerSecret(player database.Player) (pb *cow_proto.PlayerSecret) {
	pb = &cow_proto.PlayerSecret{}

	playerData := player.PlayerData()
	pb.Id = int32(playerData.Id)
	pb.WechatUid = playerData.WechatUnionid
	pb.Nickname = playerData.Nickname
	pb.Head = playerData.Head
	pb.Wechat = playerData.Wechat
	pb.Idcard = playerData.Idcard
	pb.Name = playerData.Name
	pb.Money = playerData.Money / 100
	pb.Vip = int64(playerData.Vip.Sub(time.Now()).Seconds() / (24 * 60 * 60))
	pb.Supervisor = int32(playerData.Supervisor)
	pb.CreatedAt = playerData.CreatedAt.Format("2006-01-02 15:04:05")

	localPlayer, being := my.players[player]
	if being {
		pb.Ip = localPlayer.Remote
	}
	if pb.Vip <= 0 {
		pb.Vip = 0
	}

	return pb
}

func (my *actorT) ToPlayerMap(players map[database.Player]database.Player) (pb []*cow_proto.Player) {
	for _, player := range players {
		pb = append(pb, my.ToPlayer(player))
	}
	return pb
}
