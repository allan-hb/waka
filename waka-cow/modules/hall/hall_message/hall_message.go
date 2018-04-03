package hall_message

import (
	"github.com/liuhan907/waka/waka-cow/database"
	"github.com/liuhan907/waka/waka-cow/proto"
)

type GetFlowingRoom struct {
	Respond func(response []*cow_proto.NiuniuRoomData, e error)
}

type GetPlayerRoom struct {
	Player  database.Player
	Respond func(response []*cow_proto.NiuniuRoomData, e error)
}

type GetOnlinePlayer struct {
	Respond func(response []int32, e error)
}
