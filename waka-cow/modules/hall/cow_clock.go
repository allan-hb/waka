package hall

import (
	"github.com/liuhan907/waka/waka-cow/database"
	"github.com/liuhan907/waka/waka-cow/proto"
	"github.com/sirupsen/logrus"
)

func (my *actorT) cowClock1() {
	for _, room := range my.cowRooms {
		room.Tick()
	}
}

func (my *actorT) cowClock3() {
	for _, r := range my.cowRooms.WhereSupervisor() {
		if len(r.GetPlayers())+len(r.GetObservers()) != 0 {
			my.cowIdleRooms[r.GetId()] = 0
		} else {
			c := my.cowIdleRooms[r.GetId()]
			c++
			my.cowIdleRooms[r.GetId()] = c
		}

		if my.cowIdleRooms[r.GetId()] >= 10 {
			idleRooms := my.cowRooms.
				WhereSupervisor().
				WhereCreator(r.GetCreator()).
				WhereScore(r.GetOption().GetScore()).
				WhereMode(r.GetOption().GetMode()).
				WhereIdle()
			if len(idleRooms) > 1 {
				delete(my.cowRooms, r.GetId())
				delete(my.cowIdleRooms, r.GetId())

				log.WithFields(logrus.Fields{
					"creator": r.GetCreator(),
					"mode":    r.GetOption().GetMode(),
					"score":   r.GetOption().GetScore(),
					"id":      r.GetId(),
				}).Debugln("supervisor room removed")
			} else {
				my.cowIdleRooms[r.GetId()] = 0
			}
		}
	}
}

func (my *actorT) cowClock30() {
	supervisors, err := database.QuerySupervisorList()
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Warnln("query supervisor list failed")
		return
	}

	for _, supervisor := range supervisors {
		for _, score := range supervisor.BaseScores {
			rooms := my.cowRooms.
				WhereSupervisor().
				WhereCreator(supervisor.Player).
				WhereScore(score).
				WhereMode(0)
			if len(rooms) == 0 {
				id, ok := my.cowSupervisorNumberPool.Acquire()
				if ok {
					my.cowRooms[id] = new(supervisorRoomT).CreateRoom(my, id, &waka.NiuniuRoomOption{
						Banker: 2,
						Mode:   0,
						Score:  score,
					}, supervisor.Player)

					log.WithFields(logrus.Fields{
						"supervisor": supervisor.Ref,
						"player":     supervisor.Player,
						"score":      score,
						"mode":       0,
						"room_id":    id,
					}).Debugln("supervisor room created")
				}
			}

			rooms = my.cowRooms.
				WhereSupervisor().
				WhereCreator(supervisor.Player).
				WhereScore(score).
				WhereMode(1)
			if len(rooms) == 0 {
				id, ok := my.cowSupervisorNumberPool.Acquire()
				if ok {
					my.cowRooms[id] = new(supervisorRoomT).CreateRoom(my, id, &waka.NiuniuRoomOption{
						Banker: 2,
						Mode:   1,
						Score:  score,
					}, supervisor.Player)

					log.WithFields(logrus.Fields{
						"supervisor": supervisor.Ref,
						"player":     supervisor.Player,
						"score":      score,
						"mode":       1,
						"room_id":    id,
					}).Debugln("supervisor room created")
				}
			}
		}
	}
}
