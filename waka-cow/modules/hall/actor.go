package hall

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka-cow/database"
	"github.com/liuhan907/waka/waka-cow/modules/hall/tools"
)

var (
	log = logrus.WithFields(logrus.Fields{
		"pid":    os.Getpid(),
		"module": "cow",
	})
	pid *actor.PID
)

type actorT struct {
	supervisor *actor.PID
	pid        *actor.PID

	players playerMap

	cowRooms                cowRoomMapT
	cowIdleRooms            map[int32]int32
	cowPlayerNumberPool     *tools.NumberPool
	cowSupervisorNumberPool *tools.NumberPool

	redIdPool int32
	redBags   redBagMapT

	lever28IdPool int32
	lever28Bags   lever28BagMapT

	gomokuRooms      gomokuRoomMapT
	gomokuNumberPool *tools.NumberPool
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
	if my.ReceiveBackend(context) {
		return
	}
}

func Spawn(supervisor *actor.PID) *actor.PID {
	instance := &actorT{
		supervisor:              supervisor,
		players:                 make(playerMap, 12800),
		cowRooms:                make(cowRoomMapT, 12800),
		cowIdleRooms:            make(map[int32]int32, 12800),
		cowPlayerNumberPool:     tools.NewNumberPool(10001, 89999, true),
		cowSupervisorNumberPool: tools.NewNumberPool(100001, 899999, true),
		redBags:                 make(redBagMapT, 12800),
		lever28Bags:             make(lever28BagMapT, 12800),
		gomokuRooms:             make(gomokuRoomMapT, 12800),
		gomokuNumberPool:        tools.NewNumberPool(10001, 89999, true),
	}
	instance.players[database.Player(0)] = &playerT{}
	return actor.Spawn(
		actor.FromInstance(instance),
	)
}
