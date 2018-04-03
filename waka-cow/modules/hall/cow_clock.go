package hall

import (
	"github.com/liuhan907/waka/waka-cow/database"
	"github.com/liuhan907/waka/waka-cow/proto"
	"github.com/sirupsen/logrus"
	"gopkg.in/ahmetb/go-linq.v3"
)

type flowingMode struct {
	Mode  int32
	Score int32
}

var (
	flowingModes = []flowingMode{
		{0, 2}, {0, 5}, {0, 10}, {0, 20}, {0, 50}, {0, 100}, {0, 200}, {0, 500},
		{1, 2}, {1, 5}, {1, 10}, {1, 20}, {1, 50}, {1, 100}, {1, 200}, {1, 500},
	}
)

func (my *actorT) cowClock1() {
	for _, room := range my.cowRooms {
		room.Tick()
	}
}

func (my *actorT) cowClock3() {
	for _, mode := range flowingModes {
		r1 := my.cowRooms.
			WhereFlowing().
			WhereScore(mode.Score).
			WhereMode(mode.Mode).
			WhereReady()

		if len(r1) == 0 {
			id, ok := my.cowSupervisorNumberPool.Acquire()
			if ok {
				r := new(supervisorRoomT)
				r.CreateRoom(
					my,
					id,
					cow_proto.NiuniuRoomType_Flowing,
					&cow_proto.NiuniuRoomOption{
						Banker: 2,
						Mode:   mode.Mode,
						Score:  mode.Score,
					},
					database.DefaultSupervisor,
				)
				my.cowRooms[id] = r
				log.WithFields(logrus.Fields{
					"score":   mode.Mode,
					"mode":    mode.Score,
					"room_id": id,
				}).Debugln("flowing room created")
			}
		} else {
			r2 := r1.WhereIdle()
			r3 := r1.WhereReady()

			if len(r3) > 0 && len(r2) > 0 {
				linq.From(r2).Except(linq.From(r3)).ForEachT(func(in cowRoom) {
					delete(my.cowRooms, in.GetId())
				})
			}
		}
	}
}
