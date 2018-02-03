package hall_message

import (
	"github.com/liuhan907/waka/waka-cow/database"
	"github.com/liuhan907/waka/waka-cow/proto"
)

type GetSupervisorRoom struct {
	Player  database.Player
	Respond func(*waka.GetRoomResponse, error)
}

type GetPlayerRoom struct {
	Respond func(*waka.GetRoomResponse, error)
}

type GetOnlinePlayer struct {
	Respond func(*waka.GetOnlinePlayerResponse, error)
}

type KickPlayer struct {
	Player database.Player
}

type KickRoom struct {
	Room int32
}
