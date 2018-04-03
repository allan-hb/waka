package hall

import (
	"math"
	"math/rand"
	"reflect"
	"sort"
	"time"

	"github.com/liuhan907/waka/waka-cow/database"
	"github.com/liuhan907/waka/waka-cow/modules/hall/tools"
	"github.com/liuhan907/waka/waka-cow/modules/hall/tools/cow"
	"github.com/liuhan907/waka/waka-cow/proto"
	"github.com/sirupsen/logrus"
	"gopkg.in/ahmetb/go-linq.v3"
)

type playerRoundPlayerT struct {
	// 总分
	Points int32
	// 胜利的场数
	VictoriousNumber int32

	// 本阶段消息是否已发送
	Sent bool

	// 手牌 4 和 1
	Pokers4 []string
	Pokers1 string

	// 是否抢庄
	Grab bool
	// 倍率
	Rate int32
	// 提交的配牌
	CommittedPokers []string

	// 抢庄已提交
	GrabCommitted bool
	// 倍率已提交
	RateCommitted bool
	// 配牌已提交
	PokersCommitted bool
	// 阶段完成已提交
	ContinueWithCommitted bool

	// 本回合权重
	PokersWeight int32
	// 本回合牌型
	PokersPattern string
	// 本回合牌型倍率
	PokersRate int32
	// 本回合得分
	PokersPoints int32
}

type playerPlayerT struct {
	Room *playerRoomT

	Player database.Player
	Pos    int32
	Ready  bool

	Round playerRoundPlayerT
}

func (player *playerPlayerT) NiuniuRoomDataPlayerData() (pb *cow_proto.NiuniuRoomData_PlayerData) {
	lost := false
	if player, being := player.Room.Hall.players[player.Player]; !being || player.Remote == "" {
		lost = true
	}
	return &cow_proto.NiuniuRoomData_PlayerData{
		Player: int32(player.Player),
		Pos:    player.Pos,
		Ready:  player.Ready,
		Lost:   lost,
	}
}

type playerPlayerMapT map[database.Player]*playerPlayerT

func (players playerPlayerMapT) NiuniuRoomDataPlayerData() (pb []*cow_proto.NiuniuRoomData_PlayerData) {
	for _, player := range players {
		pb = append(pb, player.NiuniuRoomDataPlayerData())
	}
	return pb
}

func (players playerPlayerMapT) ToSlice() (d []*playerPlayerT) {
	for _, player := range players {
		d = append(d, player)
	}
	return d
}

// ---------------------------------------------------------------------------------------------------------------------

type playerRoomT struct {
	Hall *actorT

	Id      int32
	Type    cow_proto.NiuniuRoomType
	Option  *cow_proto.NiuniuRoomOption
	Creator database.Player
	Owner   database.Player
	Players playerPlayerMapT

	loop func() bool
	tick func()

	Seats *tools.NumberPool

	Gaming      bool
	RoundNumber int32
	Step        cow_proto.NiuniuRoundStatus_RoundStep
	Banker      database.Player
}

// ---------------------------------------------------------------------------------------------------------------------

func (r *playerRoomT) CreateMoney() int32 {
	if r.Type == cow_proto.NiuniuRoomType_Order {
		return int32(float64(r.Option.GetScore())*0.3+0.5) * r.Option.GetGames()
	} else if r.Type == cow_proto.NiuniuRoomType_PayForAnother {
		return int32(float64(r.Option.GetScore())*0.3+0.5) * r.Option.GetGames() * 5
	} else {
		return math.MaxInt32
	}
}

func (r *playerRoomT) JoinMoney() int32 {
	if r.Type == cow_proto.NiuniuRoomType_Order {
		return r.CreateMoney()
	} else if r.Type == cow_proto.NiuniuRoomType_PayForAnother {
		return 0
	} else {
		return math.MaxInt32
	}
}

func (r *playerRoomT) StartMoney() int32 {
	return r.CreateMoney()
}

func (r *playerRoomT) GetType() cow_proto.NiuniuRoomType {
	return r.Type
}

func (r *playerRoomT) GetId() int32 {
	return r.Id
}

