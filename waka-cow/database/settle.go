package database

// 牛牛场费结算记录
type CowOrderCostData struct {
	Player Player
	Number int32
}

// 牛牛场费结算
func CowOrderCostSettle(players []*CowOrderCostData) error {
	var modifies []*modifyMoneyAction

	for i := range players {
		player := players[i]

		modifies = buildTransaction(modifies, &playerTransaction{
			Reason:    "cow.order_cost",
			Payer:     player.Player,
			Payee:     DefaultSupervisor,
			Number:    player.Number,
			Loss:      0,
			EnableTip: true,
		})
	}

	ts := mysql.Begin()

	err := applyModifyMoneyActions(ts, modifies)
	if err != nil {
		ts.Rollback()
		return err
	}

	ts.Commit()

	playersLock.Lock()
	for _, player := range modifies {
		playerData, being := playersById[player.Player]
		if being {
			delete(playersById, player.Player)
			delete(playersByWechat, playerData.WechatUnionid)
			delete(playersByToken, playerData.Token)
		}
	}
	playersLock.Unlock()

	return nil
}

// 牛牛流水结算数据
type CowFlowingCostData struct {
	Victory Player
	Loser   Player
	Number  int32
}

// 牛牛流水结算
func CowFlowingCostSettle(players []*CowFlowingCostData) error {
	var modifies []*modifyMoneyAction

	for i := range players {
		player := players[i]

		modifies = buildTransaction(modifies, &playerTransaction{
			Reason:    "cow.flowing_cost",
			Payer:     player.Loser,
			Payee:     player.Victory,
			Number:    player.Number,
			Loss:      0.05,
			EnableTip: true,
		})
	}

	ts := mysql.Begin()

	err := applyModifyMoneyActions(ts, modifies)
	if err != nil {
		ts.Rollback()
		return err
	}

	ts.Commit()

	playersLock.Lock()
	for _, player := range modifies {
		playerData, being := playersById[player.Player]
		if being {
			delete(playersById, player.Player)
			delete(playersByWechat, playerData.WechatUnionid)
			delete(playersByToken, playerData.Token)
		}
	}
	playersLock.Unlock()

	return nil
}

// 冻结玩家金币
func FreezeMoney(player Player, number int32) (Freeze, error) {
	ts := mysql.Begin()

	freeze, err := freezeMoney(ts, player, number)
	if err != nil {
		ts.Rollback()
		return 0, err
	}

	ts.Commit()

	playersLock.Lock()
	playerData, being := playersById[player]
	if being {
		delete(playersById, player)
		delete(playersByWechat, playerData.WechatUnionid)
		delete(playersByToken, playerData.Token)
	}
	playersLock.Unlock()

	return freeze, nil
}

// 解冻玩家金币
func RecoverFreezeMoney(freeze Freeze) error {
	ts := mysql.Begin()

	player, _, err := recoverFreezeMoney(ts, freeze)
	if err != nil {
		ts.Rollback()
		return err
	}

	ts.Commit()

	playersLock.Lock()
	playerData, being := playersById[player]
	if being {
		delete(playersById, player)
		delete(playersByWechat, playerData.WechatUnionid)
		delete(playersByToken, playerData.Token)
	}
	playersLock.Unlock()

	return nil
}

// 红包创建者结算数据
type RedCreatorCost struct {
	Player Player
	Freeze Freeze
}

// 红包玩家抢红包数据
type RedGrabCost struct {
	Player Player
	Number int32
	Freeze Freeze
}

// 红包玩家赔付数据
type RedPayCost struct {
	Player Player
	Number int32
}

// 红包结算数据
type RedBagCost struct {
	Creator *RedCreatorCost
	Grabs   []*RedGrabCost
	Pays    []*RedPayCost
}

