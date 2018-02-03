package session

import (
	"time"

	"github.com/AsynkronIT/protoactor-go/actor"

	"github.com/liuhan907/waka/waka/proto"
)

func (my *actorT) ReceiveClock(context actor.Context) bool {
	switch context.Message().(type) {
	case *dead:
		my.dead()
	case *sender:
		my.sender()
	default:
		return false
	}
	return true
}

// ---------------------------------------------------------------------------------------------------------------------

type dead struct{}

type sender struct{}

func (my *actorT) dead() {
	if my.option.EnableHeartLog {
		log.Debugln("heartbeat dead checkup")
	}

	if time.Now().Sub(my.heart) >= my.option.HeartDeadPeriod {
		my.conn.Close()
	} else {
		my.startHeartbeatDead()
	}
}

func (my *actorT) sender() {
	if my.option.EnableHeartLog {
		log.Debugln("heartbeat send")
	}

	my.conn.Send(&waka_proto.Heart{})

	my.startHeartbeatSender()
}

// ---------------------------------------------------------------------------------------------------------------------

func (my *actorT) startHeartbeatDead() {
	time.AfterFunc(my.option.HeartDeadPeriod, func() { my.pid.Tell(&dead{}) })
}

func (my *actorT) startHeartbeatSender() {
	time.AfterFunc(my.option.HeartPeriod, func() { my.pid.Tell(&sender{}) })
}

func (my *actorT) startHeartbeat() {
	my.startHeartbeatDead()
	my.startHeartbeatSender()
}
