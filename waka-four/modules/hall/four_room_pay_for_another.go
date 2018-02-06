package hall

import (
	"math"
	"reflect"
	"sort"

	"github.com/liuhan907/waka/waka-four/database"
	"github.com/liuhan907/waka/waka-four/modules/hall/tools"
	"github.com/liuhan907/waka/waka-four/modules/hall/tools/four"
	"github.com/liuhan907/waka/waka-four/proto"
	"github.com/sirupsen/logrus"
	linq "gopkg.in/ahmetb/go-linq.v3"
)

type fourPayForAnotherRoomPlayerRoundT struct {
	// 总分
	Points int32
	// 胜利的场数
	VictoriousNumber int32

	// 本阶段消息是否已发送
	Sent bool

	// 分配的手牌
	Pokers []string

	// 提交的配牌
	PokersFront  []string
	PokersBehind []string

	// 配牌已提交
	PokersCommitted bool
	// 阶段完成已提交
	ContinueWithCommitted bool

	// 本回合牌型
	PokersPatternFront  string
	PokersPatternBehind string
	// 本回合牌型权重
	PokersWeightFront  int32
	PokersWeightBehind int32
	// 本回合牌型得分
	PokersScoreFront  int32
	PokersScoreBehind int32
	// 本回合得分
	PokersPoints int32

	// 投票已提交
	VoteCommitted bool
	// 投票状态
	VoteStatus int32
}

type fourPayForAnotherRoomPlayerT struct {
	Room *fourPayForAnotherRoomT

	Player database.Player
	Pos    int32
	Ready  bool

	Round fourPayForAnotherRoomPlayerRoundT
}

func (player *fourPayForAnotherRoomPlayerT) FourRoom2Player() *four_proto.FourRoom2_Player {
	lost := false
	if player, being := player.Room.Hall.players[player.Player]; !being || player.Remote == "" {
		lost = true
	}
	if player.Room.Owner == player.Player {
		player.Ready = true
	}
	return &four_proto.FourRoom2_Player{
		PlayerId: int32(player.Player),
		Ready:    player.Ready,
		Lost:     lost,
	}
}

type fourPayForAnotherRoomPlayerMapT map[database.Player]*fourPayForAnotherRoomPlayerT

func (players fourPayForAnotherRoomPlayerMapT) FourRoom2Player() (d []*four_proto.FourRoom2_Player) {
	for _, player := range players {
		d = append(d, player.FourRoom2Player())
	}
	return d
}

func (players fourPayForAnotherRoomPlayerMapT) ToSlice() (d []*fourPayForAnotherRoomPlayerT) {
	for _, player := range players {
		d = append(d, player)
	}
	return d
}

// ---------------------------------------------------------------------------------------------------------------------

type fourPayForAnotherRoomT struct {
	Hall *actorT

	Id      int32
	Option  *four_proto.FourRoomOption
	Creator database.Player
	Owner   database.Player

	Players fourPayForAnotherRoomPlayerMapT

	loop func() bool
	tick func()

	Seats *tools.NumberPool

	Gaming      bool
	RoundNumber int32
	Step        string

	Compared *four_proto.FourCompare
	Settled  *four_proto.FourSettle

	VoteInitiator database.Player

	Cutter       database.Player
	CutCommitted bool
	CutPos       int32

	Distribution []map[database.Player][]string
	King         []database.Player

	LoopSwap func() bool
	StepSwap string
}

// ---------------------------------------------------------------------------------------------------------------------

func (r *fourPayForAnotherRoomT) CreateDiamonds() int32 {
	base := int32(0)
	switch r.Option.GetRounds() {
	case 8:
		base = 6
	case 16:
		base = 12
	case 24:
		base = 18
	default:
		return math.MaxInt32
	}
	return base * 8
}

func (r *fourPayForAnotherRoomT) EnterDiamonds() int32 {
	return 0
}

func (r *fourPayForAnotherRoomT) LeaveDiamonds(player database.Player) int32 {
	return 0
}

