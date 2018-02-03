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
	linq "gopkg.in/ahmetb/go-linq.v3"
)

type orderRoundPlayerT struct {
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

	// 本回合牌型
	PokersPattern string
	// 本回合牌型倍率
	PokersRate int32
	// 本回合得分
	PokersPoints int32
}

type orderPlayerT struct {
	Room *orderRoomT

	Player database.Player
	Pos    int32
	Ready  bool

	Round orderRoundPlayerT
}

func (player *orderPlayerT) NiuniuRoomData2RoomPlayer() (pb *waka.NiuniuRoomData2_RoomPlayer) {
	lost := false
	if player, being := player.Room.Hall.players[player.Player]; !being || player.Remote == "" {
		lost = true
	}
	return &waka.NiuniuRoomData2_RoomPlayer{
		Player: player.Room.Hall.ToPlayer(player.Player),
		Pos:    player.Pos,
		Ready:  player.Ready,
		Lost:   lost,
	}
}

type orderPlayerMapT map[database.Player]*orderPlayerT

func (players orderPlayerMapT) NiuniuRoomData2RoomPlayer() (pb []*waka.NiuniuRoomData2_RoomPlayer) {
	for _, player := range players {
		pb = append(pb, player.NiuniuRoomData2RoomPlayer())
	}
	return pb
}

func (players orderPlayerMapT) ToSlice() (d []*orderPlayerT) {
	for _, player := range players {
		d = append(d, player)
	}
	return d
}

// ---------------------------------------------------------------------------------------------------------------------

type orderRoomT struct {
	Hall *actorT

	Id        int32
	Option    *waka.NiuniuRoomOption
	Owner     database.Player
	Players   orderPlayerMapT
	Observers map[database.Player]database.Player

	loop func() bool
	tick func()

	Seats *tools.NumberPool

	Gaming      bool
	RoundNumber int32
	Step        string
	Banker      database.Player
}

// ---------------------------------------------------------------------------------------------------------------------

func (r *orderRoomT) CreateMoney() int32 {
	if r.Option.GetIsAA() {
		return int32(float64(r.Option.GetScore())*0.3+0.5) * r.Option.GetGames()
	} else {
		return int32(float64(r.Option.GetScore())*0.3+0.5) * r.Option.GetGames() * 5
	}
}

func (r *orderRoomT) EnterMoney() int32 {
	return r.CreateMoney()
}

func (r *orderRoomT) LeaveMoney() int32 {
	return r.EnterMoney()
}

func (r *orderRoomT) CostMoney() int32 {
	if r.Option.GetIsAA() {
		return int32(float64(r.Option.GetScore())*0.3+0.5) * r.Option.GetGames()
	} else {
		return int32(float64(r.Option.GetScore())*0.3+0.5) * r.Option.GetGames() * int32(len(r.Players))
	}
}

func (r *orderRoomT) GetType() waka.NiuniuRoomType {
	return waka.NiuniuRoomType_Order
}

func (r *orderRoomT) GetId() int32 {
	return r.Id
}

func (r *orderRoomT) GetOption() *waka.NiuniuRoomOption {
	return r.Option
}

func (r *orderRoomT) GetCreator() database.Player {
	return r.Owner
}

func (r *orderRoomT) GetOwner() database.Player {
	return r.Owner
}

func (r *orderRoomT) GetGaming() bool {
	return r.Gaming
}

func (r *orderRoomT) GetRoundNumber() int32 {
	return r.RoundNumber
}

func (r *orderRoomT) GetBanker() database.Player {
	return r.Banker
}

func (r *orderRoomT) GetPlayers() []database.Player {
	var d []database.Player
	linq.From(r.Players).SelectT(func(pair linq.KeyValue) database.Player { return pair.Key.(database.Player) }).ToSlice(&d)
	return d
}

func (r *orderRoomT) GetObservers() []database.Player {
	var d []database.Player
	linq.From(r.Observers).SelectT(func(pair linq.KeyValue) database.Player { return pair.Key.(database.Player) }).ToSlice(&d)
	return d
}

func (r *orderRoomT) NiuniuRoomData1() *waka.NiuniuRoomData1 {
	return &waka.NiuniuRoomData1{
		Id:         r.Id,
		Option:     r.GetOption(),
		Creator:    r.Owner.PlayerData().Nickname,
		Owner:      r.Owner.PlayerData().Nickname,
		Players:    int32(len(r.Players)),
		EnterMoney: r.EnterMoney(),
		LeaveMoney: r.LeaveMoney(),
		Gaming:     r.Gaming,
	}
}

