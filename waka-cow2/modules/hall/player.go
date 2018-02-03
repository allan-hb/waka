package hall

import (
	"github.com/liuhan907/waka/waka-cow2/database"
	"github.com/liuhan907/waka/waka-cow2/proto"
)

type playerT struct {
	Player database.Player

	Remote string

	InsideCow int32
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

func (my *actorT) ToPlayer(player database.Player) (pb *cow_proto.Player) {
	playerData := player.PlayerData()

	pb = &cow_proto.Player{
		Id:       int32(playerData.Id),
		Nickname: playerData.Nickname,
		Wechat:   playerData.Wechat,
	}

	localPlayer, being := my.players[player]
	if being {
		pb.Ip = localPlayer.Remote
	}

	return pb
}

func (my *actorT) ToPlayerSecret(player database.Player) (pb *cow_proto.PlayerSecret) {
	playerData := player.PlayerData()

	pb = &cow_proto.PlayerSecret{
		Id:        int32(playerData.Id),
		Nickname:  playerData.Nickname,
		Wechat:    playerData.Wechat,
		Idcard:    playerData.Idcard,
		Name:      playerData.Name,
		CreatedAt: playerData.CreatedAt.Format("2006-01-02 15:04:05"),
		Diamonds:  playerData.Diamonds,
	}

	localPlayer, being := my.players[player]
	if being {
		pb.Ip = localPlayer.Remote
	}

	return pb
}

func (my *actorT) ToPlayerMap(players map[database.Player]database.Player) (pb []*cow_proto.Player) {
	for _, player := range players {
		pb = append(pb, my.ToPlayer(player))
	}
	return pb
}
