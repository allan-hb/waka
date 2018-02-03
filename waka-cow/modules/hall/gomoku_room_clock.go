package hall

func (my *actorT) gomokuClock() {
	for _, room := range my.gomokuRooms {
		if room.Tick != nil {
			room.Tick()
		}
	}
}