// 红包结算
func RedBagCostSettle(bag *RedBagCost) error {
	var modifies []*modifyMoneyAction
	var freezes []Freeze

	freezes = append(freezes, bag.Creator.Freeze)

	for i := range bag.Grabs {
		player := bag.Grabs[i]

		modifies = buildTransaction(modifies, &playerTransaction{
			Reason:    "red.grab",
			Payer:     bag.Creator.Player,
			Payee:     player.Player,
			Number:    player.Number,
			Loss:      0.05,
			EnableTip: true,
		})
		freezes = append(freezes, player.Freeze)
	}

	for i := range bag.Pays {
		player := bag.Pays[i]

		modifies = buildTransaction(modifies, &playerTransaction{
			Reason:    "red.pay",
			Payer:     player.Player,
			Payee:     bag.Creator.Player,
			Number:    player.Number,
			Loss:      0.05,
			EnableTip: true,
		})
	}

	ts := mysql.Begin()

	for _, freeze := range freezes {
		_, _, err := recoverFreezeMoney(ts, freeze)
		if err != nil {
			ts.Rollback()
			return err
		}
	}

	err := applyModifyMoneyActions(ts, modifies)
	if err != nil {
		ts.Rollback()
		return err
	}

	ts.Commit()

	playersLock.Lock()
	for _, player := range modifies {
		playerData, being := playersById[player.Player]
		if being {
			delete(playersById, player.Player)
			delete(playersByWechat, playerData.WechatUnionid)
			delete(playersByToken, playerData.Token)
		}
	}
	playersLock.Unlock()

	return nil
}

// 二八杠玩家凑红包数据
type Lever28Cost struct {
	Player Player
	Number int32
	Freeze Freeze
}

// 二八杠玩家抢红包数据
type Lever28Grab struct {
	Player Player
	Number int32
}

// 二八杠玩家赔付数据
type Lever28Pay struct {
	Payer  Player
	Payee  Player
	Number int32
}

// 二八杠结算数据
type Lever28BagCost struct {
	Costs []*Lever28Cost
	Grabs []*Lever28Grab
	Pays  []*Lever28Pay
}

// 二八杠结算
func Lever28Settle(bag *Lever28BagCost) error {
	var modifies []*modifyMoneyAction
	var freezes []Freeze

	for i := range bag.Costs {
		player := bag.Costs[i]

		modifies = buildTransaction(modifies, &playerTransaction{
			Reason:    "lever28.cost",
			Payer:     player.Player,
			Payee:     DefaultSupervisor,
			Number:    player.Number,
			EnableTip: false,
		})
		freezes = append(freezes, player.Freeze)
	}

	for i := range bag.Grabs {
		player := bag.Grabs[i]

		modifies = buildTransaction(modifies, &playerTransaction{
			Reason:    "lever28.grab",
			Payer:     DefaultSupervisor,
			Payee:     player.Player,
			Number:    player.Number,
			EnableTip: false,
		})
	}

	for i := range bag.Grabs {
		player := bag.Pays[i]

		modifies = buildTransaction(modifies, &playerTransaction{
			Reason:    "lever28.pay",
			Payer:     player.Payer,
			Payee:     player.Payee,
			Number:    player.Number,
			Loss:      0.05,
			EnableTip: true,
		})
	}

	ts := mysql.Begin()

	for _, freeze := range freezes {
		_, _, err := recoverFreezeMoney(ts, freeze)
		if err != nil {
			ts.Rollback()
			return err
		}
	}

	err := applyModifyMoneyActions(ts, modifies)
	if err != nil {
		ts.Rollback()
		return err
	}

	ts.Commit()

	playersLock.Lock()
	for _, player := range modifies {
		playerData, being := playersById[player.Player]
		if being {
			delete(playersById, player.Player)
			delete(playersByWechat, playerData.WechatUnionid)
			delete(playersByToken, playerData.Token)
		}
	}
	playersLock.Unlock()

	return nil
}

// 五子棋结算
func GomokuSettle(master, student Player, money int32) error {
	var modifies []*modifyMoneyAction

	modifies = buildTransaction(modifies, &playerTransaction{
		Reason:    "gomoku",
		Payer:     student,
		Payee:     master,
		Number:    money,
		EnableTip: false,
	})

	ts := mysql.Begin()

	err := applyModifyMoneyActions(ts, modifies)
	if err != nil {
		ts.Rollback()
		return err
	}

	ts.Commit()

	playersLock.Lock()
	for _, player := range modifies {
		playerData, being := playersById[player.Player]
		if being {
			delete(playersById, player.Player)
			delete(playersByWechat, playerData.WechatUnionid)
			delete(playersByToken, playerData.Token)
		}
	}
	playersLock.Unlock()

	return nil
}
