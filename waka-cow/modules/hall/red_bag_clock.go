package hall

func (my *actorT) redBagClock() {
	originNumber := len(my.redBags)

	for _, bag := range my.redBags {
		bag.Clock()
	}

	if len(my.redBags) != originNumber {
		for _, player := range my.players.SelectOnline() {
			my.sendRedUpdateRedPaperBagList(player.Player, my.redBags)
		}
	}
}
