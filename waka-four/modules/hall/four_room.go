package hall

import (
	"github.com/liuhan907/waka/waka-four/database"
	"github.com/liuhan907/waka/waka-four/proto"
)

const (
	kVoteSecond = 300
)

type fourRoomT interface {
	CreateDiamonds() int32
	EnterDiamonds() int32
	LeaveDiamonds(database.Player) int32
	CostDiamonds() int32

	GetId() int32
	GetOption() *four_proto.FourRoomOption
	GetCreator() database.Player
	GetOwner() database.Player
	GetGaming() bool
	GetRoundNumber() int32

	GetPlayers() []database.Player

	FourRoom1() *four_proto.FourRoom1
	FourRoom2() *four_proto.FourRoom2
	FourRoundStatus() *four_proto.FourRoundStatus
	FourCompare() *four_proto.FourCompare
	FourSettle() *four_proto.FourSettle
	FourFinallySettle() *four_proto.FourFinallySettle
	FourUpdateDismissVoteStatus() (*four_proto.FourUpdateDismissVoteStatus, bool, bool)
	FourUpdateContinueWithStatus() *four_proto.FourUpdateContinueWithStatus
	FourGrabAnimation() *four_proto.FourGrabAnimation

	BackendRoom() map[string]interface{}

	Left(player *playerT)
	Recover(player *playerT)
	CreateRoom(hall *actorT, id int32, option *four_proto.FourRoomOption, player database.Player) fourRoomT
	JoinRoom(player *playerT)
	LeaveRoom(player *playerT)
	SwitchReady(player *playerT)
	Dismiss(player *playerT)
	DismissVote(player *playerT, passing bool)
	Start(player *playerT)
	Cut(player *playerT, pos int32)
	CommitPokers(player *playerT, front, behind []string)
	ContinueWith(player *playerT)
	SendMessage(player *playerT, messageType int32, text string)

	Loop()
	Tick()
}

type fourRoomMapT map[int32]fourRoomT

func (cows fourRoomMapT) FourRoom1() []*four_proto.FourRoom1 {
	var d []*four_proto.FourRoom1
	for _, r := range cows {
		d = append(d, r.FourRoom1())
	}
	return d
}

func (cows fourRoomMapT) FourRoom2() []*four_proto.FourRoom2 {
	var d []*four_proto.FourRoom2
	for _, r := range cows {
		d = append(d, r.FourRoom2())
	}
	return d
}

// ---------------------------------------------------------------------------------------------------------------------

func (fours fourRoomMapT) WhereOrder() fourRoomMapT {
	d := make(fourRoomMapT, len(fours))
	for _, r := range fours {
		if r.GetOption().PayMode == 1 || r.GetOption().PayMode == 2 {
			d[r.GetId()] = r
		}
	}
	return d
}

func (fours fourRoomMapT) WherePayForAnother() fourRoomMapT {
	d := make(fourRoomMapT, len(fours))
	for _, r := range fours {
		if r.GetOption().PayMode == 3 {
			d[r.GetId()] = r
		}
	}
	return d
}

func (fours fourRoomMapT) WhereCreator(player database.Player) fourRoomMapT {
	d := make(fourRoomMapT, len(fours))
	for _, r := range fours {
		if r.GetCreator() == player {
			d[r.GetId()] = r
		}
	}
	return d
}
