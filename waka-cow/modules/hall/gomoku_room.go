package hall

import (
	"github.com/liuhan907/waka/waka-cow/database"
	"github.com/liuhan907/waka/waka-cow/modules/hall/tools/gomoku"
	waka "github.com/liuhan907/waka/waka-cow/proto"
	"github.com/sirupsen/logrus"
)

const (
	kGomokuTime = 300
)

type gomokuRoomPlayerT struct {
	Room *gomokuRoomT

	Color      gomoku.PieceType
	Player     database.Player
	RemainTime int32

	Play bool

	Sent bool
}

type gomokuRoomT struct {
	Hall *actorT

	Id      int32
	Creator *gomokuRoomPlayerT
	Student *gomokuRoomPlayerT
	Cost    int32

	Gaming bool

	Board         gomoku.Board
	RoundNumber   int32
	ThisPlayer    *gomokuRoomPlayerT
	AnotherPlayer *gomokuRoomPlayerT

	Loop func() bool
	Tick func()
}

func (r *gomokuRoomT) GomokuRoom() *waka.GomokuRoom {
	pb := &waka.GomokuRoom{
		Id:   r.Id,
		Cost: r.Cost,
	}
	if r.Creator != nil {
		pb.Creator = r.Hall.ToPlayer(r.Creator.Player)
	}
	if r.Student != nil {
		pb.Student = r.Hall.ToPlayer(r.Student.Player)
	}
	return pb
}

type gomokuRoomMapT map[int32]*gomokuRoomT

// ---------------------------------------------------------------------------------------------------------------------

func (r *gomokuRoomT) Left(player *playerT) {
	if !r.Gaming {
		if player.Player == r.Creator.Player {
		} else {
			player.InsideGomoku = 0
			r.Hall.sendGomokuLeft(player.Player)

			r.Student = nil
		}
	}
	r.Hall.sendGomokuUpdateRoomForAll(r)
}

func (r *gomokuRoomT) Recover(player *playerT) {
	if r.Creator.Player == player.Player {
		r.Creator.Sent = false
	} else {
		r.Student.Sent = false
	}

	r.Hall.sendGomokuUpdateRoomForAll(r)
	if r.Gaming {
		r.Hall.sendGomokuUpdateRound(player.Player, r)
		r.Loop()
	}
}

func (r *gomokuRoomT) Create(hall *actorT, player *playerT, id int32) {
	*r = gomokuRoomT{
		Hall:  hall,
		Id:    id,
		Board: gomoku.NewBoard(),
	}

	r.Creator = &gomokuRoomPlayerT{
		Room:       r,
		Color:      gomoku.PieceType_Black,
		Player:     player.Player,
		RemainTime: kGomokuTime,
	}

	player.InsideGomoku = id

	hall.gomokuRooms[id] = r

	hall.sendGomokuRoomCreated(player.Player, id)
	hall.sendGomokuUpdateRoomForAll(r)
}

func (r *gomokuRoomT) Join(player *playerT) {
	r.Student = &gomokuRoomPlayerT{
		Room:       r,
		Color:      gomoku.PieceType_White,
		Player:     player.Player,
		RemainTime: kGomokuTime,
	}

	player.InsideGomoku = r.Id

	r.Hall.sendGomokuRoomEntered(player.Player)
	r.Hall.sendGomokuUpdateRoomForAll(r)
}

func (r *gomokuRoomT) SetCost(player *playerT, cost int32) {
	r.Cost = cost
	r.Hall.sendGomokuUpdateRoomForAll(r)
}

func (r *gomokuRoomT) Leave(player *playerT) {
	if player.Player == r.Creator.Player {
		r.Hall.sendGomokuLeftForAll(r)
		delete(r.Hall.gomokuRooms, r.Id)
		r.Hall.gomokuNumberPool.Return(r.Id)
		if r.Creator != nil {
			if player := r.Hall.players[r.Creator.Player]; player != nil {
				player.InsideGomoku = 0
			}
		}
		if r.Student != nil {
			if player := r.Hall.players[r.Student.Player]; player != nil {
				player.InsideGomoku = 0
			}
		}
	} else {
		player.InsideGomoku = 0
		r.Hall.sendGomokuLeft(player.Player)

		r.Student = nil
		r.Hall.sendGomokuUpdateRoomForAll(r)
	}
}

func (r *gomokuRoomT) Dismiss(player *playerT) {
	r.Hall.sendGomokuLeftByDismissForAll(r)
	delete(r.Hall.gomokuRooms, r.Id)
	r.Hall.gomokuNumberPool.Return(r.Id)
	if r.Creator != nil {
		if player := r.Hall.players[r.Creator.Player]; player != nil {
			player.InsideGomoku = 0
		}
	}
	if r.Student != nil {
		if player := r.Hall.players[r.Student.Player]; player != nil {
			player.InsideGomoku = 0
		}
	}
}

