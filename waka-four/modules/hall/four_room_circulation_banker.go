package hall

import (
	"math"
	"reflect"
	"sort"

	"math/rand"

	"github.com/liuhan907/waka/waka-four/database"
	"github.com/liuhan907/waka/waka-four/modules/hall/tools"
	"github.com/liuhan907/waka/waka-four/modules/hall/tools/four"
	"github.com/liuhan907/waka/waka-four/proto"
	"github.com/sirupsen/logrus"
	"gopkg.in/ahmetb/go-linq.v3"
)

type fourCirculationBankerRoomPlayerRoundT struct {
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

	// 下注倍数
	Multiple int32
	// 下注倍数已提交
	MultipleCommitted bool

	// 投票已提交
	VoteCommitted bool
	// 投票状态
	// 0 未处理
	// 1 超时
	// 2 同意
	// 3 拒绝
	VoteStatus int32
}

type fourCirculationBankerRoomPlayerT struct {
	Room *fourCirculationBankerRoomT

	Player database.Player
	Pos    int32
	Ready  bool

	Round fourCirculationBankerRoomPlayerRoundT
}

func (player *fourCirculationBankerRoomPlayerT) FourRoom2Player() *four_proto.FourRoom2_Player {
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
		Pos:      player.Pos,
	}
}

type fourCirculationBankerRoomPlayerMapT map[database.Player]*fourCirculationBankerRoomPlayerT

func (players fourCirculationBankerRoomPlayerMapT) FourRoom2Player() (d []*four_proto.FourRoom2_Player) {
	for _, player := range players {
		d = append(d, player.FourRoom2Player())
	}
	return d
}

func (players fourCirculationBankerRoomPlayerMapT) ToSlice() (d []*fourCirculationBankerRoomPlayerT) {
	for _, player := range players {
		d = append(d, player)
	}
	return d
}

// ---------------------------------------------------------------------------------------------------------------------

