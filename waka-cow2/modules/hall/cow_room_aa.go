package hall

import (
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
	linq "gopkg.in/ahmetb/go-linq.v3"

	"github.com/liuhan907/waka/waka-cow2/database"
	"github.com/liuhan907/waka/waka-cow2/modules/hall/tools"
	"github.com/liuhan907/waka/waka-cow2/modules/hall/tools/cow"
	"github.com/liuhan907/waka/waka-cow2/proto"
)

type aaPlayerRoundT struct {
	// 总分
	Points int32
	// 胜利的场数
	VictoriousNumber int32

	// 手牌
	Pokers []string

	// 最佳配牌
	BestPokers []string
	// 最佳牌型权重
	BestWeight int32
	// 最佳牌型
	BestPattern string
	// 最佳牌型倍率
	BestRate int32
	// 最佳得分
	BestPoints int32

	// 是否抢庄
	Grab bool
	// 抢庄已提交
	GrabCommitted bool
	// 倍率
	Rate int32
	// 倍率已提交
	RateCommitted bool

	// 阶段完成已提交
	ContinueWithCommitted bool

	// 本阶段消息是否已发送
	Sent bool
}

type aaPlayerT struct {
	Room *aaRoomT

	Player database.Player
	Pos    int32
	Ready  bool

	Round aaPlayerRoundT
}

func (player *aaPlayerT) NiuniuRoomData1PlayerData() (pb *cow_proto.NiuniuRoomData1_PlayerData) {
	lost := false
	if player, being := player.Room.Hall.players[player.Player]; !being || player.Remote == "" {
		lost = true
	}
	return &cow_proto.NiuniuRoomData1_PlayerData{
		Player: player.Room.Hall.ToPlayer(player.Player),
		Pos:    player.Pos,
		Ready:  player.Ready,
		Lost:   lost,
	}
}

type aaPlayerMapT map[database.Player]*aaPlayerT

func (players aaPlayerMapT) NiuniuRoomData1RoomPlayer() (pb []*cow_proto.NiuniuRoomData1_PlayerData) {
	for _, player := range players {
		pb = append(pb, player.NiuniuRoomData1PlayerData())
	}
	return pb
}

func (players aaPlayerMapT) ToSlice() (d []*aaPlayerT) {
	for _, player := range players {
		d = append(d, player)
	}
	return d
}

// ---------------------------------------------------------------------------------------------------------------------

type aaRoomT struct {
	Hall *actorT

	Id      int32
	Option  *cow_proto.NiuniuRoomOption
	Owner   database.Player
	Players aaPlayerMapT

	loop func() bool
	tick func()

	Seats *tools.NumberPool

	Gaming      bool
	RoundNumber int32
	Step        string
	Banker      database.Player
}

// ---------------------------------------------------------------------------------------------------------------------

func (r *aaRoomT) CreateDiamonds() int32 {
	switch r.Option.GetRoundNumber() {
	case 12:
		return 1
	case 20:
		return 2
	default:
		return math.MaxInt32
	}
}

func (r *aaRoomT) EnterDiamonds() int32 {
	switch r.Option.GetRoundNumber() {
	case 12:
		return 1
	case 20:
		return 2
	default:
		return math.MaxInt32
	}
}

func (r *aaRoomT) CostDiamonds() int32 {
	return r.CreateDiamonds()
}

func (r *aaRoomT) GetId() int32 {
	return r.Id
}

func (r *aaRoomT) GetOption() *cow_proto.NiuniuRoomOption {
	return r.Option
}

func (r *aaRoomT) GetCreator() database.Player {
	return r.Owner
}

func (r *aaRoomT) GetOwner() database.Player {
	return r.Owner
}

func (r *aaRoomT) GetGaming() bool {
	return r.Gaming
}

func (r *aaRoomT) GetRoundNumber() int32 {
	return r.RoundNumber
}

func (r *aaRoomT) GetBanker() database.Player {
	return r.Banker
}

func (r *aaRoomT) GetPlayers() []database.Player {
	var d []database.Player
	linq.From(r.Players).SelectT(func(pair linq.KeyValue) database.Player { return pair.Key.(database.Player) }).ToSlice(&d)
	return d
}

