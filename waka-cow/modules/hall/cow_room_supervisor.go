package hall

import (
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

type supervisorRoundPlayerT struct {
	// 得分
	Points int32

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
}

type supervisorPlayerT struct {
	Room *supervisorRoomT

	Player database.Player
	Pos    int32

	Round supervisorRoundPlayerT
}

func (player *supervisorPlayerT) NiuniuRoomDataPlayerData() (pb *cow_proto.NiuniuRoomData_PlayerData) {
	lost := false
	if player, being := player.Room.Hall.players[player.Player]; !being || player.Remote == "" {
		lost = true
	}
	return &cow_proto.NiuniuRoomData_PlayerData{
		Player: int32(player.Player),
		Pos:    player.Pos,
		Ready:  true,
		Lost:   lost,
	}
}

type supervisorPlayerMapT map[database.Player]*supervisorPlayerT

func (players supervisorPlayerMapT) NiuniuRoomDataPlayerData() (pb []*cow_proto.NiuniuRoomData_PlayerData) {
	for _, player := range players {
		pb = append(pb, player.NiuniuRoomDataPlayerData())
	}
	return pb
}

// ---------------------------------------------------------------------------------------------------------------------

type supervisorRoomT struct {
	Hall *actorT

	Id      int32
	Option  *cow_proto.NiuniuRoomOption
	Owner   database.Player
	Players supervisorPlayerMapT

	loop func() bool
	tick func()

	Seats *tools.NumberPool

	Gaming bool
	Step   cow_proto.NiuniuRoundStatus_RoundStep
	Banker database.Player
}

// ---------------------------------------------------------------------------------------------------------------------

func (r *supervisorRoomT) CreateMoney() int32 {
	return 0
}

func (r *supervisorRoomT) JoinMoney() int32 {
	if r.Option.GetMode() == 0 {
		return cow.NormalMaxRate * 5 * r.Option.GetScore() * 3
	} else {
		return cow.CrazyMaxRate * 5 * r.Option.GetScore() * 3
	}
}

func (r *supervisorRoomT) StartMoney() int32 {
	return 0
}

func (r *supervisorRoomT) GetType() cow_proto.NiuniuRoomType {
	return cow_proto.NiuniuRoomType_Flowing
}

func (r *supervisorRoomT) GetId() int32 {
	return r.Id
}

func (r *supervisorRoomT) GetOption() *cow_proto.NiuniuRoomOption {
	return r.Option
}

func (r *supervisorRoomT) GetCreator() database.Player {
	return database.DefaultSupervisor
}

func (r *supervisorRoomT) GetOwner() database.Player {
	return r.Owner
}

func (r *supervisorRoomT) GetGaming() bool {
	return r.Gaming
}

func (r *supervisorRoomT) GetRoundNumber() int32 {
	return 1
}

func (r *supervisorRoomT) GetBanker() database.Player {
	return r.Banker
}

func (r *supervisorRoomT) GetPlayers() []database.Player {
	var d []database.Player
	linq.From(r.Players).SelectT(func(pair linq.KeyValue) database.Player { return pair.Key.(database.Player) }).ToSlice(&d)
	return d
}

func (r *supervisorRoomT) NiuniuRoomData() *cow_proto.NiuniuRoomData {
	return &cow_proto.NiuniuRoomData{
		Type:      cow_proto.NiuniuRoomType_Flowing,
		RoomId:    r.Id,
		Option:    r.GetOption(),
		Creator:   int32(database.DefaultSupervisor),
		Owner:     int32(r.Owner),
		Players:   r.Players.NiuniuRoomDataPlayerData(),
		JoinMoney: r.JoinMoney(),
		Gaming:    r.Gaming,
	}
}

func (r *supervisorRoomT) NiuniuRoundStatus(player database.Player) *cow_proto.NiuniuRoundStatus {
	var pokers []string
	var players []*cow_proto.NiuniuRoundStatus_PlayerData
	for id, playerData := range r.Players {
		players = append(players, &cow_proto.NiuniuRoundStatus_PlayerData{
			Id:              int32(id),
			Points:          playerData.Player.PlayerData().Money / 100,
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
		RoundNumber: 1,
		Players:     players,
		Banker:      int32(r.Banker),
		Pokers:      pokers,
	}
}

func (r *supervisorRoomT) NiuniuRequireGrabShow() *cow_proto.NiuniuRequireGrabShow {
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

func (r *supervisorRoomT) NiuniuRoundClear() *cow_proto.NiuniuRoundClear {
	var players []*cow_proto.NiuniuRoundClear_PlayerData
	for _, player := range r.Players {
		players = append(players, &cow_proto.NiuniuRoundClear_PlayerData{
			Player:     int32(player.Player),
			Points:     player.Player.PlayerData().Money / 100,
			Type:       player.Round.PokersPattern,
			Weight:     player.Round.PokersWeight,
			Rate:       player.Round.PokersRate,
			ThisPoints: player.Round.Points,
			Pokers:     player.Round.CommittedPokers,
		})
	}
	sort.Slice(players, func(i, j int) bool {
		return players[j].Weight < players[i].Weight
	})
	return &cow_proto.NiuniuRoundClear{Players: players, FinallyAt: time.Now().Format("2006-01-02 15:04:05")}
}

func (r *supervisorRoomT) NiuniuGameFinally() *cow_proto.NiuniuGameFinally {
	panic("illegal call")
}

// ---------------------------------------------------------------------------------------------------------------------

func (r *supervisorRoomT) Loop() {
	for {
		if r.loop == nil {
			return
		}
		if !r.loop() {
			return
		}
	}
}

func (r *supervisorRoomT) Tick() {
	if r.tick != nil {
		r.tick()
	}
}

func (r *supervisorRoomT) Left(player *playerT) {
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
}

func (r *supervisorRoomT) Recover(player *playerT) {
	if _, being := r.Players[player.Player]; being {
		r.Players[player.Player].Round.Sent = false
	}

	r.Hall.sendNiuniuUpdateRoomForAll(r)
	if r.Gaming {
		r.Hall.sendNiuniuUpdateRound(player.Player, r)
		r.Loop()
	}
}

func (r *supervisorRoomT) CreateRoom(hall *actorT, id int32, roomType cow_proto.NiuniuRoomType, option *cow_proto.NiuniuRoomOption, creator database.Player) {
	*r = supervisorRoomT{
		Hall:    hall,
		Id:      id,
		Option:  option,
		Players: make(supervisorPlayerMapT, 5),
		Seats:   tools.NewNumberPool(1, 5, false),
	}
}

func (r *supervisorRoomT) JoinRoom(player *playerT) {
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

	r.Players[player.Player] = &supervisorPlayerT{
		Room:   r,
		Player: player.Player,
		Pos:    seat,
	}
	if r.Owner == 0 {
		r.Owner = player.Player
	}

	player.InsideCow = r.Id

	r.Hall.sendNiuniuJoinRoomSuccess(player.Player)
	r.Hall.sendNiuniuUpdateRoomForAll(r)

	r.buildStart()
}

func (r *supervisorRoomT) LeaveRoom(player *playerT) {
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

			if len(r.Players) < 2 {
				r.tick = nil
			}
		}
	}
}

func (r *supervisorRoomT) SwitchReady(player *playerT) {}

func (r *supervisorRoomT) Dismiss(player *playerT) {}

func (r *supervisorRoomT) Start(player *playerT) {}

func (r *supervisorRoomT) SpecifyBanker(player *playerT, banker database.Player) {}

func (r *supervisorRoomT) Grab(player *playerT, grab bool) {
	if r.Gaming {
		r.Players[player.Player].Round.Grab = grab
		r.Players[player.Player].Round.GrabCommitted = true

		r.Hall.sendNiuniuUpdateRoundForAll(r)

		r.Loop()
	}
}

func (r *supervisorRoomT) SpecifyRate(player *playerT, rate int32) {
	if r.Gaming {
		r.Players[player.Player].Round.Rate = rate
		r.Players[player.Player].Round.RateCommitted = true

		r.Hall.sendNiuniuUpdateRoundForAll(r)

		r.Loop()
	}
}

func (r *supervisorRoomT) CommitPokers(player *playerT, pokers []string) {
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

func (r *supervisorRoomT) ContinueWith(player *playerT) {
	if r.Gaming {
		r.Players[player.Player].Round.ContinueWithCommitted = true

		r.Loop()
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func (r *supervisorRoomT) loopStart() bool {
	r.Gaming = true

	r.Hall.sendNiuniuGameStartedForAll(r)
	r.Hall.sendNiuniuRoundStartedForAll(r, 1)

	r.loop = func() bool {
		return r.loopDeal4(r.loopGrab)
	}

	return true
}

func (r *supervisorRoomT) loopDeal4(loop func() bool) bool {
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

func (r *supervisorRoomT) loopGrab() bool {
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

func (r *supervisorRoomT) loopGrabContinue() bool {
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

func (r *supervisorRoomT) loopGrabAnimation() bool {
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

func (r *supervisorRoomT) loopGrabAnimationContinue() bool {
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

func (r *supervisorRoomT) loopGrabSelect() bool {
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

func (r *supervisorRoomT) loopSpecifyRate() bool {
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
		},
		func(deadline int64) {
			r.Hall.sendNiuniuDeadlineForAll(r, deadline)
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

func (r *supervisorRoomT) loopSpecifyRateContinue() bool {
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

func (r *supervisorRoomT) loopDeal1() bool {
	r.Step = cow_proto.NiuniuRoundStatus_CommitPokers
	for _, player := range r.Players {
		var pokers []string
		pokers = append(pokers, player.Round.Pokers4...)
		pokers = append(pokers, player.Round.Pokers1)

		pokers, weight, pattern, _, err := cow.SearchBestPokerPattern(pokers, r.Option.GetMode())
		if err != nil {
			log.WithFields(logrus.Fields{
				"player": player.Player,
				"pokers": pokers,
				"err":    err,
			}).Warnln("search best pokers failed")
		} else {
			player.Round.CommittedPokers = pokers
			player.Round.PokersWeight = int32(weight)
			player.Round.PokersPattern = pattern
		}

		player.Round.Sent = false
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.loop = r.loopCommitPokersContinue
	r.tick = buildTickAfter(
		3,
		func(deadline int64) {
		},
		func(deadline int64) {
			r.Hall.sendNiuniuDeadlineForAll(r, deadline)
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

func (r *supervisorRoomT) loopCommitPokersContinue() bool {
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

func (r *supervisorRoomT) loopSettle() bool {
	if r.Players[r.Banker] == nil {
		for _, player := range r.Players {
			r.Banker = player.Player
			break
		}
	}

	banker := r.Players[r.Banker]

	var players []*supervisorPlayerT
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
		} else {
			applyRate = int32(pr)
			applySign = 1
		}

		banker.Round.Points += r.Option.GetScore() * player.Round.Rate * applyRate * applySign * (-1)
		player.Round.Points += r.Option.GetScore() * player.Round.Rate * applyRate * applySign

		player.Round.PokersPattern = pp
		player.Round.PokersRate = int32(pr)
	}

	var costs []*database.CowFlowingCostData
	for _, player := range players {
		var c *database.CowFlowingCostData
		if player.Round.Points > 0 {
			c = &database.CowFlowingCostData{
				Victory: player.Player,
				Loser:   banker.Player,
				Number:  player.Round.Points * 100,
			}
		} else {
			c = &database.CowFlowingCostData{
				Victory: banker.Player,
				Loser:   player.Player,
				Number:  player.Round.Points * 100 * (-1),
			}
		}
		costs = append(costs, c)
	}
	err := database.CowFlowingCostSettle(costs)
	if err != nil {
		log.WithFields(logrus.Fields{
			"room_id": r.Id,
			"mode":    r.Option.GetMode(),
			"score":   r.Option.GetScore(),
			"cost":    costs,
			"err":     err,
		}).Warnln("supervisor cost failed")
	}

	clear := r.NiuniuRoundClear()
	for _, player := range r.Players {
		if err := database.CowAddFlowingHistory(player.Player, r.Id, clear); err != nil {
			log.WithFields(logrus.Fields{
				"err": err,
			}).Warnln("add cow supervisor record failed")
		}
	}

	r.loop = r.loopSettleSuccess

	return true
}

func (r *supervisorRoomT) loopSettleSuccess() bool {
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

func (r *supervisorRoomT) loopSettleSuccessContinue() bool {
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
	r.loop = r.loopClean

	return true
}

func (r *supervisorRoomT) loopClean() bool {
	for _, player := range r.Players {
		if playerData := r.Hall.players[player.Player]; playerData == nil || playerData.Remote == "" {
			delete(r.Players, player.Player)
			r.Seats.Return(player.Pos)
			if playerData != nil {
				playerData.InsideCow = 0
			}
		}
	}
	for _, player := range r.Players {
		if playerData := player.Player.PlayerData(); playerData.Money < r.JoinMoney()*100 || playerData.Money < r.StartMoney()*100 {
			delete(r.Players, player.Player)
			r.Seats.Return(player.Pos)
			r.Hall.players[player.Player].InsideCow = 0
			r.Hall.sendNiuniuLeftRoom(player.Player, 3)
		}
	}
	for _, player := range r.Players {
		player.Round = supervisorRoundPlayerT{}
	}

	r.tick = nil
	r.loop = nil
	r.Step = cow_proto.NiuniuRoundStatus_Idle
	r.Banker = 0
	r.Gaming = false

	r.Hall.sendNiuniuUpdateRoomForAll(r)

	r.buildStart()

	return false
}