func (r *playerRoomT) GetOption() *cow_proto.NiuniuRoomOption {
	return r.Option
}

func (r *playerRoomT) GetCreator() database.Player {
	return r.Creator
}

func (r *playerRoomT) GetOwner() database.Player {
	return r.Owner
}

func (r *playerRoomT) GetGaming() bool {
	return r.Gaming
}

func (r *playerRoomT) GetRoundNumber() int32 {
	return r.RoundNumber
}

func (r *playerRoomT) GetBanker() database.Player {
	return r.Banker
}

func (r *playerRoomT) GetPlayers() []database.Player {
	var d []database.Player
	linq.From(r.Players).SelectT(func(pair linq.KeyValue) database.Player { return pair.Key.(database.Player) }).ToSlice(&d)
	return d
}

func (r *playerRoomT) NiuniuRoomData() *cow_proto.NiuniuRoomData {
	return &cow_proto.NiuniuRoomData{
		Type:      r.Type,
		RoomId:    r.Id,
		Option:    r.GetOption(),
		Creator:   int32(r.Creator),
		Owner:     int32(r.Owner),
		Players:   r.Players.NiuniuRoomDataPlayerData(),
		JoinMoney: r.JoinMoney(),
		Gaming:    r.Gaming,
	}
}

func (r *playerRoomT) NiuniuRoundStatus(player database.Player) *cow_proto.NiuniuRoundStatus {
	var pokers []string
	var players []*cow_proto.NiuniuRoundStatus_PlayerData
	for id, playerData := range r.Players {
		players = append(players, &cow_proto.NiuniuRoundStatus_PlayerData{
			Id:              int32(id),
			Points:          playerData.Round.Points,
			Grab:            playerData.Round.Grab,
			Rate:            playerData.Round.Rate,
			GrabCommitted:   playerData.Round.GrabCommitted,
			RateCommitted:   playerData.Round.RateCommitted,
			PokersCommitted: playerData.Round.PokersCommitted,
		})
		if playerData.Player == player {
			if len(playerData.Round.Pokers4) > 0 {
				pokers = append(pokers, playerData.Round.Pokers4...)
			}
			if len(playerData.Round.Pokers1) > 0 &&
				(r.Step == cow_proto.NiuniuRoundStatus_CommitPokers ||
					r.Step == cow_proto.NiuniuRoundStatus_RoundClear ||
					r.Step == cow_proto.NiuniuRoundStatus_GameFinally) {
				pokers = append(pokers, playerData.Round.Pokers1)
			}
		}
	}

	return &cow_proto.NiuniuRoundStatus{
		Step:        r.Step,
		RoundNumber: r.RoundNumber,
		Players:     players,
		Banker:      int32(r.Banker),
		Pokers:      pokers,
	}
}

func (r *playerRoomT) NiuniuRequireGrabShow() *cow_proto.NiuniuRequireGrabShow {
	var players []*cow_proto.NiuniuRequireGrabShow_PlayerData
	for _, player := range r.Players {
		players = append(players, &cow_proto.NiuniuRequireGrabShow_PlayerData{
			Player: int32(player.Player),
			Grab:   player.Round.Grab,
		})
	}
	return &cow_proto.NiuniuRequireGrabShow{
		Players: players,
	}
}

func (r *playerRoomT) NiuniuRoundClear() *cow_proto.NiuniuRoundClear {
	var players []*cow_proto.NiuniuRoundClear_PlayerData
	for _, player := range r.Players {
		players = append(players, &cow_proto.NiuniuRoundClear_PlayerData{
			Player:     int32(player.Player),
			Points:     player.Round.Points,
			Type:       player.Round.PokersPattern,
			Weight:     player.Round.PokersWeight,
			Rate:       player.Round.PokersRate,
			ThisPoints: player.Round.PokersPoints,
			Pokers:     player.Round.CommittedPokers,
		})
	}
	sort.Slice(players, func(i, j int) bool {
		return players[j].Weight < players[i].Weight
	})
	return &cow_proto.NiuniuRoundClear{Players: players, FinallyAt: time.Now().Format("2006-01-02 15:04:05")}
}