func (r *aaRoomT) NiuniuRoomData1() *cow_proto.NiuniuRoomData1 {
	return &cow_proto.NiuniuRoomData1{
		Id:      r.Id,
		Option:  r.GetOption(),
		Creator: r.Hall.ToPlayer(r.Owner),
		Owner:   r.Hall.ToPlayer(r.Owner),
		Players: r.Players.NiuniuRoomData1RoomPlayer(),
		Gaming:  r.Gaming,
	}
}

func (r *aaRoomT) NiuniuRoundStatus(player database.Player) *cow_proto.NiuniuRoundStatus {
	var pokers []string
	var players []*cow_proto.NiuniuRoundStatus_PlayerData
	for id, playerData := range r.Players {
		players = append(players, &cow_proto.NiuniuRoundStatus_PlayerData{
			Id:            int32(id),
			Points:        playerData.Round.Points,
			GrabCommitted: playerData.Round.GrabCommitted,
			Grab:          playerData.Round.Grab,
			RateCommitted: playerData.Round.RateCommitted,
			Rate:          playerData.Round.Rate,
		})
		if playerData.Player == player {
			if r.Step == "round_clear" || r.Step == "round_finally" {
				pokers = playerData.Round.BestPokers
			} else {
				pokers = playerData.Round.Pokers[:4]
			}
		}
	}

	return &cow_proto.NiuniuRoundStatus{
		Step:        r.Step,
		RoomId:      r.Id,
		RoundNumber: r.RoundNumber,
		Players:     players,
		Banker:      int32(r.Banker),
		Pokers:      pokers,
	}
}

func (r *aaRoomT) NiuniuGrabAnimation() *cow_proto.NiuniuGrabAnimation {
	var players []*cow_proto.NiuniuGrabAnimation_PlayerData
	for _, player := range r.Players {
		players = append(players, &cow_proto.NiuniuGrabAnimation_PlayerData{
			PlayerId: int32(player.Player),
			Grab:     player.Round.Grab,
		})
	}
	return &cow_proto.NiuniuGrabAnimation{
		Players: players,
	}
}

func (r *aaRoomT) NiuniuRoundClear() *cow_proto.NiuniuRoundClear {
	var players []*cow_proto.NiuniuRoundClear_PlayerData
	for _, player := range r.Players {
		players = append(players, &cow_proto.NiuniuRoundClear_PlayerData{
			Player:     r.Hall.ToPlayer(player.Player),
			Points:     player.Round.Points,
			Pokers:     player.Round.BestPokers,
			Type:       player.Round.BestPattern,
			Rate:       player.Round.BestRate,
			ThisPoints: player.Round.BestPoints,
		})
	}
	return &cow_proto.NiuniuRoundClear{Players: players, FinallyAt: time.Now().Format("2006-01-02 15:04:05")}
}

func (r *aaRoomT) NiuniuRoundFinally() *cow_proto.NiuniuRoundFinally {
	var players []*cow_proto.NiuniuRoundFinally_PlayerData
	for _, player := range r.Players {
		players = append(players, &cow_proto.NiuniuRoundFinally_PlayerData{
			Player:    r.Hall.ToPlayer(player.Player),
			Points:    int32(player.Round.Points),
			Victories: player.Round.VictoriousNumber,
		})
	}
	return &cow_proto.NiuniuRoundFinally{Players: players, FinallyAt: time.Now().Format("2006-01-02 15:04:05")}
}

// ---------------------------------------------------------------------------------------------------------------------

func (r *aaRoomT) Loop() {
	for {
		if r.loop == nil {
			return
		}
		if !r.loop() {
			return
		}
	}
}

func (r *aaRoomT) Tick() {
	if r.tick != nil {
		r.tick()
	}
}

func (r *aaRoomT) Left(player *playerT) {
	if !r.Gaming {
		if roomPlayer, being := r.Players[player.Player]; being {
			if player.Player != r.Owner {
				delete(r.Players, player.Player)
				player.InsideCow = 0
				r.Seats.Return(roomPlayer.Pos)
			}

			r.Hall.sendNiuniuUpdateRoomForAll(r)
		}
	}
}

func (r *aaRoomT) Recover(player *playerT) {
	if _, being := r.Players[player.Player]; being {
		r.Players[player.Player].Round.Sent = false
	}

	r.Hall.sendNiuniuUpdateRoomForAll(r)
	if r.Gaming {
		r.Hall.sendNiuniuUpdateRound(player.Player, r)
		r.Loop()
	}
}

