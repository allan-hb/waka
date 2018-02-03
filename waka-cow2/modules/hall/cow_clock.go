package hall

func (my *actorT) cowClock1() {
	for _, room := range my.cowRooms {
		room.Tick()
	}
}
