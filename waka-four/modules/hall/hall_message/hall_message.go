package hall_message

type GetTotalOnlineNumber struct {
	Respond func(payload string, err error)
}

type GetTotalOnline struct {
	Respond func(payload string, err error)
}

type GetTotalRoom struct {
	Respond func(payload string, err error)
}