func (r *gomokuRoomT) Start(player *playerT) {
	if r.Creator == nil {
		return
	}
	if r.Student == nil {
		return
	}

	r.Loop = r.loopStart

	r.loop()
}

func (r *gomokuRoomT) Play(player *playerT, x, y int32) {
	if player.Player != r.ThisPlayer.Player {
		return
	}

	if r.Board.Lookup(x, y) != gomoku.PieceType_None {
		return
	}

	r.Board.Play(x, y, r.ThisPlayer.Color)

	r.ThisPlayer.Play = true

	r.loop()
}

func (r *gomokuRoomT) Surrender(player *playerT) {
	r.Tick = nil
	r.Loop = r.loopSettle

	if player.Player == r.Creator.Player {
		r.ThisPlayer = r.Student
		r.AnotherPlayer = r.Creator
	} else {
		r.ThisPlayer = r.Creator
		r.AnotherPlayer = r.Student
	}

	r.loop()
}

// ---------------------------------------------------------------------------------------------------------------------

func (r *gomokuRoomT) loop() {
	for r.Loop() {
	}
}

func (r *gomokuRoomT) loopStart() bool {
	r.Gaming = true
	r.RoundNumber = 1
	r.ThisPlayer = r.Creator
	r.AnotherPlayer = r.Student

	r.Hall.sendGomokuStartedForAll(r)

	r.Loop = r.loopPlay

	return true
}

func (r *gomokuRoomT) loopPlay() bool {
	r.Hall.sendGomokuUpdateRoundForAll(r)

	r.ThisPlayer.Sent = false
	r.AnotherPlayer.Sent = false
	r.ThisPlayer.Play = false

	r.Loop = r.loopPlayContinue
	r.Tick = buildTick(
		&r.ThisPlayer.RemainTime,
		func(number int32) {
			r.Hall.sendGomokuUpdatePlayCountdownForAll(r, number)
		},
		func() {
			r.Tick = nil
			r.ThisPlayer, r.AnotherPlayer = r.AnotherPlayer, r.ThisPlayer
			r.Loop = r.loopSettle
		},
		r.loop,
	)

	return true
}

func (r *gomokuRoomT) loopPlayContinue() bool {
	finally := true
	if !r.ThisPlayer.Sent {
		r.Hall.sendGomokuRequirePlay(r.ThisPlayer.Player, true)
		r.ThisPlayer.Sent = true
		finally = false
	}
	if !r.AnotherPlayer.Sent {
		r.Hall.sendGomokuRequirePlay(r.AnotherPlayer.Player, false)
		r.AnotherPlayer.Sent = true
		finally = false
	}
	if !r.ThisPlayer.Play {
		finally = false
	}

	if !finally {
		return false
	}

	r.Tick = nil

	victory, finally := r.Board.Judge()
	if !finally {
		r.Loop = r.loopSwitch
	} else {
		if victory == "black" {
			r.ThisPlayer = r.Creator
			r.AnotherPlayer = r.Student
		} else {
			r.ThisPlayer = r.Student
			r.AnotherPlayer = r.Creator
		}

		r.Loop = r.loopSettle
	}

	return true
}

func (r *gomokuRoomT) loopSwitch() bool {
	r.ThisPlayer, r.AnotherPlayer = r.AnotherPlayer, r.ThisPlayer
	r.RoundNumber++
	r.Loop = r.loopPlay

	return true
}

func (r *gomokuRoomT) loopSettle() bool {
	r.Hall.sendGomokuUpdateRoundForAll(r)

	err := database.GomokuSettle(r.ThisPlayer.Player, r.AnotherPlayer.Player, r.Cost*100)
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Warnln("gomoku settle failed")
	}

	err = database.GomokuAddWarHistory(r.ThisPlayer.Player, r.AnotherPlayer.Player, r.Cost*100)
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Warnln("gomoku add history failed")
	}

	r.Hall.sendGomokuVictory(r.ThisPlayer.Player)
	r.Hall.sendGomokuLost(r.AnotherPlayer.Player)
	delete(r.Hall.gomokuRooms, r.Id)
	r.Hall.gomokuNumberPool.Return(r.Id)

	if player := r.Hall.players[r.ThisPlayer.Player]; player != nil {
		player.InsideGomoku = 0
	}
	if player := r.Hall.players[r.AnotherPlayer.Player]; player != nil {
		player.InsideGomoku = 0
	}

	r.Hall.sendPlayerSecret(r.ThisPlayer.Player)
	r.Hall.sendPlayerSecret(r.AnotherPlayer.Player)

	return false
}
