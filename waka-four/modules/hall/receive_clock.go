package hall

import (
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (my *actorT) ReceiveClock(context actor.Context) bool {
	switch context.Message().(type) {
	case *clock1:
		my.clock1()
	default:
		return false
	}
	return true
}

// ---------------------------------------------------------------------------------------------------------------------

type clock1 struct{}

func (my *actorT) clock1() {
	defer func() {
		time.AfterFunc(time.Second, func() { my.pid.Tell(&clock1{}) })
	}()

	for _, room := range my.fourRooms {
		room.Tick()
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) startClock() {
	my.clock1()
}
