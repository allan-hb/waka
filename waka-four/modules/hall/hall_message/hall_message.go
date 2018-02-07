package hall_message

import "github.com/liuhan907/waka/waka-four/database"

type GetTotalOnlineNumber struct {
	Respond func(payload string, err error)
}

type GetTotalOnline struct {
	Respond func(payload string, err error)
}

type GetTotalRoom struct {
	Respond func(payload string, err error)
}

type UpdatePlayerSecret struct {
	Player database.Player
}
