package hall

func buildTick(number *int32, sender func(int32), completed func(), loop func()) func() {
	return func() {
		sender(*number)
		if *number == 0 {
			completed()
			loop()
		} else {
			*number--
		}
	}
}

func buildTickNumber(number int32, sender func(int32), completed func(), loop func()) func() {
	val := new(int32)
	*val = number
	return buildTick(val, sender, completed, loop)
}