func (r *orderRoomT) NiuniuRoomData2() *waka.NiuniuRoomData2 {
	return &waka.NiuniuRoomData2{
		Type:       waka.NiuniuRoomType_Order,
		Id:         r.Id,
		Option:     r.GetOption(),
		Creator:    r.Hall.ToPlayer(r.Owner),
		Owner:      r.Hall.ToPlayer(r.Owner),
		Players:    r.Players.NiuniuRoomData2RoomPlayer(),
		Observers:  r.Hall.ToPlayerMap(r.Observers),
		EnterMoney: r.EnterMoney(),
		LeaveMoney: r.LeaveMoney(),
		Gaming:     r.Gaming,
	}
}

func (r *orderRoomT) NiuniuRoundStatus(player database.Player) *waka.NiuniuRoundStatus {
	var pokers []string
	var players []*waka.NiuniuRoundStatus_RoundPlayer
	for id, playerData := range r.Players {
		players = append(players, &waka.NiuniuRoundStatus_RoundPlayer{
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
				(r.Step == "require_commit_pokers" || r.Step == "round_clear" || r.Step == "round_finally") {
				pokers = append(pokers, playerData.Round.Pokers1)
			}
		}
	}

	return &waka.NiuniuRoundStatus{
		Step:        r.Step,
		RoomId:      r.Id,
		RoundNumber: r.RoundNumber,
		Players:     players,
		Banker:      int32(r.Banker),
		Pokers:      pokers,
	}
}

func (r *orderRoomT) NiuniuGrabAnimation() *waka.NiuniuGrabAnimation {
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

func (r *orderRoomT) NiuniuRoundClear() *waka.NiuniuRoundClear {
	var players []*waka.NiuniuRoundClear_RoundClearPlayer
	for _, player := range r.Players {
		players = append(players, &waka.NiuniuRoundClear_RoundClearPlayer{
			Player:     r.Hall.ToPlayer(player.Player),
			Type:       player.Round.PokersPattern,
			Rate:       player.Round.PokersRate,
			ThisPoints: player.Round.PokersPoints,
			Pokers:     player.Round.CommittedPokers,
			Points:     player.Round.Points,
		})
	}
	return &waka.NiuniuRoundClear{Players: players, FinallyAt: time.Now().Format("2006-01-02 15:04:05")}
}

func (r *orderRoomT) NiuniuRoundFinally() *waka.NiuniuRoundFinally {
	var players []*waka.NiuniuRoundFinally_RoundFinallyPlayer
	for _, player := range r.Players {
		players = append(players, &waka.NiuniuRoundFinally_RoundFinallyPlayer{
			Player:    r.Hall.ToPlayer(player.Player),
			Points:    int32(player.Round.Points),
			Victories: player.Round.VictoriousNumber,
		})
	}
	return &waka.NiuniuRoundFinally{Players: players, FinallyAt: time.Now().Format("2006-01-02 15:04:05")}
}

// ---------------------------------------------------------------------------------------------------------------------

func (r *orderRoomT) Loop() {
	for {
		if r.loop == nil {
			return
		}
		if !r.loop() {
			return
		}
	}
}

func (r *orderRoomT) Tick() {
	if r.tick != nil {
		r.tick()
	}
}

func (r *orderRoomT) Left(player *playerT) {
	if !r.Gaming {
		if _, being := r.Observers[player.Player]; being {
			delete(r.Observers, player.Player)
			player.InsideCow = 0
		} else if roomPlayer, being := r.Players[player.Player]; being {
			if player.Player == r.Owner {
			} else {
				delete(r.Players, player.Player)
				player.InsideCow = 0
				r.Seats.Return(roomPlayer.Pos)
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

func (r *orderRoomT) Recover(player *playerT) {
	if _, being := r.Players[player.Player]; being {
		r.Players[player.Player].Round.Sent = false
	}

	r.Hall.sendNiuniuUpdateRoomForAll(r)
	if r.Gaming {
		r.Hall.sendNiuniuUpdateRound(player.Player, r)
		r.Loop()
	}
}

func (r *orderRoomT) CreateRoom(hall *actorT, id int32, option *waka.NiuniuRoomOption, creator database.Player) cowRoom {
	*r = orderRoomT{
		Hall:      hall,
		Id:        id,
		Option:    option,
		Owner:     creator,
		Players:   make(orderPlayerMapT, 5),
		Observers: map[database.Player]database.Player{},
		Seats:     tools.NewNumberPool(1, 5, false),
	}

	pos, _ := r.Seats.Acquire()

	r.Players[creator] = &orderPlayerT{
		Room:   r,
		Player: creator,
		Pos:    pos,
	}

	if creator.PlayerData().Money < r.CreateMoney()*100 {
		r.Hall.sendNiuniuCreateRoomFailed(creator, 1)
		return nil
	} else {
		r.Hall.cowRooms[id] = r

		r.Hall.players[creator].InsideCow = id

		r.Hall.sendNiuniuRoomCreated(creator)
		r.Hall.sendNiuniuRoomJoined(creator, r)
		r.Hall.sendNiuniuUpdateRoomForAll(r)

		return r
	}
}

func (r *orderRoomT) JoinRoom(player *playerT) {
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
			r.Players[player.Player] = &orderPlayerT{
				Room:   r,
				Player: player.Player,
				Pos:    seat,
			}
		}
	}

	player.InsideCow = r.GetId()

	r.Hall.sendNiuniuRoomJoined(player.Player, r)
	r.Hall.sendNiuniuUpdateRoomForAll(r)
}

func (r *orderRoomT) LeaveRoom(player *playerT) {
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
						r.Hall.sendNiuniuRoomLeft(player.Player)
					}
					for _, observer := range r.Observers {
						r.Hall.players[observer].InsideCow = 0
						r.Hall.sendNiuniuRoomLeft(observer)
					}
				} else {
					r.Hall.sendNiuniuUpdateRoomForAll(r)
				}
			}
		}
	}
}