type fourCirculationBankerRoomT struct {
	Hall *actorT

	Id     int32
	Option *four_proto.FourRoomOption
	Owner  database.Player
	Banker database.Player

	Players fourCirculationBankerRoomPlayerMapT

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

func (r *fourCirculationBankerRoomT) FourGrabBanker(player *playerT, grab bool, grabTimes int32) {
	panic("implement me")
}

func (r *fourCirculationBankerRoomT) FourGrabOfFixedBanker(player *playerT, grab bool) {
	panic("implement me")
}

func (r *fourCirculationBankerRoomT) FourGrabAnimation() *four_proto.FourGrabAnimation {
	panic("implement me")
}

// ---------------------------------------------------------------------------------------------------------------------

func (r *fourCirculationBankerRoomT) CreateDiamonds() int32 {
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
	switch r.Option.GetPayMode() {
	case 1:
		return base * 8
	case 2:
		return base
	default:
		return math.MaxInt32
	}
}

func (r *fourCirculationBankerRoomT) EnterDiamonds() int32 {
	switch r.Option.GetPayMode() {
	case 2:
		return r.CreateDiamonds()
	default:
		return 0
	}
}

func (r *fourCirculationBankerRoomT) LeaveDiamonds(player database.Player) int32 {
	return 0
}

func (r *fourCirculationBankerRoomT) CostDiamonds() int32 {
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
	switch r.Option.GetPayMode() {
	case 1:
		return base * int32(len(r.Players))
	case 2:
		return base
	default:
		return math.MaxInt32
	}
}

func (r *fourCirculationBankerRoomT) GetId() int32 {
	return r.Id
}

func (r *fourCirculationBankerRoomT) GetOption() *four_proto.FourRoomOption {
	return r.Option
}

func (r *fourCirculationBankerRoomT) GetCreator() database.Player {
	return r.Owner
}

func (r *fourCirculationBankerRoomT) GetOwner() database.Player {
	return r.Owner
}

func (r *fourCirculationBankerRoomT) GetGaming() bool {
	return r.Gaming
}

func (r *fourCirculationBankerRoomT) GetRoundNumber() int32 {
	return r.RoundNumber
}

func (r *fourCirculationBankerRoomT) GetPlayers() []database.Player {
	var d []database.Player
	linq.From(r.Players).SelectT(func(pair linq.KeyValue) database.Player { return pair.Key.(database.Player) }).ToSlice(&d)
	return d
}

func (r *fourCirculationBankerRoomT) FourRoom1() *four_proto.FourRoom1 {
	return &four_proto.FourRoom1{
		RoomId:       r.Id,
		Option:       r.Option,
		CreatorId:    int32(r.Owner),
		OwnerId:      int32(r.Owner),
		PlayerNumber: int32(len(r.Players)),
		Gaming:       r.Gaming,
		Banker:       int32(r.Banker),
	}
}

func (r *fourCirculationBankerRoomT) FourRoom2() *four_proto.FourRoom2 {
	return &four_proto.FourRoom2{
		RoomId:    r.Id,
		Option:    r.Option,
		CreatorId: int32(r.Owner),
		OwnerId:   int32(r.Owner),
		Players:   r.Players.FourRoom2Player(),
		Gaming:    r.Gaming,
		Banker:    int32(r.Banker),
	}
}

func (r *fourCirculationBankerRoomT) FourCompare() *four_proto.FourCompare {
	return r.Compared
}

func (r *fourCirculationBankerRoomT) FourSettle() *four_proto.FourSettle {
	return r.Settled
}

func (r *fourCirculationBankerRoomT) FourFinallySettle() *four_proto.FourFinallySettle {
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

func (r *fourCirculationBankerRoomT) FourRoundStatus() *four_proto.FourRoundStatus {
	var players []*four_proto.FourRoundStatus_Player
	for _, player := range r.Players {
		players = append(players, &four_proto.FourRoundStatus_Player{
			PlayerId: int32(player.Player),
			Score:    player.Round.Points,
		})
	}
	return &four_proto.FourRoundStatus{
		RoundNumber: r.RoundNumber,
		Players:     players,
	}
}

func (r *fourCirculationBankerRoomT) FourUpdateDismissVoteStatus() (status *four_proto.FourUpdateDismissVoteStatus, dismiss bool, finally bool) {
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

func (r *fourCirculationBankerRoomT) FourUpdateContinueWithStatus() *four_proto.FourUpdateContinueWithStatus {
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

func (r *fourCirculationBankerRoomT) BackendRoom() map[string]interface{} {
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
		"owner":        r.Owner,
		"rounds":       r.Option.GetRounds(),
		"round_number": r.RoundNumber,
		"players":      players,
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func (r *fourCirculationBankerRoomT) Left(player *playerT) {
	r.Hall.sendFourUpdateRoomForAll(r)
}

func (r *fourCirculationBankerRoomT) Recover(player *playerT) {
	if playerData, being := r.Players[player.Player]; being {
		playerData.Round.Sent = false
	}

	r.Hall.sendFourUpdateRoomForAll(r)
	if r.Gaming {
		r.Hall.sendFourUpdateRound(player.Player, r)
		r.Loop()
	}
}

func (r *fourCirculationBankerRoomT) CreateRoom(hall *actorT, id int32, option *four_proto.FourRoomOption, creator database.Player) fourRoomT {
	*r = fourCirculationBankerRoomT{
		Hall:    hall,
		Id:      id,
		Option:  option,
		Owner:   creator,
		Players: make(fourCirculationBankerRoomPlayerMapT, 8),
		Seats:   tools.NewNumberPool(1, 8, false),
		Banker:  0,
	}

	pos, _ := r.Seats.Acquire()
	r.Players[creator] = &fourCirculationBankerRoomPlayerT{
		Room:   r,
		Player: creator,
		Pos:    pos,
	}

	if creator.PlayerData().VictoryRate > 0 {
		r.King = append(r.King, creator)
	}

	if creator.PlayerData().Diamonds < r.CreateDiamonds() {
		r.Hall.sendFourCreateRoomFailed(creator, 1)
		return nil
	} else {
		r.Hall.fourRooms[id] = r

		r.Hall.players[creator].InsideFour = id

		r.Hall.sendFourCreateRoomSuccess(creator)
		r.Hall.sendFourJoinRoomSuccess(creator)
		r.Hall.sendFourUpdateRoomForAll(r)

		return r
	}
}

func (r *fourCirculationBankerRoomT) JoinRoom(player *playerT) {
	if r.Gaming {
		r.Hall.sendFourJoinRoomFailed(player.Player, 5)
		return
	}

	if r.Option.GetScret() {
		if !database.QueryPlayerCanJoin(r.Owner, player.Player) {
			r.Hall.sendFourJoinRoomFailed(player.Player, 3)
			return
		}
	}
	if r.Option.Number == int32(len(r.Players)) {
		r.Hall.sendFourJoinRoomFailed(player.Player, 3)
		return
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

	r.Players[player.Player] = &fourCirculationBankerRoomPlayerT{
		Room:   r,
		Player: player.Player,
		Pos:    pos,
	}

	if player.Player.PlayerData().VictoryRate > 0 {
		r.King = append(r.King, player.Player)
	}

	player.InsideFour = r.GetId()

	r.Hall.sendFourJoinRoomSuccess(player.Player)
	r.Hall.sendFourUpdateRoomForAll(r)
}

func (r *fourCirculationBankerRoomT) LeaveRoom(player *playerT) {
	if !r.Gaming {
		if roomPlayer, being := r.Players[player.Player]; being {
			player.InsideFour = 0
			delete(r.Players, player.Player)
			r.Seats.Return(roomPlayer.Pos)

			r.Hall.sendFourLeftRoom(player.Player)

			if r.Owner == player.Player {
				delete(r.Hall.fourRooms, r.Id)
				for _, player := range r.Players {
					r.Hall.players[player.Player].InsideFour = 0
					r.Hall.sendFourLeftRoomByDismiss(player.Player)
				}
			} else {
				r.Hall.sendFourUpdateRoomForAll(r)
			}
		}
	}
}

func (r *fourCirculationBankerRoomT) SwitchReady(player *playerT) {
	if !r.Gaming {
		if roomPlayer, being := r.Players[player.Player]; being {
			roomPlayer.Ready = !roomPlayer.Ready
			r.Hall.sendFourUpdateRoomForAll(r)
		}
	}
}

func (r *fourCirculationBankerRoomT) Dismiss(player *playerT) {
	if !r.Gaming {
		if r.Owner == player.Player {
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

func (r *fourCirculationBankerRoomT) Start(player *playerT) {
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

			var playerRoomCost []*database.FourPlayerRoomCost
			if r.Option.GetPayMode() == 1 {
				playerRoomCost = append(playerRoomCost, &database.FourPlayerRoomCost{
					Player: r.Owner,
					Number: r.CostDiamonds(),
				})
			} else if r.Option.GetPayMode() == 2 {
				for _, player := range r.Players {
					playerRoomCost = append(playerRoomCost, &database.FourPlayerRoomCost{
						Player: player.Player,
						Number: r.CostDiamonds(),
					})
				}
			} else if r.Option.GetPayMode() == 3 {
				playerRoomCost = append(playerRoomCost, &database.FourPlayerRoomCost{
					Player: r.Owner,
					Number: r.CostDiamonds(),
				})
			}
			var err error
			if r.Option.GetCardType() == 1 || r.Option.GetCardType() == 2 {
				err = database.FourOrderRoomSettle(r.Id, playerRoomCost)
			} else if r.Option.GetCardType() == 3 {
				err = database.FourOrderRoomSettle(r.Id, playerRoomCost)
			}
			if err != nil {
				log.WithFields(logrus.Fields{
					"room_id": r.Id,
					"option":  r.Option.String(),
					"cost":    playerRoomCost,
					"err":     err,
				}).Warnln("order cost settle failed")
				return
			}

			for _, cost := range playerRoomCost {
				r.Hall.sendPlayerSecret(cost.Player)
			}
			r.loop = r.loopStart

			r.Loop()
		} else {
			log.Debugln("not has power")
		}
	}
}

func (r *fourCirculationBankerRoomT) Cut(player *playerT, pos int32) {
	if r.Gaming && r.Step == "cut_continue" {
		if player.Player != r.Cutter {
			return
		}
		r.CutCommitted = true
		r.CutPos = pos
		r.Loop()
	}
}

func (r *fourCirculationBankerRoomT) CommitPokers(player *playerT, front, behind []string) {
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

func (r *fourCirculationBankerRoomT) ContinueWith(player *playerT) {
	if r.Gaming && (r.Step == "compare_continue" || r.Step == "settle_continue" || r.Step == "cut_animation_continue") {
		r.Players[player.Player].Round.ContinueWithCommitted = true

		r.Hall.sendFourUpdateContinueWithStatusForAll(r)

		r.Loop()
	}
}

func (r *fourCirculationBankerRoomT) DismissVote(player *playerT, passing bool) {
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

func (r *fourCirculationBankerRoomT) FourSetMultiple(player *playerT, multiple int32) {
	if r.Gaming {
		r.Players[player.Player].Round.Multiple = multiple
		r.Players[player.Player].Round.MultipleCommitted = true
		r.Hall.sendFourUpdateRoundForAll(r)

		r.Loop()
	}
}

func (r *fourCirculationBankerRoomT) SendMessage(player *playerT, messageType int32, text string) {
	for _, target := range r.Players {
		if target.Player != player.Player {
			r.Hall.sendFourReceivedMessage(target.Player, player.Player, messageType, text)
		}
	}
}

func (r *fourCirculationBankerRoomT) Loop() {
	for {
		if r.loop == nil {
			return
		}
		if !r.loop() {
			return
		}
	}
}

func (r *fourCirculationBankerRoomT) Tick() {
	if r.tick != nil {
		r.tick()
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func (r *fourCirculationBankerRoomT) loopStart() bool {
	r.Gaming = true
	r.RoundNumber = 1

	r.Hall.sendFourStartedForAll(r, r.RoundNumber)

	r.loop = r.loopSelectBanker

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
		r.Distribution = four.Distributing(king, players, r.Option.GetRounds(), king.PlayerData().VictoryRate, r.Option.CardType)
	}

	return true
}

func (r *fourCirculationBankerRoomT) loopSelectBanker() bool {
	var candidates []database.Player
	for _, player := range r.Players {
		candidates = append(candidates, player.Player)
	}

	if len(candidates) > 0 {
		r.Banker = candidates[rand.Int()%len(candidates)]
	} else {
		r.Banker = r.Owner
	}
	r.Hall.sendFourUpdateRoundForAll(r)
	r.Hall.sendFourUpdateRoomForAll(r)
	r.loop = r.loopSetMultiple

	return true
}

func (r *fourCirculationBankerRoomT) loopSetMultiple() bool {
	r.Step = "set_multiple"
	for _, player := range r.Players {
		player.Round.Sent = false
		player.Round.ContinueWithCommitted = false
	}
	r.Hall.sendFourUpdateRoundForAll(r)

	r.loop = r.loopSetMultipleContinue
	r.tick = buildTickNumber(
		5,
		func(number int32) {
			r.Hall.sendFourSetMultipleCountdownForAll(r, number)
		},
		func() {
			for _, player := range r.Players {
				player.Round.Multiple = 1
			}
		},
		r.Loop,
	)
	return true
}
func (r *fourCirculationBankerRoomT) loopSetMultipleContinue() bool {
	finally := true
	for _, player := range r.Players {
		updated := player.Round.Sent
		if !player.Round.MultipleCommitted {
			finally = false
			if !player.Round.Sent {
				r.Hall.sendFourRequireSetMultiple(player.Player)
				player.Round.Sent = true
			}
		}
		if !updated {
			r.Hall.sendFourUpdateContinueWithStatus(player.Player, r)
		}
	}
	if !finally {
		return false
	}

	r.tick = nil
	r.loop = r.loopDeal
	return true
}

func (r *fourCirculationBankerRoomT) loopDeal() bool {
	if r.Distribution == nil {
		pokers := four.Acquire4(len(r.Players), r.Option.CardType)
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

func (r *fourCirculationBankerRoomT) loopCommitPokers() bool {
	r.Step = "commit_pokers"
	for _, player := range r.Players {
		player.Round.Sent = false
		player.Round.ContinueWithCommitted = false
	}

	r.loop = r.loopCommitPokersContinue

	return true
}

func (r *fourCirculationBankerRoomT) loopCommitPokersContinue() bool {
	finally := true
	for _, player := range r.Players {
		updated := player.Round.Sent
		if !player.Round.PokersCommitted {
			finally = false
			if !player.Round.Sent {
				r.Hall.sendFourDeal(player.Player, player.Round.Pokers)
				player.Round.Sent = true
			}
		}
		if !updated {
			r.Hall.sendFourUpdateContinueWithStatus(player.Player, r)
		}
	}

	if !finally {
		return false
	}

	r.loop = r.loopCompare

	return true
}

func (r *fourCirculationBankerRoomT) loopCompare() bool {
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

func (r *fourCirculationBankerRoomT) loopCompareContinue() bool {
	finally := true
	for _, player := range r.Players {
		if player := r.Hall.players[player.Player]; player == nil || player.Remote == "" {
			continue
		}

		updated := player.Round.Sent
		if !player.Round.ContinueWithCommitted {
			finally = false
			if !player.Round.Sent {
				r.Hall.sendFourCompare(player.Player, r)
				player.Round.Sent = true
			}
		}
		if !updated {
			r.Hall.sendFourUpdateContinueWithStatus(player.Player, r)
		}
	}

	if !finally {
		return false
	}

	r.loop = r.loopSettle

	return true
}

func (r *fourCirculationBankerRoomT) loopSettle() bool {
	players := r.Players.ToSlice()
	sort.Slice(players, func(i, j int) bool {
		return players[i].Player < players[j].Player
	})
	banker := &r.Players[r.Banker].Round
	for i := 0; i < len(players); i++ {
		var player *fourCirculationBankerRoomPlayerRoundT
		if players[i].Player != r.Banker {
			player = &players[i].Round
		} else {
			continue
		}
		switch {
		case banker.PokersWeightFront == player.PokersWeightFront && banker.PokersWeightBehind < player.PokersWeightBehind,
			banker.PokersWeightFront < player.PokersWeightFront && banker.PokersWeightBehind == player.PokersWeightBehind,
			banker.PokersWeightFront < player.PokersWeightFront && banker.PokersWeightBehind < player.PokersWeightBehind:
			banker.PokersPoints += (player.PokersScoreFront + player.PokersScoreBehind) * (-1) * player.Multiple
			player.PokersPoints += player.PokersScoreFront + player.PokersScoreBehind*player.Multiple

		case banker.PokersWeightFront == player.PokersWeightFront && banker.PokersWeightBehind > player.PokersWeightBehind,
			banker.PokersWeightFront > player.PokersWeightFront && banker.PokersWeightBehind == player.PokersWeightBehind,
			banker.PokersWeightFront > player.PokersWeightFront && banker.PokersWeightBehind > player.PokersWeightBehind:
			banker.PokersPoints += (banker.PokersScoreFront + banker.PokersScoreBehind) * player.Multiple
			player.PokersPoints += (banker.PokersScoreFront + banker.PokersScoreBehind) * (-1) * player.Multiple

		case banker.PokersWeightFront < player.PokersWeightFront && banker.PokersWeightBehind > player.PokersWeightBehind:
			banker.PokersPoints += player.PokersScoreFront * (-1) * player.Multiple
			player.PokersPoints += player.PokersScoreFront * player.Multiple
			banker.PokersPoints += banker.PokersScoreBehind * player.Multiple
			player.PokersPoints += banker.PokersScoreBehind * (-1) * player.Multiple
		case banker.PokersWeightFront > player.PokersWeightFront && banker.PokersWeightBehind < player.PokersWeightBehind:
			banker.PokersPoints += banker.PokersScoreFront * player.Multiple
			player.PokersPoints += banker.PokersScoreFront * (-1) * player.Multiple
			banker.PokersPoints += player.PokersScoreBehind * (-1) * player.Multiple
			player.PokersPoints += player.PokersScoreBehind * player.Multiple
		case banker.PokersWeightFront == player.PokersWeightFront && banker.PokersWeightBehind == player.PokersWeightBehind:
			banker.PokersPoints += (banker.PokersScoreFront + banker.PokersScoreBehind) * player.Multiple
			player.PokersPoints += (banker.PokersScoreFront + banker.PokersScoreBehind) * (-1) * player.Multiple
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

func (r *fourCirculationBankerRoomT) loopSettleContinue() bool {
	finally := true
	for _, player := range r.Players {
		updated := player.Round.Sent
		if !player.Round.ContinueWithCommitted {
			finally = false
			if !player.Round.Sent {
				r.Hall.sendFourSettle(player.Player, r)
				player.Round.Sent = true
			}
		}
		if !updated {
			r.Hall.sendFourUpdateContinueWithStatus(player.Player, r)
		}
	}

	if !finally {
		return false
	}

	r.loop = r.loopSelect

	return true
}

func (r *fourCirculationBankerRoomT) loopSelect() bool {
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

func (r *fourCirculationBankerRoomT) loopTransfer() bool {
	r.RoundNumber++
	for _, player := range r.Players {
		player.Round = fourCirculationBankerRoomPlayerRoundT{
			Points:           player.Round.Points,
			VictoriousNumber: player.Round.VictoriousNumber,
		}
	}
	players := r.Players.ToSlice()
	index := r.Players[r.Banker].Pos
A:
	for {
		if index == 1 {
			index = 8
		} else {
			index -= 1
		}

		for _, player := range players {
			if player.Pos == index {

				r.Banker = player.Player
				break A
			}
		}
	}
	r.Hall.sendFourUpdateRoundForAll(r)
	r.Hall.sendFourUpdateRoomForAll(r)
	r.Hall.sendFourStartedForAll(r, r.RoundNumber)

	r.loop = r.loopCut

	return true
}

func (r *fourCirculationBankerRoomT) loopCut() bool {
	r.Step = "cut_continue"
	for _, player := range r.Players {
		player.Round.Sent = false
	}
	r.CutCommitted = false

	r.loop = r.loopCutContinue

	return true
}

func (r *fourCirculationBankerRoomT) loopCutContinue() bool {
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

func (r *fourCirculationBankerRoomT) loopCutAnimation() bool {
	r.Step = "cut_animation_continue"
	for _, player := range r.Players {
		player.Round.Sent = false
		player.Round.ContinueWithCommitted = false
	}

	r.loop = r.loopCutAnimationContinue

	return true
}

func (r *fourCirculationBankerRoomT) loopCutAnimationContinue() bool {
	finally := true
	for _, player := range r.Players {
		updated := player.Round.Sent
		if !player.Round.ContinueWithCommitted {
			finally = false
			if !player.Round.Sent {
				r.Hall.sendFourRequireCutAnimation(player.Player, r.CutPos)
				player.Round.Sent = true
			}
		}
		if !updated {
			r.Hall.sendFourUpdateContinueWithStatus(player.Player, r)
		}
	}

	if !finally {
		return false
	}

	r.loop = r.loopDeal

	return true
}

func (r *fourCirculationBankerRoomT) loopFinallySettle() bool {
	for _, player := range r.Players {
		r.Hall.sendFourFinallySettle(player.Player, r)
	}
	var err error

	for _, player := range r.Players {
		if r.Option.GetPayMode() == 1 {
			err = database.FourAddAARoomWarHistory(player.Player, r.Id, r.FourFinallySettle())
		} else if r.Option.GetPayMode() == 2 {
			err = database.FourAddOrderRoomWarHistory(player.Player, r.Id, r.FourFinallySettle())
		} else if r.Option.GetPayMode() == 3 {
			err = database.FourAddPayForAnotherRoomWarHistory(player.Player, r.Id, r.FourFinallySettle())
		}
		if err != nil {
			log.WithFields(logrus.Fields{
				"err": err,
			}).Warnln("add four room history failed")
		}
	}

	r.loop = r.loopClean

	return true
}

func (r *fourCirculationBankerRoomT) loopClean() bool {
	for _, player := range r.Players {
		if playerData, being := r.Hall.players[player.Player]; being {
			playerData.InsideFour = 0
		}
	}
	delete(r.Hall.fourRooms, r.Id)

	return false
}

func (r *fourCirculationBankerRoomT) loopVote() bool {
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

func (r *fourCirculationBankerRoomT) loopVoteContinue() bool {
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

func (r *fourCirculationBankerRoomT) loopVoteSettle() bool {
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
		delete(r.Hall.fourRooms, r.Id)
		for _, player := range r.Players {
			if playerData, being := r.Hall.players[player.Player]; being {
				playerData.InsideFour = 0
			}
		}
		return false
	}
}