func (r *aaRoomT) CreateRoom(hall *actorT, id int32, option *cow_proto.NiuniuRoomOption, creator database.Player) cowRoomT {
	*r = aaRoomT{
		Hall:    hall,
		Id:      id,
		Option:  option,
		Owner:   creator,
		Players: make(aaPlayerMapT, 5),
		Seats:   tools.NewNumberPool(1, 5, false),
	}

	pos, _ := r.Seats.Acquire()

	r.Players[creator] = &aaPlayerT{
		Room:   r,
		Player: creator,
		Pos:    pos,
	}

	if creator.PlayerData().Diamonds < r.CreateDiamonds() {
		r.Hall.sendNiuniuCreateRoomFailed(creator, 1)
		return nil
	} else {
		r.Hall.cowRooms[id] = r
		r.Hall.sendNiuniuCreateRoomSuccess(creator)

		r.Hall.players[creator].InsideCow = id
		r.Hall.sendNiuniuJoinRoomSuccess(creator, r)
		r.Hall.sendNiuniuUpdateRoomForAll(r)
		return r
	}
}

func (r *aaRoomT) JoinRoom(player *playerT) {
	if player.Player.PlayerData().Diamonds < r.EnterDiamonds() {
		r.Hall.sendNiuniuJoinRoomFailed(player.Player, 1)
		return
	}

	_, being := r.Players[player.Player]
	if being {
		r.Hall.sendNiuniuJoinRoomFailed(player.Player, 2)
		return
	}

	if r.Gaming {
		r.Hall.sendNiuniuJoinRoomFailed(player.Player, 4)
		return
	}

	seat, has := r.Seats.Acquire()
	if !has {
		r.Hall.sendNiuniuJoinRoomFailed(player.Player, 0)
		return
	}

	r.Players[player.Player] = &aaPlayerT{
		Room:   r,
		Player: player.Player,
		Pos:    seat,
	}

	player.InsideCow = r.GetId()

	r.Hall.sendNiuniuJoinRoomSuccess(player.Player, r)
	r.Hall.sendNiuniuUpdateRoomForAll(r)
}

func (r *aaRoomT) LeaveRoom(player *playerT) {
	if !r.Gaming {
		if roomPlayer, being := r.Players[player.Player]; being {
			player.InsideCow = 0
			delete(r.Players, player.Player)
			r.Seats.Return(roomPlayer.Pos)

			r.Hall.sendNiuniuLeftRoom(player.Player)

			if r.Owner == player.Player {
				r.Owner = 0
				if len(r.Players) > 0 {
					for _, player := range r.Players {
						r.Owner = player.Player
						break
					}
				}
			}

			if r.Owner == 0 {
				delete(r.Hall.cowRooms, r.Id)
				for _, player := range r.Players {
					r.Hall.players[player.Player].InsideCow = 0
					r.Hall.sendNiuniuLeftRoom(player.Player)
				}
			} else {
				r.Hall.sendNiuniuUpdateRoomForAll(r)
			}
		}
	}
}

func (r *aaRoomT) SwitchReady(player *playerT) {
	if !r.Gaming {
		if roomPlayer, being := r.Players[player.Player]; being {
			roomPlayer.Ready = !roomPlayer.Ready
			r.Hall.sendNiuniuUpdateRoomForAll(r)
		}
	}
}

func (r *aaRoomT) Dismiss(player *playerT) {
	if !r.Gaming {
		if r.Owner == player.Player {
			delete(r.Hall.cowRooms, r.Id)
			for _, player := range r.Players {
				r.Hall.players[player.Player].InsideCow = 0
				r.Hall.sendNiuniuLeftRoomByDismiss(player.Player)
			}
		}
	}
}

func (r *aaRoomT) KickPlayer(player *playerT, target database.Player) {
	if !r.Gaming {
		if r.Owner == player.Player {
			if targetPlayer, being := r.Hall.players[target]; being {
				targetPlayer.InsideCow = 0
			}

			if targetPlayer, being := r.Players[target]; being {
				delete(r.Players, target)
				r.Seats.Return(targetPlayer.Pos)
				r.Hall.sendNiuniuLeftRoom(player.Player)

				if r.Owner == player.Player {
					r.Owner = 0
					if len(r.Players) > 0 {
						for _, player := range r.Players {
							r.Owner = player.Player
							break
						}
					}
				}

				r.Hall.sendNiuniuUpdateRoomForAll(r)
			}
		}
	}
}

