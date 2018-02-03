package hall

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka-cow2/database"
	"github.com/liuhan907/waka/waka-cow2/modules/hall/tools"
)

var (
	log = logrus.WithFields(logrus.Fields{
		"pid":    os.Getpid(),
		"module": "cow2",
	})
	pid *actor.PID
)

type actorT struct {
	supervisor *actor.PID
	pid        *actor.PID

	players playerMap

	cowRooms      cowRoomMapT
	cowNumberPool *tools.NumberPool
}

func (my *actorT) Receive(context actor.Context) {
	defer func() {
		val := recover()
		if val != nil {
			stack := debug.Stack()
			fmt.Println(string(stack))
			log.WithFields(logrus.Fields{
				"recover": val,
				"trace":   string(stack),
			}).Errorln("panic!")
		}
	}()

	if my.ReceiveActor(context) {
		return
	}
	if my.ReceiveClock(context) {
		return
	}
	if my.ReceiveSupervisor(context) {
		return
	}
}

func Spawn(supervisor *actor.PID) *actor.PID {
	instance := &actorT{
		supervisor:    supervisor,
		players:       make(playerMap, 12800),
		cowRooms:      make(cowRoomMapT, 12800),
		cowNumberPool: tools.NewNumberPool(10001, 89999, true),
	}
	instance.players[database.Player(0)] = &playerT{}
	return actor.Spawn(
		actor.FromInstance(instance),
	)
}
