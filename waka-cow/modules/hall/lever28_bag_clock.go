package hall

func (my *actorT) lever28BagClock() {
	originNumber := len(my.lever28Bags)

	for _, bag := range my.lever28Bags {
		bag.Clock()
	}

	if len(my.lever28Bags) != originNumber {
		for _, player := range my.players.SelectOnline() {
			my.sendLever28UpdateBagList(player.Player, my.lever28Bags)
		}
	}
}
