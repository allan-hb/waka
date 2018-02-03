package player

import (
	"os"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka-four/database"
)

var (
	log = logrus.WithFields(logrus.Fields{
		"pid":    os.Getpid(),
		"module": "four.player",
	})
)

type actorT struct {
	hall   *actor.PID
	remote string
	conn   *actor.PID

	log *logrus.Entry
	pid *actor.PID

	player database.Player
}

func (my *actorT) Receive(context actor.Context) {
	if my.ReceiveActor(context) {
		return
	}
	if my.ReceiveSession(context) {
		return
	}
	if my.ReceiveSupervisor(context) {
		return
	}
}

func Spawn(hall *actor.PID, remote string, conn *actor.PID) *actor.PID {
	return actor.Spawn(
		actor.FromInstance(
			&actorT{
				hall:   hall,
				remote: remote,
				conn:   conn,
			},
		),
	)
}
