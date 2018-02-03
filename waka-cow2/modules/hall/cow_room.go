package hall

import (
	"github.com/liuhan907/waka/waka-cow2/database"
	"github.com/liuhan907/waka/waka-cow2/proto"
)

type cowRoomT interface {
	Loop()
	Tick()

	Left(player *playerT)
	Recover(player *playerT)

	CreateRoom(hall *actorT, id int32, option *cow_proto.NiuniuRoomOption, creator database.Player) cowRoomT
	JoinRoom(player *playerT)
	LeaveRoom(player *playerT)
	SwitchReady(player *playerT)
	Dismiss(player *playerT)
	KickPlayer(player *playerT, target database.Player)
	Start(player *playerT)
	SpecifyBanker(player *playerT, banker database.Player)
	Grab(player *playerT, grab bool)
	SpecifyRate(player *playerT, rate int32)
	ContinueWith(player *playerT)

	CreateDiamonds() int32
	EnterDiamonds() int32
	CostDiamonds() int32

	GetId() int32
	GetOption() *cow_proto.NiuniuRoomOption
	GetCreator() database.Player
	GetOwner() database.Player
	GetGaming() bool
	GetRoundNumber() int32
	GetBanker() database.Player

	GetPlayers() []database.Player

	NiuniuRoomData1() *cow_proto.NiuniuRoomData1
	NiuniuRoundStatus(player database.Player) *cow_proto.NiuniuRoundStatus
	NiuniuGrabAnimation() *cow_proto.NiuniuGrabAnimation
	NiuniuRoundClear() *cow_proto.NiuniuRoundClear
	NiuniuRoundFinally() *cow_proto.NiuniuRoundFinally
}

type cowRoomMapT map[int32]cowRoomT

func (cows cowRoomMapT) NiuniuRoomData1() []*cow_proto.NiuniuRoomData1 {
	var d []*cow_proto.NiuniuRoomData1
	for _, r := range cows {
		d = append(d, r.NiuniuRoomData1())
	}
	return d
}

// ---------------------------------------------------------------------------------------------------------------------

func (cows cowRoomMapT) WherePayForAnother() cowRoomMapT {
	d := make(cowRoomMapT, len(cows))
	for _, r := range cows {
		if r.GetOption().GetPayMode() == 1 {
			d[r.GetId()] = r
		}
	}
	return d
}

func (cows cowRoomMapT) WhereOrder() cowRoomMapT {
	d := make(cowRoomMapT, len(cows))
	for _, r := range cows {
		if r.GetOption().GetPayMode() == 2 {
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