func (r *fourPayForAnotherRoomT) CostDiamonds() int32 {
	base := int32(0)
	switch r.Option.GetRounds() {
	case 8:
		base = 6
	case 16:
		base = 12
	case 24:
		base = 18
	default:
		return math.MaxInt32
	}
	return base * int32(len(r.Players))
}

func (r *fourPayForAnotherRoomT) GetId() int32 {
	return r.Id
}

func (r *fourPayForAnotherRoomT) GetOption() *four_proto.FourRoomOption {
	return r.Option
}

func (r *fourPayForAnotherRoomT) GetCreator() database.Player {
	return r.Creator
}

func (r *fourPayForAnotherRoomT) GetOwner() database.Player {
	return r.Owner
}

func (r *fourPayForAnotherRoomT) GetGaming() bool {
	return r.Gaming
}

func (r *fourPayForAnotherRoomT) GetRoundNumber() int32 {
	return r.RoundNumber
}

func (r *fourPayForAnotherRoomT) GetPlayers() []database.Player {
	var d []database.Player
	linq.From(r.Players).SelectT(func(pair linq.KeyValue) database.Player { return pair.Key.(database.Player) }).ToSlice(&d)
	return d
}

func (r *fourPayForAnotherRoomT) FourRoom1() *four_proto.FourRoom1 {
	return &four_proto.FourRoom1{
		RoomId:       r.Id,
		Option:       r.Option,
		CreatorId:    int32(r.Creator),
		OwnerId:      int32(r.Owner),
		PlayerNumber: int32(len(r.Players)),
		Gaming:       r.Gaming,
	}
}

func (r *fourPayForAnotherRoomT) FourRoom2() *four_proto.FourRoom2 {
	return &four_proto.FourRoom2{
		RoomId:    r.Id,
		Option:    r.Option,
		CreatorId: int32(r.Creator),
		OwnerId:   int32(r.Owner),
		Players:   r.Players.FourRoom2Player(),
		Gaming:    r.Gaming,
	}
}

func (r *fourPayForAnotherRoomT) FourCompare() *four_proto.FourCompare {
	return r.Compared
}

func (r *fourPayForAnotherRoomT) FourSettle() *four_proto.FourSettle {
	return r.Settled
}

func (r *fourPayForAnotherRoomT) FourFinallySettle() *four_proto.FourFinallySettle {
	settled := &four_proto.FourFinallySettle{}
	for _, player := range r.Players {
		settled.Players = append(settled.Players, &four_proto.FourFinallySettle_Player{
			PlayerId:      int32(player.Player),
			Nickname:      player.Player.PlayerData().Nickname,
			Score:         player.Round.Points,
			VictoryNumber: player.Round.VictoriousNumber,
		})
	}

	sort.Slice(settled.Players, func(i, j int) bool {
		return settled.Players[i].Score < settled.Players[j].Score
	})
	return settled
}

func (r *fourPayForAnotherRoomT) FourRoundStatus() *four_proto.FourRoundStatus {
	var players []*four_proto.FourRoundStatus_Player
	for _, player := range r.Players {
		players = append(players, &four_proto.FourRoundStatus_Player{
			PlayerId: int32(player.Player),
			Score:    player.Round.Points,
		})
	}
	return &four_proto.FourRoundStatus{
		Players: players,
	}
}

func (r *fourPayForAnotherRoomT) FourUpdateDismissVoteStatus() (status *four_proto.FourUpdateDismissVoteStatus, dismiss bool, finally bool) {
	status = &four_proto.FourUpdateDismissVoteStatus{}
	for _, player := range r.Players {
		status.Players = append(status.Players, &four_proto.FourUpdateDismissVoteStatus_Player{
			Id:     int32(player.Player),
			Status: player.Round.VoteStatus,
		})
	}

	dismiss = true
	finally = false
	for _, player := range r.Players {
		if player.Round.VoteStatus == 0 || player.Round.VoteStatus == 3 {
			dismiss = false
			if player.Round.VoteStatus == 3 {
				finally = true
			}
		}
	}

	return status, dismiss, finally
}

