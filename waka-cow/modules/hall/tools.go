package hall

import "time"

func buildTickDeadline(deadline int64, sender func(int64), completed func(), loop func()) func() {
	return func() {
		sender(deadline)
		if time.Now().Unix() >= deadline {
			completed()
			loop()
		}
	}
}

func buildTickAfter(after int32, starter func(int64), sender func(int64), completed func(), loop func()) func() {
	deadline := time.Now().Unix() + int64(after)
	starter(deadline)
	return buildTickDeadline(deadline, sender, completed, loop)
}
