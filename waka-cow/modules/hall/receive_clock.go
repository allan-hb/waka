package hall

import (
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"
)

func (my *actorT) ReceiveClock(context actor.Context) bool {
	switch context.Message().(type) {
	case *clock1:
		my.clock1()
	case *clock3:
		my.clock3()
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

	my.cowClock1()
	my.redBagClock()
	my.lever28BagClock()
	my.gomokuClock()
}

// ---------------------------------------------------------------------------------------------------------------------

type clock3 struct{}

func (my *actorT) clock3() {
	defer func() {
		time.AfterFunc(time.Second*3, func() { my.pid.Tell(&clock3{}) })
	}()

	my.cowClock3()
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) startClock() {
	my.clock1()
	my.clock3()
}
