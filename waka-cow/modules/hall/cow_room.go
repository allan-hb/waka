package hall

import (
	"github.com/liuhan907/waka/waka-cow/database"
	"github.com/liuhan907/waka/waka-cow/proto"
)

type cowRoom interface {
	Loop()
	Tick()

	Left(player *playerT)
	Recover(player *playerT)

	CreateRoom(hall *actorT, id int32, option *waka.NiuniuRoomOption, creator database.Player) cowRoom
	JoinRoom(player *playerT)
	LeaveRoom(player *playerT)
	SwitchReady(player *playerT)
	SwitchRole(player *playerT)
	Dismiss(player *playerT)
	Start(player *playerT)
	SpecifyBanker(player *playerT, banker database.Player)
	Grab(player *playerT, grab bool)
	SpecifyRate(player *playerT, rate int32)
	CommitPokers(player *playerT, pokers []string)
	ContinueWith(player *playerT)

	CreateMoney() int32
	EnterMoney() int32
	LeaveMoney() int32
	CostMoney() int32

	GetType() waka.NiuniuRoomType
	GetId() int32
	GetOption() *waka.NiuniuRoomOption
	GetCreator() database.Player
	GetOwner() database.Player
	GetGaming() bool
	GetRoundNumber() int32
	GetBanker() database.Player

	GetPlayers() []database.Player
	GetObservers() []database.Player

	NiuniuRoomData1() *waka.NiuniuRoomData1
	NiuniuRoomData2() *waka.NiuniuRoomData2
	NiuniuRoundStatus(player database.Player) *waka.NiuniuRoundStatus
	NiuniuGrabAnimation() *waka.NiuniuGrabAnimation
	NiuniuRoundClear() *waka.NiuniuRoundClear
	NiuniuRoundFinally() *waka.NiuniuRoundFinally
}

type cowRoomMapT map[int32]cowRoom

func (cows cowRoomMapT) NiuniuRoomData1() []*waka.NiuniuRoomData1 {
	var d []*waka.NiuniuRoomData1
	for _, r := range cows {
		d = append(d, r.NiuniuRoomData1())
	}
	return d
}

func (cows cowRoomMapT) NiuniuRoomData2() []*waka.NiuniuRoomData2 {
	var d []*waka.NiuniuRoomData2
	for _, r := range cows {
		d = append(d, r.NiuniuRoomData2())
	}
	return d
}

// ---------------------------------------------------------------------------------------------------------------------

func (cows cowRoomMapT) WherePlayer() cowRoomMapT {
	d := make(cowRoomMapT, len(cows))
	for _, r := range cows {
		if r.GetType() != waka.NiuniuRoomType_Agent {
			d[r.GetId()] = r
		}
	}
	return d
}

func (cows cowRoomMapT) WhereOrder() cowRoomMapT {
	d := make(cowRoomMapT, len(cows))
	for _, r := range cows {
		if r.GetType() == waka.NiuniuRoomType_Order {
			d[r.GetId()] = r
		}
	}
	return d
}

func (cows cowRoomMapT) WherePayForAnother() cowRoomMapT {
	d := make(cowRoomMapT, len(cows))
	for _, r := range cows {
		if r.GetType() == waka.NiuniuRoomType_PayForAnother {
			d[r.GetId()] = r
		}
	}
	return d
}

func (cows cowRoomMapT) WhereSupervisor() cowRoomMapT {
	d := make(cowRoomMapT, len(cows))
	for _, r := range cows {
		if r.GetType() == waka.NiuniuRoomType_Agent {
			d[r.GetId()] = r
		}
	}
	return d
}

func (cows cowRoomMapT) WhereIdle() cowRoomMapT {
	d := make(cowRoomMapT, len(cows))
	for _, r := range cows {
		if len(r.GetPlayers())+len(r.GetObservers()) == 0 {
			d[r.GetId()] = r
		}
	}
	return d
}

func (cows cowRoomMapT) WhereCreator(player database.Player) cowRoomMapT {
	d := make(cowRoomMapT, len(cows))
	for _, r := range cows {
		if r.GetCreator() == player {
			d[r.GetId()] = r
		}
	}
	return d
}

func (cows cowRoomMapT) WhereScore(score int32) cowRoomMapT {
	d := make(cowRoomMapT, len(cows))
	for _, r := range cows {
		if r.GetOption().GetScore() == score {
			d[r.GetId()] = r
		}
	}
	return d
}

func (cows cowRoomMapT) WhereMode(mode int32) cowRoomMapT {
	d := make(cowRoomMapT, len(cows))
	for _, r := range cows {
		if r.GetOption().GetMode() == mode {
			d[r.GetId()] = r
		}
	}
	return d
}
