package hall

func (r *supervisorRoomT) buildStart() {
	if r.tick == nil && len(r.Players) >= 2 {
		r.tick = buildTickNumber(
			5,
			func(number int32) {
				if !r.Gaming && len(r.Players) >= 2 {
					r.Hall.sendNiuniuCountdownForAll(r, number)
				} else {
					r.Hall.sendNiuniuCountdownForAll(r, 0)
					r.tick = nil
				}
			},
			func() {
				r.tick = nil
				if !r.Gaming && len(r.Players) >= 2 {
					r.loop = r.loopStart
				} else {

				}
			},
			r.Loop,
		)
	}
}