func (r *fourPayForAnotherRoomT) FourUpdateContinueWithStatus() *four_proto.FourUpdateContinueWithStatus {
	var players []*four_proto.FourUpdateContinueWithStatus_Player
	for _, player := range r.Players {
		d := &four_proto.FourUpdateContinueWithStatus_Player{
			Id: int32(player.Player),
		}
		if player.Round.ContinueWithCommitted {
			d.Status = 1
		} else {
			d.Status = 0
		}
		players = append(players, d)
	}

	return &four_proto.FourUpdateContinueWithStatus{r.Step, players}
}

func (r *fourPayForAnotherRoomT) BackendRoom() map[string]interface{} {
	var players []map[string]interface{}
	for _, player := range r.Players {
		playerData := player.Player.PlayerData()
		lost := false
		if playerData, being := r.Hall.players[player.Player]; !being || playerData.Remote == "" {
			lost = true
		}
		d := map[string]interface{}{
			"id":       player.Player,
			"nickname": playerData.Nickname,
			"head":     playerData.Head,
			"pos":      player.Pos,
			"ready":    player.Ready,
			"offline":  lost,
			"score":    player.Round.Points,
		}
		players = append(players, d)
	}
	return map[string]interface{}{
		"id":           r.Id,
		"status":       r.Step,
		"creator":      r.Creator,
		"owner":        r.Owner,
		"rounds":       r.Option.GetRounds(),
		"round_number": r.RoundNumber,
		"players":      players,
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func (r *fourPayForAnotherRoomT) Left(player *playerT) {
	if !r.Gaming {
		if roomPlayer, being := r.Players[player.Player]; being {
			delete(r.Players, player.Player)
			player.InsideFour = 0
			r.Seats.Return(roomPlayer.Pos)

			if player.Player == r.Owner {
				r.Owner = 0

				if len(r.Players) > 0 {
					for _, player := range r.Players {
						r.Owner = player.Player
						break
					}
				}
			}
		}
	} else {
	}

	r.Hall.sendFourUpdateRoomForAll(r)
}

func (r *fourPayForAnotherRoomT) Recover(player *playerT) {
	if playerData, being := r.Players[player.Player]; being {
		playerData.Round.Sent = false
	}

	r.Hall.sendFourUpdateRoomForAll(r)
	if r.Gaming {
		r.Hall.sendFourUpdateRound(player.Player, r)
		r.Loop()
	}
}

func (r *fourPayForAnotherRoomT) CreateRoom(hall *actorT, id int32, option *four_proto.FourRoomOption, creator database.Player) fourRoomT {
	*r = fourPayForAnotherRoomT{
		Hall:    hall,
		Id:      id,
		Option:  option,
		Creator: creator,
		Players: make(fourPayForAnotherRoomPlayerMapT, 8),
		Seats:   tools.NewNumberPool(1, 8, false),
	}

	if creator.PlayerData().Diamonds < r.CreateDiamonds() {
		r.Hall.sendFourCreateRoomFailed(creator, 1)
		return nil
	} else {
		r.Hall.fourRooms[id] = r

		r.Hall.sendFourCreateRoomSuccess(creator)

		return r
	}
}

func (r *fourPayForAnotherRoomT) JoinRoom(player *playerT) {
	if r.Gaming {
		r.Hall.sendFourJoinRoomFailed(player.Player, 5)
		return
	}

	if r.Option.GetScret() && player.Player != r.Creator {
		if !database.QueryPlayerCanJoin(r.Creator, player.Player) {
			r.Hall.sendFourJoinRoomFailed(player.Player, 3)
			return
		}
	}

	if player.Player.PlayerData().Diamonds < r.EnterDiamonds() {
		r.Hall.sendFourJoinRoomFailed(player.Player, 2)
		return
	}

	_, being := r.Players[player.Player]
	if being {
		r.Hall.sendFourJoinRoomFailed(player.Player, 4)
		return
	}

	pos, has := r.Seats.Acquire()
	if !has {
		r.Hall.sendFourJoinRoomFailed(player.Player, 0)
		return
	}

	r.Players[player.Player] = &fourPayForAnotherRoomPlayerT{
		Room:   r,
		Player: player.Player,
		Pos:    pos,
	}

	if player.Player.PlayerData().VictoryRate > 0 {
		r.King = append(r.King, player.Player)
	}

	player.InsideFour = r.GetId()

	if r.Owner == 0 {
		r.Owner = player.Player
	}

	r.Hall.sendFourJoinRoomSuccess(player.Player)
	r.Hall.sendFourUpdateRoomForAll(r)
}

func (r *fourPayForAnotherRoomT) LeaveRoom(player *playerT) {
	if !r.Gaming {
		if roomPlayer, being := r.Players[player.Player]; being {
			player.InsideFour = 0
			delete(r.Players, player.Player)
			r.Seats.Return(roomPlayer.Pos)

			r.Hall.sendFourLeftRoom(player.Player)

			if r.Owner == player.Player {
				r.Owner = 0
				if len(r.Players) > 0 {
					for _, player := range r.Players {
						r.Owner = player.Player
						break
					}
				}
			}

			r.Hall.sendFourUpdateRoomForAll(r)
		}
	}
}

func (r *fourPayForAnotherRoomT) SwitchReady(player *playerT) {
	if !r.Gaming {
		if roomPlayer, being := r.Players[player.Player]; being {
			roomPlayer.Ready = !roomPlayer.Ready
			r.Hall.sendFourUpdateRoomForAll(r)
		}
	}
}

func (r *fourPayForAnotherRoomT) Dismiss(player *playerT) {
	if !r.Gaming {
		if r.Creator == player.Player {
			delete(r.Hall.fourRooms, r.Id)
			for _, player := range r.Players {
				r.Hall.players[player.Player].InsideFour = 0
				r.Hall.sendFourLeftRoomByDismiss(player.Player)
			}
		}
	} else {
		r.VoteInitiator = player.Player
		r.LoopSwap = r.loop
		r.StepSwap = r.Step
		r.loop = r.loopVote

		r.Loop()
	}
}

func (r *fourPayForAnotherRoomT) Start(player *playerT) {
	if !r.Gaming {
		if len(r.Players) <= 1 {
			return
		}
		if r.Owner == player.Player {
			started := true
			for _, target := range r.Players {
				if target.Player == r.Owner {
					continue
				}
				if !target.Ready {
					started = false
				}
			}
			if !started {
				log.Debugln("not ready all")
				return
			}

			err := database.FourPayForAnotherRoomSettle(r.Id, &database.FourPlayerRoomCost{
				Player: r.Creator,
				Number: r.CostDiamonds(),
			})
			if err != nil {
				log.WithFields(logrus.Fields{
					"room_id": r.Id,
					"creator": r.Creator,
					"option":  r.Option.String(),
					"cost":    r.CostDiamonds(),
					"err":     err,
				}).Warnln("pay for another cost settle failed")
				return
			}

			r.loop = r.loopStart

			r.Loop()
		} else {
			log.Debugln("not has power")
		}
	}
}

func (r *fourPayForAnotherRoomT) Cut(player *playerT, pos int32) {
	if r.Gaming && r.Step == "cut_continue" {
		if player.Player != r.Cutter {
			return
		}

		r.CutCommitted = true
		r.CutPos = pos

		r.Loop()
	}
}

func (r *fourPayForAnotherRoomT) CommitPokers(player *playerT, front, behind []string) {
	if r.Gaming && r.Step == "commit_pokers" {
		log.WithFields(logrus.Fields{
			"player": player.Player,
			"front":  front,
			"behind": behind,
		}).Debug("commit pokers")

		playerData := r.Players[player.Player]

		origin := playerData.Round.Pokers
		sort.Slice(origin, func(i, j int) bool {
			return origin[i] < origin[j]
		})
		committed := append(append([]string{}, front...), behind...)
		sort.Slice(committed, func(i, j int) bool {
			return committed[i] < committed[j]
		})
		if !reflect.DeepEqual(origin, committed) {
			log.WithFields(logrus.Fields{
				"player":    player.Player,
				"origin":    origin,
				"committed": committed,
			}).Warnln("commit pokers not equal origin")
			return
		}

		playerData.Round.PokersFront = front
		playerData.Round.PokersBehind = behind
		playerData.Round.PokersCommitted = true
		playerData.Round.ContinueWithCommitted = true

		r.Hall.sendFourUpdateContinueWithStatusForAll(r)

		r.Loop()
	}
}

func (r *fourPayForAnotherRoomT) ContinueWith(player *playerT) {
	if r.Gaming && (r.Step == "compare_continue" || r.Step == "settle_continue" || r.Step == "cut_animation_continue") {
		r.Players[player.Player].Round.ContinueWithCommitted = true

		r.Hall.sendFourUpdateContinueWithStatusForAll(r)

		r.Loop()
	}
}

func (r *fourPayForAnotherRoomT) DismissVote(player *playerT, passing bool) {
	if r.Gaming && r.Step == "vote_continue" {
		if playerData, being := r.Players[player.Player]; being {
			if !playerData.Round.VoteCommitted {
				playerData.Round.VoteCommitted = true
				if passing {
					playerData.Round.VoteStatus = 2
				} else {
					playerData.Round.VoteStatus = 3
				}
			}

			r.Hall.sendFourUpdateDismissVoteStatusForAll(r)

			r.Loop()
		}
	}
}

func (r *fourPayForAnotherRoomT) SendMessage(player *playerT, messageType int32, text string) {
	for _, target := range r.Players {
		if target.Player != player.Player {
			r.Hall.sendFourReceivedMessage(target.Player, player.Player, messageType, text)
		}
	}
}

func (r *fourPayForAnotherRoomT) Loop() {
	for {
		if r.loop == nil {
			return
		}
		if !r.loop() {
			return
		}
	}
}

func (r *fourPayForAnotherRoomT) Tick() {
	if r.tick != nil {
		r.tick()
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func (r *fourPayForAnotherRoomT) loopStart() bool {
	r.Gaming = true
	r.RoundNumber = 1

	r.Hall.sendFourStartedForAll(r, r.RoundNumber)

	r.loop = r.loopDeal

	var king database.Player
	for _, k := range r.King {
		if _, being := r.Players[k]; being {
			king = k
			break
		}
	}
	if king != 0 {
		var players []database.Player
		linq.From(r.Players).SelectT(func(x linq.KeyValue) database.Player {
			return x.Key.(database.Player)
		}).ToSlice(&players)
		r.Distribution = four.Distributing(king, players, r.Option.GetRounds(), king.PlayerData().VictoryRate)
	}

	return true
}

func (r *fourPayForAnotherRoomT) loopDeal() bool {
	if r.Distribution == nil {
		pokers := four.Acquire4(len(r.Players))
		i := 0
		for _, player := range r.Players {
			player.Round.Pokers = append(player.Round.Pokers, pokers[i]...)
			i++
		}
	} else {
		roundMahjong := r.Distribution[r.RoundNumber-1]
		for _, player := range r.Players {
			player.Round.Pokers = roundMahjong[player.Player]
		}
	}

	r.loop = r.loopCommitPokers

	return true
}

func (r *fourPayForAnotherRoomT) loopCommitPokers() bool {
	r.Step = "commit_pokers"
	for _, player := range r.Players {
		player.Round.Sent = false
		player.Round.ContinueWithCommitted = false
	}

	r.loop = r.loopCommitPokersContinue

	return true
}

func (r *fourPayForAnotherRoomT) loopCommitPokersContinue() bool {
	finally := true
	for _, player := range r.Players {
		if !player.Round.PokersCommitted {
			finally = false
			if !player.Round.Sent {
				r.Hall.sendFourDeal(player.Player, player.Round.Pokers)
				r.Hall.sendFourUpdateContinueWithStatus(player.Player, r)
				player.Round.Sent = true
			}
		}
	}

	if !finally {
		return false
	}

	r.loop = r.loopCompare

	return true
}

func (r *fourPayForAnotherRoomT) loopCompare() bool {
	for _, player := range r.Players {
		w1, s1, p1, err := four.GetPattern(player.Round.PokersFront)
		if err != nil {
			log.WithFields(logrus.Fields{
				"player": player.Player,
				"front":  player.Round.PokersFront,
				"err":    err,
			}).Warnln("get front pokers pattern failed")
		} else {
			player.Round.PokersWeightFront = w1
			player.Round.PokersScoreFront = s1
			player.Round.PokersPatternFront = p1
		}
		w2, s2, p2, err := four.GetPattern(player.Round.PokersBehind)
		if err != nil {
			log.WithFields(logrus.Fields{
				"player": player.Player,
				"behind": player.Round.PokersBehind,
				"err":    err,
			}).Warnln("get behind pokers pattern failed")
		} else {
			player.Round.PokersWeightBehind = w2
			player.Round.PokersScoreBehind = s2
			player.Round.PokersPatternBehind = p2
		}
	}

	var fronts []*four_proto.FourCompare_Player
	var behinds []*four_proto.FourCompare_Player
	for _, player := range r.Players {
		fronts = append(fronts, &four_proto.FourCompare_Player{
			PlayerId: int32(player.Player),
			Pokers:   player.Round.PokersFront,
			Pattern:  player.Round.PokersPatternFront,
			Weight:   player.Round.PokersWeightFront,
		})
		behinds = append(behinds, &four_proto.FourCompare_Player{
			PlayerId: int32(player.Player),
			Pokers:   player.Round.PokersBehind,
			Pattern:  player.Round.PokersPatternBehind,
			Weight:   player.Round.PokersWeightBehind,
		})
	}
	sort.Slice(fronts, func(i, j int) bool {
		return fronts[i].Weight < fronts[j].Weight
	})
	sort.Slice(behinds, func(i, j int) bool {
		return behinds[i].Weight < behinds[j].Weight
	})

	r.Compared = &four_proto.FourCompare{}
	r.Compared.Players = append(r.Compared.Players, fronts...)
	r.Compared.Players = append(r.Compared.Players, behinds...)

	r.Step = "compare_continue"
	for _, player := range r.Players {
		player.Round.Sent = false
		player.Round.ContinueWithCommitted = false
	}

	r.loop = r.loopCompareContinue

	return true
}

func (r *fourPayForAnotherRoomT) loopCompareContinue() bool {
	finally := true
	for _, player := range r.Players {
		if !player.Round.ContinueWithCommitted {
			finally = false
			if !player.Round.Sent {
				r.Hall.sendFourCompare(player.Player, r)
				r.Hall.sendFourUpdateContinueWithStatus(player.Player, r)
				player.Round.Sent = true
			}
		}
	}

	if !finally {
		return false
	}

	r.loop = r.loopSettle

	return true
}

func (r *fourPayForAnotherRoomT) loopSettle() bool {
	players := r.Players.ToSlice()
	sort.Slice(players, func(i, j int) bool {
		return players[i].Player < players[j].Player
	})
	for i := 0; i < len(players); i++ {
		for k := i + 1; k < len(players); k++ {
			banker := &players[i].Round
			player := &players[k].Round

			switch {
			case banker.PokersWeightFront == player.PokersWeightFront && banker.PokersWeightBehind < player.PokersWeightBehind,
				banker.PokersWeightFront < player.PokersWeightFront && banker.PokersWeightBehind == player.PokersWeightBehind,
				banker.PokersWeightFront < player.PokersWeightFront && banker.PokersWeightBehind < player.PokersWeightBehind:
				banker.PokersPoints += (player.PokersScoreFront + player.PokersScoreBehind) * (-1)
				player.PokersPoints += player.PokersScoreFront + player.PokersScoreBehind

			case banker.PokersWeightFront == player.PokersWeightFront && banker.PokersWeightBehind > player.PokersWeightBehind,
				banker.PokersWeightFront > player.PokersWeightFront && banker.PokersWeightBehind == player.PokersWeightBehind,
				banker.PokersWeightFront > player.PokersWeightFront && banker.PokersWeightBehind > player.PokersWeightBehind:
				banker.PokersPoints += banker.PokersScoreFront + banker.PokersScoreBehind
				player.PokersPoints += (banker.PokersScoreFront + banker.PokersScoreBehind) * (-1)

			case banker.PokersWeightFront < player.PokersWeightFront && banker.PokersWeightBehind > player.PokersWeightBehind:
				banker.PokersPoints += player.PokersScoreFront * (-1)
				player.PokersPoints += player.PokersScoreFront
				banker.PokersPoints += banker.PokersScoreBehind
				player.PokersPoints += banker.PokersScoreBehind * (-1)
			case banker.PokersWeightFront > player.PokersWeightFront && banker.PokersWeightBehind < player.PokersWeightBehind:
				banker.PokersPoints += banker.PokersScoreFront
				player.PokersPoints += banker.PokersScoreFront * (-1)
				banker.PokersPoints += player.PokersScoreBehind * (-1)
				player.PokersPoints += player.PokersScoreBehind
			}
		}
	}

	sort.Slice(players, func(i, j int) bool {
		return players[i].Round.PokersPoints < players[j].Round.PokersPoints
	})

	for _, player := range players {
		player.Round.PokersPoints *= r.Option.GetRate()
	}

	r.Settled = &four_proto.FourSettle{}
	for _, player := range players {
		player.Round.Points += player.Round.PokersPoints
		if player.Round.PokersPoints > 0 {
			player.Round.VictoriousNumber++
		}
		r.Settled.Players = append(r.Settled.Players, &four_proto.FourSettle_Player{
			PlayerId:      int32(player.Player),
			Pokers:        append(append([]string{}, player.Round.PokersFront...), player.Round.PokersBehind...),
			PokersPattern: append(append([]string{}, player.Round.PokersPatternFront), player.Round.PokersPatternBehind),
			Score:         player.Round.PokersPoints,
		})
	}

	r.Hall.sendFourUpdateRoundForAll(r)

	r.Step = "settle_continue"
	for _, player := range r.Players {
		player.Round.Sent = false
		player.Round.ContinueWithCommitted = false
	}

	r.loop = r.loopSettleContinue

	return true
}

func (r *fourPayForAnotherRoomT) loopSettleContinue() bool {
	finally := true
	for _, player := range r.Players {
		if !player.Round.ContinueWithCommitted {
			finally = false
			if !player.Round.Sent {
				r.Hall.sendFourSettle(player.Player, r)
				r.Hall.sendFourUpdateContinueWithStatus(player.Player, r)
				player.Round.Sent = true
			}
		}
	}

	if !finally {
		return false
	}

	r.loop = r.loopSelect

	return true
}

func (r *fourPayForAnotherRoomT) loopSelect() bool {
	r.Cutter = database.Player(r.Settled.Players[0].PlayerId)

	r.Compared = nil
	r.Settled = nil
	r.Step = ""

	if r.RoundNumber < r.Option.GetRounds() {
		r.loop = r.loopTransfer
	} else {
		r.loop = r.loopFinallySettle
	}
	return true
}

func (r *fourPayForAnotherRoomT) loopTransfer() bool {
	r.RoundNumber++
	for _, player := range r.Players {
		player.Round = fourPayForAnotherRoomPlayerRoundT{
			Points:           player.Round.Points,
			VictoriousNumber: player.Round.VictoriousNumber,
		}
	}

	r.Hall.sendFourStartedForAll(r, r.RoundNumber)

	r.loop = r.loopCut

	return true
}

func (r *fourPayForAnotherRoomT) loopCut() bool {
	r.Step = "cut_continue"
	for _, player := range r.Players {
		player.Round.Sent = false
	}
	r.CutCommitted = false

	r.loop = r.loopCutContinue

	return true
}

func (r *fourPayForAnotherRoomT) loopCutContinue() bool {
	finally := true
	for _, player := range r.Players {
		if !r.CutCommitted {
			finally = false
			if !player.Round.Sent {
				r.Hall.sendFourRequireCut(player.Player, player.Player == r.Cutter)
				player.Round.Sent = true
			}
		}
	}

	if !finally {
		return false
	}

	r.loop = r.loopCutAnimation

	return true
}

func (r *fourPayForAnotherRoomT) loopCutAnimation() bool {
	r.Step = "cut_animation_continue"
	for _, player := range r.Players {
		player.Round.Sent = false
		player.Round.ContinueWithCommitted = false
	}

	r.loop = r.loopCutAnimationContinue

	return true
}

func (r *fourPayForAnotherRoomT) loopCutAnimationContinue() bool {
	finally := true
	for _, player := range r.Players {
		if !player.Round.ContinueWithCommitted {
			finally = false
			if !player.Round.Sent {
				r.Hall.sendFourRequireCutAnimation(player.Player, r.CutPos)
				r.Hall.sendFourUpdateContinueWithStatus(player.Player, r)
				player.Round.Sent = true
			}
		}
	}

	if !finally {
		return false
	}

	r.loop = r.loopDeal

	return true
}

func (r *fourPayForAnotherRoomT) loopFinallySettle() bool {
	for _, player := range r.Players {
		r.Hall.sendFourFinallySettle(player.Player, r)
	}

	for _, player := range r.Players {
		if err := database.FourAddPayForAnotherRoomWarHistory(player.Player, r.Id, r.FourFinallySettle()); err != nil {
			log.WithFields(logrus.Fields{
				"err": err,
			}).Warnln("add four pay for another room history failed")
		}
	}

	r.loop = r.loopClean

	return true
}

func (r *fourPayForAnotherRoomT) loopClean() bool {
	for _, player := range r.Players {
		if playerData, being := r.Hall.players[player.Player]; being {
			playerData.InsideFour = 0
		}
	}
	delete(r.Hall.fourRooms, r.Id)

	return false
}

func (r *fourPayForAnotherRoomT) loopVote() bool {
	r.Step = "vote_continue"
	for _, player := range r.Players {
		player.Round.Sent = false
		player.Round.VoteCommitted = false
		player.Round.VoteStatus = 0
	}

	r.loop = r.loopVoteContinue
	r.tick = buildTickNumber(kVoteSecond,
		func(number int32) {
			r.Hall.sendFourDismissVoteCountdownForAll(r, number)
		},
		func() {
			r.tick = nil
			for _, player := range r.Players {
				if !player.Round.VoteCommitted {
					player.Round.VoteCommitted = true
					player.Round.VoteStatus = 1
				}
			}
		},
		r.Loop,
	)

	return true
}

func (r *fourPayForAnotherRoomT) loopVoteContinue() bool {
	_, _, voteFinally := r.FourUpdateDismissVoteStatus()
	if !voteFinally {
		continueFinally := true
		for _, player := range r.Players {
			updated := player.Round.Sent
			if !player.Round.VoteCommitted {
				continueFinally = false
				if !player.Round.Sent {
					r.Hall.sendFourDismissRequireVote(player.Player, r.VoteInitiator)
					player.Round.Sent = true
				}
			}
			if !updated {
				r.Hall.sendFourUpdateDismissVoteStatus(player.Player, r)
			}
		}

		if !continueFinally {
			return false
		}
	}

	r.loop = r.loopVoteSettle

	return true
}

func (r *fourPayForAnotherRoomT) loopVoteSettle() bool {
	_, dismiss, _ := r.FourUpdateDismissVoteStatus()

	r.Hall.sendFourDismissFinallyForAll(r, dismiss)

	if !dismiss {
		r.tick = nil
		r.loop = r.LoopSwap
		r.Step = r.StepSwap
		r.LoopSwap = nil
		r.StepSwap = ""
		return true
	} else {
		for _, player := range r.Players {
			if playerData, being := r.Hall.players[player.Player]; being {
				playerData.InsideFour = 0
			}
			r.Hall.sendFourLeftRoomByDismiss(player.Player)
		}
		delete(r.Hall.fourRooms, r.Id)

		return false
	}
}