func (r *aaRoomT) Start(player *playerT) {
	if !r.Gaming {
		if r.Owner == player.Player {
			started := true
			for _, target := range r.Players {
				if !target.Ready {
					started = false
				}
			}
			if !started {
				return
			}

			var playerRoomCost []*database.CowPlayerRoomCost
			for _, player := range r.Players {
				playerRoomCost = append(playerRoomCost, &database.CowPlayerRoomCost{
					Player: player.Player,
					Number: r.CostDiamonds(),
				})
			}
			err := database.CowOrderSettle(r.Id, playerRoomCost)
			if err != nil {
				log.WithFields(logrus.Fields{
					"room_id": r.Id,
					"option":  r.Option.String(),
					"cost":    playerRoomCost,
					"err":     err,
				}).Warnln("aa room cost settle failed")
				return
			}

			for _, player := range r.Players {
				r.Hall.sendPlayerSecret(player.Player)
			}

			r.loop = r.loopStart

			r.Loop()
		}
	}
}

func (r *aaRoomT) SpecifyBanker(player *playerT, banker database.Player) {
	if r.Gaming {
		if _, being := r.Players[banker]; being {
			r.Banker = banker

			r.Loop()
		}
	}
}

func (r *aaRoomT) Grab(player *playerT, grab bool) {
	if r.Gaming {
		r.Players[player.Player].Round.Grab = grab
		r.Players[player.Player].Round.GrabCommitted = true

		r.Hall.sendNiuniuUpdateRoundForAll(r)

		r.Loop()
	}
}

func (r *aaRoomT) SpecifyRate(player *playerT, rate int32) {
	if r.Gaming {
		r.Players[player.Player].Round.Rate = rate
		r.Players[player.Player].Round.RateCommitted = true

		r.Hall.sendNiuniuUpdateRoundForAll(r)

		r.Loop()
	}
}

