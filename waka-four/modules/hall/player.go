package hall

import (
	"github.com/liuhan907/waka/waka-four/database"
	"github.com/liuhan907/waka/waka-four/proto"
)

type playerT struct {
	Player database.Player

	Remote           string
	BackgroundRemote string

	InsideFour int32
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

func (my playerMap) ToSlice() (d []int32) {
	for _, player := range my {
		d = append(d, int32(player.Player))
	}
	return d
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) ToPlayer(player database.Player) (pb *four_proto.Player) {
	pb = &four_proto.Player{}

	playerData := player.PlayerData()
	pb.Id = int32(playerData.Id)
	pb.Nickname = playerData.Nickname
	pb.Head = playerData.Head
	pb.Wechat = playerData.Wechat

	localPlayer, being := my.players[player]
	if being {
		pb.Ip = localPlayer.Remote
	}

	return pb
}

func (my *actorT) ToPlayerSecret(player database.Player) (pb *four_proto.PlayerSecret) {
	pb = &four_proto.PlayerSecret{}

	playerData := player.PlayerData()
	pb.Id = int32(playerData.Id)
	pb.WechatUid = playerData.UnionId
	pb.Nickname = playerData.Nickname
	pb.Head = playerData.Head
	pb.Wechat = playerData.Wechat
	pb.Idcard = playerData.Idcard
	pb.Name = playerData.Name
	pb.Diamonds = playerData.Diamonds
	pb.Supervisor = int32(playerData.Supervisor)
	pb.CreatedAt = playerData.CreatedAt.Unix()

	localPlayer, being := my.players[player]
	if being {
		pb.Ip = localPlayer.Remote
	}

	return pb
}

func (my *actorT) ToPlayerMap(players map[database.Player]database.Player) (pb []*four_proto.Player) {
	for _, player := range players {
		pb = append(pb, my.ToPlayer(player))
	}
	return pb
}
