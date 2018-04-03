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

	CreateRoom(hall *actorT, id int32, roomType cow_proto.NiuniuRoomType, option *cow_proto.NiuniuRoomOption, creator database.Player)
	JoinRoom(player *playerT)
	LeaveRoom(player *playerT)
	SwitchReady(player *playerT)
	Dismiss(player *playerT)
	Start(player *playerT)
	SpecifyBanker(player *playerT, banker database.Player)
	Grab(player *playerT, grab bool)
	SpecifyRate(player *playerT, rate int32)
	CommitPokers(player *playerT, pokers []string)
	ContinueWith(player *playerT)

	CreateMoney() int32
	JoinMoney() int32
	StartMoney() int32

	GetType() cow_proto.NiuniuRoomType
	GetId() int32
	GetOption() *cow_proto.NiuniuRoomOption
	GetCreator() database.Player
	GetOwner() database.Player
	GetGaming() bool
	GetRoundNumber() int32
	GetBanker() database.Player
	GetPlayers() []database.Player

	NiuniuRoomData() *cow_proto.NiuniuRoomData
	NiuniuRoundStatus(player database.Player) *cow_proto.NiuniuRoundStatus
	NiuniuRequireGrabShow() *cow_proto.NiuniuRequireGrabShow
	NiuniuRoundClear() *cow_proto.NiuniuRoundClear
	NiuniuGameFinally() *cow_proto.NiuniuGameFinally
}

type cowRoomMapT map[int32]cowRoom

func (cows cowRoomMapT) NiuniuRoomData() []*cow_proto.NiuniuRoomData {
	var d []*cow_proto.NiuniuRoomData
	for _, r := range cows {
		d = append(d, r.NiuniuRoomData())
	}
	return d
}

// ---------------------------------------------------------------------------------------------------------------------

func (cows cowRoomMapT) WherePlayer() cowRoomMapT {
	d := make(cowRoomMapT, len(cows))
	for _, r := range cows {
		if r.GetType() != cow_proto.NiuniuRoomType_Flowing {
			d[r.GetId()] = r
		}
	}
	return d
}

func (cows cowRoomMapT) WhereOrder() cowRoomMapT {
	d := make(cowRoomMapT, len(cows))
	for _, r := range cows {
		if r.GetType() == cow_proto.NiuniuRoomType_Order {
			d[r.GetId()] = r
		}
	}
	return d
}

func (cows cowRoomMapT) WherePayForAnother() cowRoomMapT {
	d := make(cowRoomMapT, len(cows))
	for _, r := range cows {
		if r.GetType() == cow_proto.NiuniuRoomType_PayForAnother {
			d[r.GetId()] = r
		}
	}
	return d
}

func (cows cowRoomMapT) WhereFlowing() cowRoomMapT {
	d := make(cowRoomMapT, len(cows))
	for _, r := range cows {
		if r.GetType() == cow_proto.NiuniuRoomType_Flowing {
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

func (cows cowRoomMapT) WhereCreator(creator database.Player) cowRoomMapT {
	d := make(cowRoomMapT, len(cows))
	for _, r := range cows {
		if r.GetCreator() == creator {
			d[r.GetId()] = r
		}
	}
	return d
}

func (cows cowRoomMapT) WhereIdle() cowRoomMapT {
	d := make(cowRoomMapT, len(cows))
	for _, r := range cows {
		if len(r.GetPlayers()) == 0 {
			d[r.GetId()] = r
		}
	}
	return d
}

func (cows cowRoomMapT) WhereReady() cowRoomMapT {
	d := make(cowRoomMapT, len(cows))
	for _, r := range cows {
		if !r.GetGaming() && len(r.GetPlayers()) < 5 {
			d[r.GetId()] = r
		}
	}
	return d
}