func (r *aaRoomT) ContinueWith(player *playerT) {
	if r.Gaming {
		r.Players[player.Player].Round.ContinueWithCommitted = true

		r.Loop()
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func (r *aaRoomT) loopStart() bool {
	r.Gaming = true
	r.RoundNumber = 1

	r.Hall.sendNiuniuStartedForAll(r, r.RoundNumber)

	if r.Option.GetBankerMode() == 0 || r.Option.GetBankerMode() == 1 {
		r.loop = r.loopSpecifyBanker
	} else if r.Option.GetBankerMode() == 2 {
		r.loop = func() bool {
			return r.loopDeal4(r.loopGrab)
		}
	}

	return true
}

func (r *aaRoomT) loopSpecifyBanker() bool {
	r.Step = "require_specify_banker"
	for _, player := range r.Players {
		player.Round.Sent = false
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.loop = r.loopSpecifyBankerContinue
	r.tick = buildTickNumber(
		8,
		func(number int32) {
			r.Hall.sendNiuniuCountdownForAll(r, number)
		},
		func() {
			r.Banker = r.Owner
		},
		r.Loop,
	)

	return true
}

func (r *aaRoomT) loopSpecifyBankerContinue() bool {
	if r.Banker == 0 {
		for _, player := range r.Players {
			if !player.Round.Sent {
				r.Hall.sendNiuniuRequireSpecifyBanker(player.Player, player.Player == r.Owner)
				player.Round.Sent = true
			}
		}

		return false
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.tick = nil
	r.loop = func() bool {
		return r.loopDeal4(r.loopSpecifyRate)
	}

	return true
}

func (r *aaRoomT) loopDeal4(loop func() bool) bool {
	pokers := cow.Acquire5(len(r.Players))
	i := 0
	for _, player := range r.Players {
		player.Round.Pokers = pokers[i]
		player.Round.BestPokers, player.Round.BestWeight, player.Round.BestPattern, player.Round.BestRate, _ =
			cow.SearchBestPokerPattern(player.Round.Pokers, r.Option.GetMode(), r.Option.GetAdditionalPokers() == 1)
		r.Hall.sendNiuniuDeal4(player.Player, player.Round.Pokers[:4])
		i++
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.loop = loop

	return true
}

func (r *aaRoomT) loopGrab() bool {
	r.Step = "require_grab"
	for _, player := range r.Players {
		player.Round.Sent = false
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.loop = r.loopGrabContinue
	r.tick = buildTickNumber(
		8,
		func(number int32) {
			r.Hall.sendNiuniuCountdownForAll(r, number)
		},
		func() {
			for _, player := range r.Players {
				if !player.Round.GrabCommitted {
					player.Round.Grab = false
					player.Round.GrabCommitted = true
				}
			}
		},
		r.Loop,
	)

	return true
}

func (r *aaRoomT) loopGrabContinue() bool {
	finally := true
	for _, player := range r.Players {
		if !player.Round.GrabCommitted {
			finally = false
			if !player.Round.Sent {
				r.Hall.sendNiuniuRequireGrab(player.Player)
				player.Round.Sent = true
			}
		}
	}

	if !finally {
		return false
	}

	r.tick = nil
	r.loop = r.loopGrabAnimation

	return true
}

func (r *aaRoomT) loopGrabAnimation() bool {
	r.Step = "grab_animation"
	for _, player := range r.Players {
		player.Round.Sent = false
		player.Round.ContinueWithCommitted = false
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.loop = r.loopGrabAnimationContinue
	r.tick = buildTickNumber(
		8,
		func(number int32) {
			r.Hall.sendNiuniuCountdownForAll(r, number)
		},
		func() {
			for _, player := range r.Players {
				player.Round.ContinueWithCommitted = true
			}
		},
		r.Loop,
	)

	return true
}

func (r *aaRoomT) loopGrabAnimationContinue() bool {
	finally := true
	for _, player := range r.Players {
		if !player.Round.ContinueWithCommitted {
			finally = false
			if !player.Round.Sent {
				r.Hall.sendNiuniuGrabAnimation(player.Player, r)
				player.Round.Sent = true
			}
		}
	}

	if !finally {
		return false
	}

	r.tick = nil
	r.loop = r.loopGrabSelect

	return true
}

func (r *aaRoomT) loopGrabSelect() bool {
	var candidates []database.Player
	for _, player := range r.Players {
		if player.Round.Grab {
			candidates = append(candidates, player.Player)
		}
	}

	if len(candidates) > 0 {
		r.Banker = candidates[rand.Int()%len(candidates)]

		log.WithFields(logrus.Fields{
			"candidates": candidates,
			"banker":     r.Banker,
		}).Debugln("grab")
	} else {
		r.Banker = r.Owner

		log.WithFields(logrus.Fields{
			"banker": r.Banker,
		}).Debugln("no player grab")
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.loop = r.loopSpecifyRate

	return true
}

func (r *aaRoomT) loopSpecifyRate() bool {
	r.Step = "require_specify_rate"
	for _, player := range r.Players {
		player.Round.Sent = false
		if player.Player == r.Banker {
			player.Round.Rate = 1
			player.Round.RateCommitted = true
		}
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.loop = r.loopSpecifyRateContinue
	r.tick = buildTickNumber(
		5,
		func(number int32) {
			r.Hall.sendNiuniuCountdownForAll(r, number)
		},
		func() {
			for _, player := range r.Players {
				if !player.Round.RateCommitted {
					player.Round.Rate = 1
					player.Round.RateCommitted = true
				}
			}
		},
		r.Loop,
	)

	return true
}

func (r *aaRoomT) loopSpecifyRateContinue() bool {
	finally := true
	for _, player := range r.Players {
		if !player.Round.RateCommitted {
			finally = false
			if !player.Round.Sent {
				r.Hall.sendNiuniuRequireSpecifyRate(player.Player, player.Player != r.Banker)
				player.Round.Sent = true
			}
		}
	}

	if !finally {
		return false
	}

	r.tick = nil
	r.loop = r.loopSettle

	return true
}

func (r *aaRoomT) loopSettle() bool {
	if r.Players[r.Banker] == nil {
		for _, player := range r.Players {
			r.Banker = player.Player
			break
		}
	}

	banker := r.Players[r.Banker]

	var players []*aaPlayerT
	for _, player := range r.Players {
		if player.Player != r.Banker {
			players = append(players, player)
		}
	}

	for _, player := range players {
		var applyRate int32
		var applySign int32
		if banker.Round.BestWeight >= player.Round.BestWeight {
			applyRate = int32(player.Round.BestRate)
			applySign = -1
			banker.Round.VictoriousNumber++
		} else {
			applyRate = int32(banker.Round.BestRate)
			applySign = 1
			player.Round.VictoriousNumber++
		}

		bs := r.Option.GetScore() * player.Round.Rate * applyRate * applySign * (-1)
		ps := r.Option.GetScore() * player.Round.Rate * applyRate * applySign

		banker.Round.BestPoints += bs
		player.Round.BestPoints += ps

		banker.Round.Points += bs
		player.Round.Points += ps
	}

	r.loop = r.loopRoundClear

	return true
}

func (r *aaRoomT) loopRoundClear() bool {
	r.Step = "round_clear"
	for _, player := range r.Players {
		player.Round.Sent = false
		player.Round.ContinueWithCommitted = false
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.loop = r.loopRoundClearContinue
	r.tick = buildTickNumber(
		8,
		func(number int32) {
		},
		func() {
			for _, player := range r.Players {
				player.Round.ContinueWithCommitted = true
			}
		},
		r.Loop,
	)

	return true
}

func (r *aaRoomT) loopRoundClearContinue() bool {
	finally := true
	for _, player := range r.Players {
		if !player.Round.ContinueWithCommitted {
			finally = false
			if !player.Round.Sent {
				r.Hall.sendNiuniuRoundClear(player.Player, r)
				player.Round.Sent = true
			}
		}
	}

	if !finally {
		return false
	}

	r.tick = nil
	r.loop = r.loopSelect

	return true
}

func (r *aaRoomT) loopSelect() bool {
	if r.RoundNumber < r.Option.GetRoundNumber() {
		r.loop = r.loopTransfer
	} else {
		r.loop = r.loopFinally
	}
	return true
}

func (r *aaRoomT) loopTransfer() bool {
	r.RoundNumber++
	if r.Option.GetBankerMode() == 1 {
		players := r.Players.ToSlice()
		sort.Slice(players, func(i, j int) bool {
			return players[i].Pos < players[j].Pos
		})
		for i, player := range players {
			if player.Player == r.Banker {
				if i < len(players)-1 {
					r.Banker = players[i+1].Player
				} else {
					r.Banker = players[0].Player
				}
			}
		}
	} else if r.Option.GetBankerMode() == 2 {
		r.Banker = 0
	}
	for _, player := range r.Players {
		player.Round = aaPlayerRoundT{
			Points:           player.Round.Points,
			VictoriousNumber: player.Round.VictoriousNumber,
		}
	}

	r.Hall.sendNiuniuStartedForAll(r, r.RoundNumber)

	if r.Option.GetBankerMode() == 0 || r.Option.GetBankerMode() == 1 {
		r.loop = func() bool {
			return r.loopDeal4(r.loopSpecifyRate)
		}
	} else if r.Option.GetBankerMode() == 2 {
		r.loop = func() bool {
			return r.loopDeal4(r.loopGrab)
		}
	}

	return true
}

func (r *aaRoomT) loopFinally() bool {
	r.Step = "round_finally"
	for _, player := range r.Players {
		player.Round.Sent = false
		player.Round.ContinueWithCommitted = false
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.loop = r.loopFinallyContinue
	r.tick = buildTickNumber(
		8,
		func(number int32) {
		},
		func() {
			for _, player := range r.Players {
				player.Round.ContinueWithCommitted = true
			}
		},
		r.Loop,
	)

	for _, player := range r.Players {
		if err := database.CowAddOrderWarHistory(player.Player, r.Id, r.NiuniuRoundFinally()); err != nil {
			log.WithFields(logrus.Fields{
				"err": err,
			}).Warnln("add cow player history failed")
		}
	}

	return true
}

func (r *aaRoomT) loopFinallyContinue() bool {
	finally := true
	for _, player := range r.Players {
		if !player.Round.ContinueWithCommitted {
			finally = false
			if !player.Round.Sent {
				r.Hall.sendNiuniuRoundFinally(player.Player, r)
				player.Round.Sent = true
			}
		}
	}

	if !finally {
		return false
	}

	r.loop = r.loopClean

	return true
}

func (r *aaRoomT) loopClean() bool {
	for _, player := range r.Players {
		if playerData, being := r.Hall.players[player.Player]; being {
			playerData.InsideCow = 0
		}
	}
	delete(r.Hall.cowRooms, r.Id)

	return false
}