func (r *orderRoomT) SwitchReady(player *playerT) {
	if !r.Gaming {
		if roomPlayer, being := r.Players[player.Player]; being {
			roomPlayer.Ready = !roomPlayer.Ready
			r.Hall.sendNiuniuUpdateRoomForAll(r)
		}
	}
}

func (r *orderRoomT) SwitchRole(player *playerT) {
	if !r.Gaming {
		if roomPlayer, being := r.Players[player.Player]; being {
			r.Seats.Return(roomPlayer.Pos)
			delete(r.Players, player.Player)
			r.Observers[player.Player] = player.Player

			if r.Owner == player.Player {
				r.Owner = 0
				if len(r.Players) > 0 {
					for _, player := range r.Players {
						r.Owner = player.Player
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
			r.Players[player.Player] = &orderPlayerT{
				Room:   r,
				Player: player.Player,
				Pos:    seat,
			}

			if r.Owner == 0 {
				r.Owner = player.Player
			}
		}

		r.Hall.sendNiuniuUpdateRoomForAll(r)
	}
}

func (r *orderRoomT) Dismiss(player *playerT) {
	if !r.Gaming {
		if r.Owner == player.Player {
			delete(r.Hall.cowRooms, r.Id)
			for _, player := range r.Players {
				r.Hall.players[player.Player].InsideCow = 0
				r.Hall.sendNiuniuRoomLeftByDismiss(player.Player)
			}
			for _, observer := range r.Observers {
				r.Hall.players[observer].InsideCow = 0
				r.Hall.sendNiuniuRoomLeftByDismiss(observer)
			}
		}
	}
}

func (r *orderRoomT) Start(player *playerT) {
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
			if r.Option.IsAA {
				for _, player := range r.Players {
					playerRoomCost = append(playerRoomCost, &database.CowPlayerRoomCost{
						Player: player.Player,
						Number: r.CostMoney() * 100,
					})
				}
			} else {
				playerRoomCost = append(playerRoomCost, &database.CowPlayerRoomCost{
					Player: r.Owner,
					Number: r.CostMoney() * 100,
				})
			}
			err := database.CowOrderRoomSettle(r.Id, playerRoomCost)
			if err != nil {
				log.WithFields(logrus.Fields{
					"room_id": r.Id,
					"option":  r.Option.String(),
					"cost":    playerRoomCost,
					"err":     err,
				}).Warnln("order cost settle failed")
				return
			}

			if r.Option.IsAA {
				for _, player := range r.Players {
					r.Hall.sendPlayerSecret(player.Player)
				}
			} else {
				r.Hall.sendPlayerSecret(r.Owner)
			}

			r.loop = r.loopStart

			r.Loop()
		}
	}
}

func (r *orderRoomT) SpecifyBanker(player *playerT, banker database.Player) {
	if r.Gaming {
		if _, being := r.Players[banker]; being {
			r.Banker = banker

			r.Loop()
		}
	}
}

func (r *orderRoomT) Grab(player *playerT, grab bool) {
	if r.Gaming {
		r.Players[player.Player].Round.Grab = grab
		r.Players[player.Player].Round.GrabCommitted = true

		r.Hall.sendNiuniuUpdateRoundForAll(r)

		r.Loop()
	}
}

func (r *orderRoomT) SpecifyRate(player *playerT, rate int32) {
	if r.Gaming {
		r.Players[player.Player].Round.Rate = rate
		r.Players[player.Player].Round.RateCommitted = true

		r.Hall.sendNiuniuUpdateRoundForAll(r)

		r.Loop()
	}
}

func (r *orderRoomT) CommitPokers(player *playerT, pokers []string) {
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

func (r *orderRoomT) ContinueWith(player *playerT) {
	if r.Gaming {
		r.Players[player.Player].Round.ContinueWithCommitted = true

		r.Loop()
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func (r *orderRoomT) loopStart() bool {
	r.Gaming = true
	r.RoundNumber = 1

	r.Hall.sendNiuniuStartedForAll(r, r.RoundNumber)

	if r.Option.Banker == 0 || r.Option.Banker == 1 {
		r.loop = r.loopSpecifyBanker
	} else if r.Option.Banker == 2 {
		r.loop = func() bool {
			return r.loopDeal4(r.loopGrab)
		}
	}

	return true
}

func (r *orderRoomT) loopSpecifyBanker() bool {
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

func (r *orderRoomT) loopSpecifyBankerContinue() bool {
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

func (r *orderRoomT) loopDeal4(loop func() bool) bool {
	pokers := cow.Acquire5(len(r.Players))
	i := 0
	for _, player := range r.Players {
		pokers := pokers[i]
		player.Round.Pokers4 = append(player.Round.Pokers4, pokers[:4]...)
		player.Round.Pokers1 = pokers[4]
		r.Hall.sendNiuniuDeal4(player.Player, player.Round.Pokers4)
		i++
	}

	r.Hall.sendNiuniuUpdateRoundForAll(r)

	r.loop = loop

	return true
}

func (r *orderRoomT) loopGrab() bool {
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

func (r *orderRoomT) loopGrabContinue() bool {
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

func (r *orderRoomT) loopGrabAnimation() bool {
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

func (r *orderRoomT) loopGrabAnimationContinue() bool {
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

func (r *orderRoomT) loopGrabSelect() bool {
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

func (r *orderRoomT) loopSpecifyRate() bool {
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

func (r *orderRoomT) loopSpecifyRateContinue() bool {
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

func (r *orderRoomT) loopDeal1() bool {
	r.Step = "require_commit_pokers"
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
	r.tick = buildTickNumber(
		5,
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

func (r *orderRoomT) loopCommitPokersContinue() bool {
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

func (r *orderRoomT) loopSettle() bool {
	if r.Players[r.Banker] == nil {
		for _, player := range r.Players {
			r.Banker = player.Player
			break
		}
	}

	banker := r.Players[r.Banker]

	var players []*orderPlayerT
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

func (r *orderRoomT) loopSettleSuccess() bool {
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

func (r *orderRoomT) loopSettleSuccessContinue() bool {
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
	r.loop = r.loopSelect

	return true
}

func (r *orderRoomT) loopSelect() bool {
	if r.RoundNumber < r.Option.GetGames() {
		r.loop = r.loopTransfer
	} else {
		r.loop = r.loopFinally
	}
	return true
}

func (r *orderRoomT) loopTransfer() bool {
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
		player.Round = orderRoundPlayerT{
			Points:           player.Round.Points,
			VictoriousNumber: player.Round.VictoriousNumber,
		}
	}

	r.Hall.sendNiuniuStartedForAll(r, r.RoundNumber)

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

func (r *orderRoomT) loopFinally() bool {
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
		if err := database.CowAddPlayerWarHistory(player.Player, r.Id, r.NiuniuRoundFinally()); err != nil {
			log.WithFields(logrus.Fields{
				"err": err,
			}).Warnln("add cow player history failed")
		}
	}

	return true
}

func (r *orderRoomT) loopFinallyContinue() bool {
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

func (r *orderRoomT) loopClean() bool {
	r.tick = nil
	r.loop = nil
	r.Step = ""
	r.Banker = 0
	r.Gaming = false

	for _, player := range r.Players {
		if r.Hall.players[player.Player].Remote == "" {
			if player.Player != r.Owner {
				delete(r.Players, player.Player)
				r.Hall.players[player.Player].InsideCow = 0
				r.Seats.Return(player.Pos)
			}
		}
		if player.Player.PlayerData().Money < r.LeaveMoney()*100 {
			if player.Player != r.Owner {
				delete(r.Players, player.Player)
				r.Hall.players[player.Player].InsideCow = 0
				r.Seats.Return(player.Pos)
			} else {
				delete(r.Hall.cowRooms, r.Id)
				for _, player := range r.Players {
					r.Hall.players[player.Player].InsideCow = 0
					r.Hall.sendNiuniuRoomLeft(player.Player)
				}
				for _, observer := range r.Observers {
					r.Hall.players[observer].InsideCow = 0
					r.Hall.sendNiuniuRoomLeft(observer)
				}
			}
		} else {
			player.Ready = false
		}
	}
	for _, player := range r.Players {
		player.Round = orderRoundPlayerT{}
	}

	r.Hall.sendNiuniuUpdateRoomForAll(r)

	return false
}