func (r *playerRoomT) NiuniuGameFinally() *cow_proto.NiuniuGameFinally {
	var players []*cow_proto.NiuniuGameFinally_PlayerData
	for _, player := range r.Players {
		players = append(players, &cow_proto.NiuniuGameFinally_PlayerData{
			Player:    int32(player.Player),
			Points:    int32(player.Round.Points),
			Victories: player.Round.VictoriousNumber,
		})
	}
	return &cow_proto.NiuniuGameFinally{Players: players, FinallyAt: time.Now().Format("2006-01-02 15:04:05")}
}

// ---------------------------------------------------------------------------------------------------------------------

func (r *playerRoomT) Loop() {
	for {
		if r.loop == nil {
			return
		}
		if !r.loop() {
			return
		}
	}
}

func (r *playerRoomT) Tick() {
	if r.tick != nil {
		r.tick()
	}
}

func (r *playerRoomT) Left(player *playerT) {
	if r.Type == cow_proto.NiuniuRoomType_Order {
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
	} else if r.Type == cow_proto.NiuniuRoomType_PayForAnother {
		if !r.Gaming {
			if roomPlayer, being := r.Players[player.Player]; being {
				delete(r.Players, player.Player)
				player.InsideCow = 0
				r.Seats.Return(roomPlayer.Pos)

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
	} else {
		panic("illegal room type")
	}

}

func (r *playerRoomT) Recover(player *playerT) {
	if _, being := r.Players[player.Player]; being {
		r.Players[player.Player].Round.Sent = false
	}

	r.Hall.sendNiuniuUpdateRoomForAll(r)
	if r.Gaming {
		r.Hall.sendNiuniuUpdateRound(player.Player, r)
		r.Loop()
	}
}

func (r *playerRoomT) CreateRoom(hall *actorT, id int32, roomType cow_proto.NiuniuRoomType, option *cow_proto.NiuniuRoomOption, creator database.Player) {
	*r = playerRoomT{
		Hall:    hall,
		Id:      id,
		Type:    roomType,
		Option:  option,
		Creator: creator,
		Players: make(playerPlayerMapT, 5),
		Seats:   tools.NewNumberPool(1, 5, false),
	}

	if roomType == cow_proto.NiuniuRoomType_Order {
		pos, _ := r.Seats.Acquire()
		r.Players[creator] = &playerPlayerT{
			Room:   r,
			Player: creator,
			Pos:    pos,
		}
		r.Owner = creator
	}

	if creator.PlayerData().Money < r.CreateMoney()*100 {
		r.Hall.sendNiuniuCreateRoomFailed(creator, 1)
	} else {
		r.Hall.cowRooms[id] = r
		r.Hall.sendNiuniuCreateRoomSuccess(creator, r.Id)

		if roomType == cow_proto.NiuniuRoomType_Order {
			r.Hall.players[creator].InsideCow = id
			r.Hall.sendNiuniuJoinRoomSuccess(creator)
			r.Hall.sendNiuniuUpdateRoomForAll(r)
		}
	}
}

func (r *playerRoomT) JoinRoom(player *playerT) {
	if player.Player.PlayerData().Money < r.JoinMoney()*100 {
		r.Hall.sendNiuniuJoinRoomFailed(player.Player, 1)
		return
	}

	_, being := r.Players[player.Player]
	if being {
		r.Hall.sendNiuniuJoinRoomFailed(player.Player, -1)
		return
	}

	if r.Gaming {
		r.Hall.sendNiuniuJoinRoomFailed(player.Player, 2)
		return
	}

	seat, has := r.Seats.Acquire()
	if !has {
		r.Hall.sendNiuniuJoinRoomFailed(player.Player, 0)
		return
	}

	r.Players[player.Player] = &playerPlayerT{
		Room:   r,
		Player: player.Player,
		Pos:    seat,
	}
	player.InsideCow = r.Id

	if r.Owner == 0 {
		r.Owner = player.Player
	}

	r.Hall.sendNiuniuJoinRoomSuccess(player.Player)
	r.Hall.sendNiuniuUpdateRoomForAll(r)
}

func (r *playerRoomT) LeaveRoom(player *playerT) {
	if r.Type == cow_proto.NiuniuRoomType_Order {
		if !r.Gaming {
			if player.Player == r.Owner {
				delete(r.Hall.cowRooms, r.Id)
				for _, player := range r.Players {
					r.Hall.players[player.Player].InsideCow = 0
					r.Hall.sendNiuniuLeftRoom(player.Player, 2)
				}
			} else {
				if roomPlayer, being := r.Players[player.Player]; being {
					player.InsideCow = 0
					delete(r.Players, player.Player)
					r.Seats.Return(roomPlayer.Pos)

					r.Hall.sendNiuniuLeftRoom(player.Player, 1)
					r.Hall.sendNiuniuUpdateRoomForAll(r)
				}
			}
		}
	} else if r.Type == cow_proto.NiuniuRoomType_PayForAnother {
		if !r.Gaming {
			if roomPlayer, being := r.Players[player.Player]; being {
				player.InsideCow = 0
				delete(r.Players, player.Player)
				r.Seats.Return(roomPlayer.Pos)

				r.Hall.sendNiuniuLeftRoom(player.Player, 1)

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
	} else {
		panic("illegal room type")
	}
}

func (r *playerRoomT) SwitchReady(player *playerT) {
	if !r.Gaming {
		if roomPlayer, being := r.Players[player.Player]; being {
			roomPlayer.Ready = !roomPlayer.Ready
			r.Hall.sendNiuniuUpdateRoomForAll(r)
		}
	}
}

func (r *playerRoomT) Dismiss(player *playerT) {
	if !r.Gaming {
		if (r.Type == cow_proto.NiuniuRoomType_Order && r.Owner == player.Player) ||
			(r.Type == cow_proto.NiuniuRoomType_PayForAnother && r.Creator == player.Player) {
			delete(r.Hall.cowRooms, r.Id)
			for _, player := range r.Players {
				r.Hall.players[player.Player].InsideCow = 0
				r.Hall.sendNiuniuLeftRoom(player.Player, 2)
			}
		}
	}
}

func (r *playerRoomT) Start(player *playerT) {
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

			var costs []*database.CowOrderCostData
			if r.Type == cow_proto.NiuniuRoomType_Order {
				for _, player := range r.Players {
					costs = append(costs, &database.CowOrderCostData{
						Player: player.Player,
						Number: r.StartMoney() * 100,
					})
				}
			} else if r.Type == cow_proto.NiuniuRoomType_PayForAnother {
				costs = append(costs, &database.CowOrderCostData{
					Player: r.Creator,
					Number: r.StartMoney() * 100,
				})
			} else {
				panic("illegal room type")
			}

			err := database.CowOrderCostSettle(costs)
			if err != nil {
				log.WithFields(logrus.Fields{
					"id":     r.Id,
					"type":   r.Type,
					"option": r.Option.String(),
					"cost":   costs,
					"err":    err,
				}).Warnln("cow cost settle failed")
				return
			}

			r.loop = r.loopStart
			r.Loop()
		}
	}
}

func (r *playerRoomT) SpecifyBanker(player *playerT, banker database.Player) {
	if r.Gaming {
		if _, being := r.Players[banker]; being {
			r.Banker = banker

			r.Loop()
		}
	}
}

func (r *playerRoomT) Grab(player *playerT, grab bool) {
	if r.Gaming {
		r.Players[player.Player].Round.Grab = grab
		r.Players[player.Player].Round.GrabCommitted = true

		r.Hall.sendNiuniuUpdateRoundForAll(r)

		r.Loop()
	}
}

func (r *playerRoomT) SpecifyRate(player *playerT, rate int32) {
	if r.Gaming {
		r.Players[player.Player].Round.Rate = rate
		r.Players[player.Player].Round.RateCommitted = true

		r.Hall.sendNiuniuUpdateRoundForAll(r)

		r.Loop()
	}
}

func (r *playerRoomT) CommitPokers(player *playerT, pokers []string) {
	if r.Gaming {
		var origin []string
		var committed []string
		origin = append(origin, r.Players[player.Player].Round.CommittedPokers...)
		committed = append(committed, pokers...)
		sort.Slice(origin, func(i, j int) bool {
			return origin[i] < origin[j]
		})
		sort.Slice(committed, func(i, j int) bool {
			return committed[i] < committed[j]
		})
		if !reflect.DeepEqual(origin, committed) {
			log.WithFields(logrus.Fields{
				"origin":    origin,
				"committed": committed,
			}).Warnln("committed pokers not equal origin pokers")
			return
		}

		r.Players[player.Player].Round.CommittedPokers = pokers
		r.Players[player.Player].Round.PokersCommitted = true

		r.Loop()
	}
}

func (r *playerRoomT) ContinueWith(player *playerT) {
	if r.Gaming {
		r.Players[player.Player].Round.ContinueWithCommitted = true

		r.Loop()
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func (r *playerRoomT) loopStart() bool {
	r.Gaming = true
	r.RoundNumber = 1

	r.Hall.sendNiuniuGameStartedForAll(r)
	r.Hall.sendNiuniuRoundStartedForAll(r, r.RoundNumber)

	if r.Option.Banker == 0 || r.Option.Banker == 1 {
		r.loop = r.loopSpecifyBanker
	} else if r.Option.Banker == 2 {
		r.loop = func() bool {
			return r.loopDeal4(r.loopGrab)
		}
	}

	return true
}

func (r *playerRoomT) loopSpecifyBanker() bool {
	r.Step = cow_proto.NiuniuRoundStatus_SpecifyBanker
	for _, player := range r.Players {
		player.Round.Sent = false
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.loop = r.loopSpecifyBankerContinue
	r.tick = buildTickAfter(
		8,
		func(deadline int64) {
			r.Hall.sendNiuniuDeadlineForAll(r, deadline)
		},
		func(deadline int64) {
		},
		func() {
			r.Banker = r.Owner
		},
		r.Loop,
	)

	return true
}

func (r *playerRoomT) loopSpecifyBankerContinue() bool {
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

func (r *playerRoomT) loopDeal4(loop func() bool) bool {
	pokers := cow.Acquire5(len(r.Players))
	i := 0
	for _, player := range r.Players {
		pokers := pokers[i]
		player.Round.Pokers4 = append(player.Round.Pokers4, pokers[:4]...)
		player.Round.Pokers1 = pokers[4]
		i++
	}

	for _, player := range r.Players {
		r.Hall.sendNiuniuDeal4(player.Player, player.Round.Pokers4)
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.loop = loop

	return true
}

func (r *playerRoomT) loopGrab() bool {
	r.Step = cow_proto.NiuniuRoundStatus_Grab
	for _, player := range r.Players {
		player.Round.Sent = false
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.loop = r.loopGrabContinue
	r.tick = buildTickAfter(
		6,
		func(deadline int64) {
		},
		func(deadline int64) {
			r.Hall.sendNiuniuDeadlineForAll(r, deadline)
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

func (r *playerRoomT) loopGrabContinue() bool {
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

func (r *playerRoomT) loopGrabAnimation() bool {
	r.Step = cow_proto.NiuniuRoundStatus_GrabShow
	for _, player := range r.Players {
		player.Round.Sent = false
		player.Round.ContinueWithCommitted = false
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.loop = r.loopGrabAnimationContinue
	r.tick = buildTickAfter(
		8,
		func(deadline int64) {
			r.Hall.sendNiuniuDeadlineForAll(r, deadline)
		},
		func(deadline int64) {
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

func (r *playerRoomT) loopGrabAnimationContinue() bool {
	finally := true
	for _, player := range r.Players {
		if !player.Round.ContinueWithCommitted {
			finally = false
			if !player.Round.Sent {
				r.Hall.sendNiuniuRequireGrabShow(player.Player, r)
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

func (r *playerRoomT) loopGrabSelect() bool {
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

func (r *playerRoomT) loopSpecifyRate() bool {
	r.Step = cow_proto.NiuniuRoundStatus_SpecifyRate
	for _, player := range r.Players {
		player.Round.Sent = false
		if player.Player == r.Banker {
			player.Round.Rate = 1
			player.Round.RateCommitted = true
		}
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.loop = r.loopSpecifyRateContinue
	r.tick = buildTickAfter(
		5,
		func(deadline int64) {
			r.Hall.sendNiuniuDeadlineForAll(r, deadline)
		},
		func(deadline int64) {
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

func (r *playerRoomT) loopSpecifyRateContinue() bool {
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
	r.loop = r.loopDeal1

	return true
}

func (r *playerRoomT) loopDeal1() bool {
	r.Step = cow_proto.NiuniuRoundStatus_CommitPokers
	for _, player := range r.Players {
		var pokers []string
		pokers = append(pokers, player.Round.Pokers4...)
		pokers = append(pokers, player.Round.Pokers1)

		pokers, _, pattern, _, err := cow.SearchBestPokerPattern(pokers, r.Option.GetMode())
		if err != nil {
			log.WithFields(logrus.Fields{
				"player": player.Player,
				"pokers": pokers,
				"err":    err,
			}).Warnln("search best pokers failed")
		} else {
			player.Round.CommittedPokers = pokers
			player.Round.PokersPattern = pattern
		}

		player.Round.Sent = false
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.loop = r.loopCommitPokersContinue
	r.tick = buildTickAfter(
		3,
		func(deadline int64) {
			r.Hall.sendNiuniuDeadlineForAll(r, deadline)
		},
		func(deadline int64) {
		},
		func() {
			for _, player := range r.Players {
				player.Round.PokersCommitted = true
			}
		},
		r.Loop,
	)

	return true
}

func (r *playerRoomT) loopCommitPokersContinue() bool {
	finally := true
	for _, player := range r.Players {
		if !player.Round.PokersCommitted {
			finally = false
			if !player.Round.Sent {
				r.Hall.sendNiuniuDeal1(player.Player, player.Round.Pokers1, player.Round.PokersPattern, player.Round.CommittedPokers)
				player.Round.Sent = true
			}
		}
	}

	if !finally {
		return false
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.tick = nil
	r.loop = r.loopSettle

	return true
}

func (r *playerRoomT) loopSettle() bool {
	if r.Players[r.Banker] == nil {
		for _, player := range r.Players {
			r.Banker = player.Player
			break
		}
	}

	banker := r.Players[r.Banker]

	var players []*playerPlayerT
	for _, player := range r.Players {
		if player.Player != r.Banker {
			players = append(players, player)
		}
	}

	bw, bp, br, _ := cow.GetPokersPattern(banker.Round.CommittedPokers, r.Option.GetMode())
	banker.Round.PokersPattern = bp
	banker.Round.PokersRate = int32(br)
	for _, player := range players {
		var applyRate int32
		var applySign int32
		pw, pp, pr, _ := cow.GetPokersPattern(player.Round.CommittedPokers, r.Option.GetMode())
		if bw >= pw {
			applyRate = int32(br)
			applySign = -1
			banker.Round.VictoriousNumber++
		} else {
			applyRate = int32(pr)
			applySign = 1
			player.Round.VictoriousNumber++
		}

		bs := r.Option.GetScore() * player.Round.Rate * applyRate * applySign * (-1)
		ps := r.Option.GetScore() * player.Round.Rate * applyRate * applySign

		banker.Round.PokersPoints += bs
		player.Round.PokersPoints += ps

		banker.Round.Points += bs
		player.Round.Points += ps

		player.Round.PokersPattern = pp
		player.Round.PokersRate = int32(pr)
	}

	r.loop = r.loopSettleSuccess

	return true
}

func (r *playerRoomT) loopSettleSuccess() bool {
	r.Step = cow_proto.NiuniuRoundStatus_RoundClear
	for _, player := range r.Players {
		player.Round.Sent = false
		player.Round.ContinueWithCommitted = false
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.loop = r.loopSettleSuccessContinue
	r.tick = buildTickAfter(
		8,
		func(deadline int64) {
		},
		func(deadline int64) {
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

func (r *playerRoomT) loopSettleSuccessContinue() bool {
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

func (r *playerRoomT) loopSelect() bool {
	if r.RoundNumber < r.Option.GetGames() {
		r.loop = r.loopTransfer
	} else {
		r.loop = r.loopFinally
	}
	return true
}

func (r *playerRoomT) loopTransfer() bool {
	r.RoundNumber++
	if r.Option.Banker == 1 {
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
	} else if r.Option.Banker == 2 {
		r.Banker = 0
	}
	for _, player := range r.Players {
		player.Round = playerRoundPlayerT{
			Points:           player.Round.Points,
			VictoriousNumber: player.Round.VictoriousNumber,
		}
	}

	r.Hall.sendNiuniuRoundStartedForAll(r, r.RoundNumber)

	if r.Option.Banker == 0 || r.Option.Banker == 1 {
		r.loop = func() bool {
			return r.loopDeal4(r.loopSpecifyRate)
		}
	} else if r.Option.Banker == 2 {
		r.loop = func() bool {
			return r.loopDeal4(r.loopGrab)
		}
	}

	return true
}

func (r *playerRoomT) loopFinally() bool {
	r.Step = cow_proto.NiuniuRoundStatus_GameFinally
	for _, player := range r.Players {
		player.Round.Sent = false
		player.Round.ContinueWithCommitted = false
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.loop = r.loopFinallyContinue
	r.tick = buildTickAfter(
		8,
		func(deadline int64) {
		},
		func(deadline int64) {
		},
		func() {
			for _, player := range r.Players {
				player.Round.ContinueWithCommitted = true
			}
		},
		r.Loop,
	)

	for _, player := range r.Players {
		if err := database.CowAddHistory(player.Player, r.Id, r.Type, r.NiuniuGameFinally()); err != nil {
			log.WithFields(logrus.Fields{
				"err": err,
			}).Warnln("add cow player history failed")
		}
	}

	return true
}

func (r *playerRoomT) loopFinallyContinue() bool {
	finally := true
	for _, player := range r.Players {
		if !player.Round.ContinueWithCommitted {
			finally = false
			if !player.Round.Sent {
				r.Hall.sendNiuniuGameFinally(player.Player, r)
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

func (r *playerRoomT) loopClean() bool {
	for _, player := range r.Players {
		if playerData := r.Hall.players[player.Player]; playerData == nil || playerData.Remote == "" {
			if player.Player != r.Owner {
				delete(r.Players, player.Player)
				r.Seats.Return(player.Pos)
				if playerData != nil {
					playerData.InsideCow = 0
				}
			}
		}
	}
	for _, player := range r.Players {
		if r.Type == cow_proto.NiuniuRoomType_Order {
			if player.Player.PlayerData().Money < r.JoinMoney()*100 {
				if player.Player == r.Owner {
					delete(r.Hall.cowRooms, r.Id)
					for _, player := range r.Players {
						if playerData := r.Hall.players[player.Player]; playerData != nil {
							playerData.InsideCow = 0
						}
						r.Hall.sendNiuniuLeftRoom(player.Player, 2)
					}
					return false
				} else {
					delete(r.Players, player.Player)
					r.Seats.Return(player.Pos)
					if playerData := r.Hall.players[player.Player]; playerData != nil {
						playerData.InsideCow = 0
					}
					r.Hall.sendNiuniuLeftRoom(player.Player, 3)
				}
			}
		} else if r.Type == cow_proto.NiuniuRoomType_PayForAnother {
			if r.Creator.PlayerData().Money < r.CreateMoney()*100 {
				delete(r.Hall.cowRooms, r.Id)
				for _, player := range r.Players {
					if playerData := r.Hall.players[player.Player]; playerData != nil {
						playerData.InsideCow = 0
					}
					r.Hall.sendNiuniuLeftRoom(player.Player, 3)
				}
				return false
			}
		} else {
			player.Ready = false
		}
	}
	for _, player := range r.Players {
		player.Round = playerRoundPlayerT{}
	}

	r.tick = nil
	r.loop = nil
	r.Step = cow_proto.NiuniuRoundStatus_Idle
	r.Banker = 0
	r.Gaming = false

	r.Hall.sendNiuniuUpdateRoomForAll(r)

	return false
}
