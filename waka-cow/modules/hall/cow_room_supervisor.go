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

var (
	rnd = rand.New(rand.NewSource(time.Now().Unix()))
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

func (player *supervisorPlayerT) NiuniuRoomData2RoomPlayer() (pb *waka.NiuniuRoomData2_RoomPlayer) {
	lost := false
	if player, being := player.Room.Hall.players[player.Player]; !being || player.Remote == "" {
		lost = true
	}
	return &waka.NiuniuRoomData2_RoomPlayer{
		Player: player.Room.Hall.ToPlayer(player.Player),
		Pos:    player.Pos,
		Ready:  true,
		Lost:   lost,
	}
}

type supervisorPlayerMapT map[database.Player]*supervisorPlayerT

func (players supervisorPlayerMapT) NiuniuRoomData2RoomPlayer() (pb []*waka.NiuniuRoomData2_RoomPlayer) {
	for _, player := range players {
		pb = append(pb, player.NiuniuRoomData2RoomPlayer())
	}
	return pb
}

// ---------------------------------------------------------------------------------------------------------------------

type supervisorRoomT struct {
	Hall *actorT

	Id        int32
	Mode      int32
	Score     int32
	Creator   database.Player
	Owner     database.Player
	Players   supervisorPlayerMapT
	Observers map[database.Player]database.Player

	loop func() bool
	tick func()

	Seats *tools.NumberPool

	Gaming bool
	Step   string
	Banker database.Player

	King         []database.Player
	Distribution map[database.Player][]string
}

// ---------------------------------------------------------------------------------------------------------------------

func (r *supervisorRoomT) CreateMoney() int32 {
	return 0
}

func (r *supervisorRoomT) EnterMoney() int32 {
	if r.Mode == 0 {
		return cow.NormalMaxRate * 5 * r.Score * 3
	} else {
		return cow.CrazyMaxRate * 5 * r.Score * 3
	}
}

func (r *supervisorRoomT) LeaveMoney() int32 {
	return r.EnterMoney()
}

func (r *supervisorRoomT) CostMoney() int32 {
	return 0
}

func (r *supervisorRoomT) GetType() waka.NiuniuRoomType {
	return waka.NiuniuRoomType_Agent
}

func (r *supervisorRoomT) GetId() int32 {
	return r.Id
}

func (r *supervisorRoomT) GetOption() *waka.NiuniuRoomOption {
	return &waka.NiuniuRoomOption{
		Banker: 2,
		Mode:   r.Mode,
		Score:  r.Score,
		Games:  1,
		IsAA:   false,
	}
}

func (r *supervisorRoomT) GetCreator() database.Player {
	return r.Creator
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

func (r *supervisorRoomT) GetObservers() []database.Player {
	var d []database.Player
	linq.From(r.Observers).SelectT(func(pair linq.KeyValue) database.Player { return pair.Key.(database.Player) }).ToSlice(&d)
	return d
}

func (r *supervisorRoomT) NiuniuRoomData1() *waka.NiuniuRoomData1 {
	return &waka.NiuniuRoomData1{
		Id:         r.Id,
		Option:     r.GetOption(),
		Creator:    r.Creator.PlayerData().Nickname,
		Owner:      r.Owner.PlayerData().Nickname,
		Players:    int32(len(r.Players)),
		EnterMoney: r.EnterMoney(),
		LeaveMoney: r.LeaveMoney(),
		Gaming:     r.Gaming,
	}
}

func (r *supervisorRoomT) NiuniuRoomData2() *waka.NiuniuRoomData2 {
	return &waka.NiuniuRoomData2{
		Type:       waka.NiuniuRoomType_Agent,
		Id:         r.Id,
		Option:     r.GetOption(),
		Creator:    r.Hall.ToPlayer(r.Creator),
		Owner:      r.Hall.ToPlayer(r.Owner),
		Players:    r.Players.NiuniuRoomData2RoomPlayer(),
		Observers:  r.Hall.ToPlayerMap(r.Observers),
		EnterMoney: r.EnterMoney(),
		LeaveMoney: r.LeaveMoney(),
		Gaming:     r.Gaming,
	}
}

func (r *supervisorRoomT) NiuniuRoundStatus(player database.Player) *waka.NiuniuRoundStatus {
	var pokers []string
	var players []*waka.NiuniuRoundStatus_RoundPlayer
	for id, playerData := range r.Players {
		players = append(players, &waka.NiuniuRoundStatus_RoundPlayer{
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
				(r.Step == "require_commit_pokers" || r.Step == "round_clear" || r.Step == "round_finally") {
				pokers = append(pokers, playerData.Round.Pokers1)
			}
		}
	}

	return &waka.NiuniuRoundStatus{
		Step:        r.Step,
		RoomId:      r.Id,
		RoundNumber: 1,
		Players:     players,
		Banker:      int32(r.Banker),
		Pokers:      pokers,
	}
}

func (r *supervisorRoomT) NiuniuGrabAnimation() *waka.NiuniuGrabAnimation {
	var players []*waka.NiuniuGrabAnimation_GrabPlayer
	for _, player := range r.Players {
		players = append(players, &waka.NiuniuGrabAnimation_GrabPlayer{
			PlayerId: int32(player.Player),
			Grab:     player.Round.Grab,
		})
	}
	return &waka.NiuniuGrabAnimation{
		Players: players,
	}
}

func (r *supervisorRoomT) NiuniuRoundClear() *waka.NiuniuRoundClear {
	var players []*waka.NiuniuRoundClear_RoundClearPlayer
	for _, player := range r.Players {
		players = append(players, &waka.NiuniuRoundClear_RoundClearPlayer{
			Player:     r.Hall.ToPlayer(player.Player),
			Type:       player.Round.PokersPattern,
			Rate:       player.Round.PokersRate,
			ThisPoints: player.Round.Points,
			Pokers:     player.Round.CommittedPokers,
			Points:     player.Player.PlayerData().Money / 100,
			Weight:     player.Round.PokersWeight,
		})
	}
	sort.Slice(players, func(i, j int) bool {
		return players[j].Weight < players[i].Weight
	})
	return &waka.NiuniuRoundClear{Players: players, FinallyAt: time.Now().Format("2006-01-02 15:04:05")}
}

func (r *supervisorRoomT) NiuniuRoundFinally() *waka.NiuniuRoundFinally {
	var players []*waka.NiuniuRoundFinally_RoundFinallyPlayer
	for _, player := range r.Players {
		players = append(players, &waka.NiuniuRoundFinally_RoundFinallyPlayer{
			Player:    r.Hall.ToPlayer(player.Player),
			Points:    int32(player.Round.Points),
			Victories: 0,
		})
	}
	return &waka.NiuniuRoundFinally{Players: players, FinallyAt: time.Now().Format("2006-01-02 15:04:05")}
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
		if _, being := r.Observers[player.Player]; being {
			delete(r.Observers, player.Player)
			player.InsideCow = 0
		} else if roomPlayer, being := r.Players[player.Player]; being {
			delete(r.Players, player.Player)
			player.InsideCow = 0
			r.Seats.Return(roomPlayer.Pos)

			if r.Owner == player.Player {
				r.Owner = 0

				log.WithFields(logrus.Fields{
					"type":    "owner_left",
					"room_id": r.Id,
					"owner":   r.Owner,
				}).Debugln("owner changed")

				if len(r.Players) > 0 {
					for _, player := range r.Players {
						r.Owner = player.Player

						log.WithFields(logrus.Fields{
							"type":    "left",
							"room_id": r.Id,
							"owner":   r.Owner,
						}).Debugln("owner changed")

						break
					}
				}
			}
		}
	} else {
		if _, being := r.Observers[player.Player]; being {
			delete(r.Observers, player.Player)
			player.InsideCow = 0
		}
	}

	r.Hall.sendNiuniuUpdateRoomForAll(r)
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

func (r *supervisorRoomT) CreateRoom(hall *actorT, id int32, option *waka.NiuniuRoomOption, creator database.Player) cowRoom {
	return &supervisorRoomT{
		Hall:      hall,
		Id:        id,
		Mode:      option.GetMode(),
		Score:     option.GetScore(),
		Creator:   creator,
		Players:   make(supervisorPlayerMapT, 5),
		Observers: map[database.Player]database.Player{},
		Seats:     tools.NewNumberPool(1, 5, false),
	}
}

func (r *supervisorRoomT) JoinRoom(player *playerT) {
	if player.Player.PlayerData().Money < r.EnterMoney()*100 {
		r.Hall.sendNiuniuJoinRoomFailed(player.Player, 1)
		return
	}

	_, being := r.Players[player.Player]
	if being {
		r.Hall.sendNiuniuJoinRoomFailed(player.Player, 2)
		return
	}

	_, being = r.Observers[player.Player]
	if being {
		r.Hall.sendNiuniuJoinRoomFailed(player.Player, 2)
		return
	}

	if r.Gaming {
		r.Observers[player.Player] = player.Player
	} else {
		seat, has := r.Seats.Acquire()
		if !has {
			r.Observers[player.Player] = player.Player
		} else {
			r.Players[player.Player] = &supervisorPlayerT{
				Room:   r,
				Player: player.Player,
				Pos:    seat,
			}
			if r.Owner == 0 {
				r.Owner = player.Player

				log.WithFields(logrus.Fields{
					"type":    "join",
					"room_id": r.Id,
					"owner":   r.Owner,
				}).Debugln("owner changed")
			}

			if player.Player.PlayerData().VictoryRate > 0 {
				r.King = append(r.King, player.Player)
			}
		}
	}

	player.InsideCow = r.Id

	r.Hall.sendNiuniuRoomJoined(player.Player, r)
	r.Hall.sendNiuniuUpdateRoomForAll(r)

	r.buildStart()
}

func (r *supervisorRoomT) LeaveRoom(player *playerT) {
	if _, being := r.Observers[player.Player]; being {
		player.InsideCow = 0
		delete(r.Observers, player.Player)

		r.Hall.sendNiuniuRoomLeft(player.Player)
		r.Hall.sendNiuniuUpdateRoomForAll(r)
	} else {
		if !r.Gaming {
			if roomPlayer, being := r.Players[player.Player]; being {
				player.InsideCow = 0
				delete(r.Players, player.Player)
				r.Seats.Return(roomPlayer.Pos)

				r.Hall.sendNiuniuRoomLeft(player.Player)

				if r.Owner == player.Player {
					r.Owner = 0

					log.WithFields(logrus.Fields{
						"type":    "owner_leave",
						"room_id": r.Id,
						"owner":   r.Owner,
					}).Debugln("owner changed")

					if len(r.Players) > 0 {
						for _, player := range r.Players {
							r.Owner = player.Player

							log.WithFields(logrus.Fields{
								"type":    "leave",
								"room_id": r.Id,
								"owner":   r.Owner,
							}).Debugln("owner changed")

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
}

func (r *supervisorRoomT) SwitchReady(player *playerT) {}

func (r *supervisorRoomT) SwitchRole(player *playerT) {
	if !r.Gaming {
		if roomPlayer, being := r.Players[player.Player]; being {
			r.Seats.Return(roomPlayer.Pos)
			delete(r.Players, player.Player)
			r.Observers[player.Player] = player.Player

			if r.Owner == player.Player {
				r.Owner = 0

				log.WithFields(logrus.Fields{
					"type":    "owner_switch_role_to_observer",
					"room_id": r.Id,
					"owner":   r.Owner,
				}).Debugln("owner changed")

				if len(r.Players) > 0 {
					for _, player := range r.Players {
						r.Owner = player.Player

						log.WithFields(logrus.Fields{
							"type":    "switch_role_to_observer",
							"room_id": r.Id,
							"owner":   r.Owner,
						}).Debugln("owner changed")

						break
					}
				}
			}
		} else if _, being := r.Observers[player.Player]; being {
			seat, ok := r.Seats.Acquire()
			if !ok {
				return
			}
			delete(r.Observers, player.Player)
			r.Players[player.Player] = &supervisorPlayerT{
				Room:   r,
				Player: player.Player,
				Pos:    seat,
			}

			if r.Owner == 0 {
				r.Owner = player.Player

				log.WithFields(logrus.Fields{
					"type":    "switch_role_to_player",
					"room_id": r.Id,
					"owner":   r.Owner,
				}).Debugln("owner changed")
			}

			if player.Player.PlayerData().VictoryRate > 0 {
				r.King = append(r.King, player.Player)
			}

			r.buildStart()
		}

		r.Hall.sendNiuniuUpdateRoomForAll(r)
	}
}

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
	var idx int
	var king database.Player
	for i, k := range r.King {
		if _, being := r.Players[k]; being {
			king = k
			idx = i
			break
		}
	}
	if king != 0 {
		r.King = r.King[idx:]

		var players []database.Player
		linq.From(r.Players).SelectT(func(x linq.KeyValue) database.Player {
			return x.Key.(database.Player)
		}).ToSlice(&players)

		victoryRate := float64(king.PlayerData().VictoryRate) / 100
		randRate := rnd.Float64()
		log.WithFields(logrus.Fields{
			"king":         king,
			"victory_rate": victoryRate,
			"rand_rate":    randRate,
		}).Debugln("control rate")
		if randRate < victoryRate {
			r.Distribution = cow.DistributingOnce(king, players, r.Mode)
		}
	}

	r.Gaming = true

	r.Hall.sendNiuniuStartedForAll(r, 1)

	r.loop = func() bool {
		return r.loopDeal4(r.loopGrab)

	}

	idleRooms := r.Hall.cowRooms.
		WhereSupervisor().
		WhereScore(r.Score).
		WhereMode(r.Mode).
		WhereIdle()

	if len(idleRooms) == 0 {
		id, ok := r.Hall.cowSupervisorNumberPool.Acquire()
		if !ok {
			log.Warnln("acquire supervisor room id failed")
		} else {
			r.Hall.cowRooms[id] = new(supervisorRoomT).CreateRoom(r.Hall, id, &waka.NiuniuRoomOption{
				Banker: 2,
				Mode:   r.Mode,
				Score:  r.Score,
			}, r.Creator)

			log.WithFields(logrus.Fields{
				"player": r.Creator,
				"score":  r.Score,
				"mode":   r.Mode,
				"id":     id,
			}).Debugln("supervisor room created")
		}
	}

	return true
}

func (r *supervisorRoomT) loopDeal4(loop func() bool) bool {
	if r.Distribution == nil {
		pokers := cow.Acquire5(len(r.Players))
		i := 0
		for _, player := range r.Players {
			pokers := pokers[i]
			player.Round.Pokers4 = append(player.Round.Pokers4, pokers[:4]...)
			player.Round.Pokers1 = pokers[4]
			i++
		}
	} else {
		for _, player := range r.Players {
			pokers := r.Distribution[player.Player]
			player.Round.Pokers4 = append(player.Round.Pokers4, pokers[:4]...)
			player.Round.Pokers1 = pokers[4]
		}
	}

	for _, player := range r.Players {
		r.Hall.sendNiuniuDeal4(player.Player, player.Round.Pokers4)
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.loop = loop

	return true
}

func (r *supervisorRoomT) loopGrab() bool {
	r.Step = "require_grab"
	for _, player := range r.Players {
		player.Round.Sent = false
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.loop = r.loopGrabContinue
	r.tick = buildTickNumber(
		6,
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
	r.Step = "require_commit_pokers"
	for _, player := range r.Players {
		var pokers []string
		pokers = append(pokers, player.Round.Pokers4...)
		pokers = append(pokers, player.Round.Pokers1)

		pokers, weight, pattern, _, err := cow.SearchBestPokerPattern(pokers, r.Mode)
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
	r.tick = buildTickNumber(
		3,
		func(number int32) {
			r.Hall.sendNiuniuCountdownForAll(r, number)
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

	bw, bp, br, _ := cow.GetPokersPattern(banker.Round.CommittedPokers, r.Mode)
	banker.Round.PokersPattern = bp
	banker.Round.PokersRate = int32(br)
	for _, player := range players {
		var applyRate int32
		var applySign int32
		pw, pp, pr, _ := cow.GetPokersPattern(player.Round.CommittedPokers, r.Mode)
		if bw >= pw {
			applyRate = int32(br)
			applySign = -1
		} else {
			applyRate = int32(pr)
			applySign = 1
		}

		banker.Round.Points += r.Score * player.Round.Rate * applyRate * applySign * (-1)
		player.Round.Points += r.Score * player.Round.Rate * applyRate * applySign

		player.Round.PokersPattern = pp
		player.Round.PokersRate = int32(pr)
	}

	var goldRoomCost []*database.CowGoldCost
	for _, player := range players {
		var c *database.CowGoldCost
		if player.Round.Points > 0 {
			c = &database.CowGoldCost{
				Victory: player.Player,
				Loser:   banker.Player,
				Number:  player.Round.Points * 100,
			}
		} else {
			c = &database.CowGoldCost{
				Victory: banker.Player,
				Loser:   player.Player,
				Number:  player.Round.Points * 100 * (-1),
			}
		}
		goldRoomCost = append(goldRoomCost, c)
	}
	err := database.CowGoldRoomSettle(r.Id, goldRoomCost)
	if err != nil {
		log.WithFields(logrus.Fields{
			"room_id": r.Id,
			"mode":    r.Mode,
			"score":   r.Score,
			"cost":    goldRoomCost,
			"err":     err,
		}).Warnln("supervisor cost failed")
	}

	clear := r.NiuniuRoundClear()
	for _, player := range r.Players {
		if err := database.CowAddGoldWarHistory(player.Player, r.Id, clear); err != nil {
			log.WithFields(logrus.Fields{
				"err": err,
			}).Warnln("add cow supervisor record failed")
		}
	}

	for _, player := range players {
		r.Hall.sendPlayerSecret(player.Player)
	}

	r.loop = r.loopSettleSuccess

	return true
}

func (r *supervisorRoomT) loopSettleSuccess() bool {
	r.Step = "round_clear"
	for _, player := range r.Players {
		player.Round.Sent = false
		player.Round.ContinueWithCommitted = false
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.loop = r.loopSettleSuccessContinue
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

func (r *supervisorRoomT) loopSettleSuccessContinue() bool {
	finally := true
	for _, player := range r.Players {
		if !player.Round.ContinueWithCommitted {
			finally = false
			if !player.Round.Sent {
				r.Hall.sendNiuniuSettleSuccess(player.Player)
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
	r.tick = nil
	r.loop = nil
	r.Step = ""
	r.Banker = 0
	r.Gaming = false
	r.Distribution = nil

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
		if player.Player.PlayerData().Money < r.LeaveMoney()*100 {
			delete(r.Players, player.Player)
			r.Seats.Return(player.Pos)
			r.Hall.players[player.Player].InsideCow = 0
			r.Hall.sendNiuniuRoomLeftByMoneyNotEnough(player.Player)
		}
	}
	for _, player := range r.Players {
		player.Round = supervisorRoundPlayerT{}
	}

	r.Hall.sendNiuniuUpdateRoomForAll(r)

	r.buildStart()

	return false
}
