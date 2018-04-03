package hall

func (r *supervisorRoomT) buildStart() {
	if r.tick == nil && len(r.Players) >= 2 {
		r.tick = buildTickAfter(
			5,
			func(deadline int64) {
				r.Hall.sendNiuniuDeadlineForAll(r, deadline)
			},
			func(number int64) {
				if r.Gaming || len(r.Players) < 2 {
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
