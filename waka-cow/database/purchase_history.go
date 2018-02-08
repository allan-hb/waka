package database

import (
	"time"
)

// 牛牛场费消费记录
type CowRoomPurchaseHistory struct {
	// 主键
	Ref int32 `gorm:"index;primary_key;AUTO_INCREMENT"`
	// 谁
	Player Player `gorm:"index"`
	// 房间 ID
	Room int32 `gorm:"column:room_id"`
	// 金币数
	Number int32
	// 时间
	CreatedAt time.Time
}

func (CowRoomPurchaseHistory) TableName() string {
	return "cow_player_room_purchase_histories"
}

// ---------------------------------------------------------------------------------------------------------------------

// 牛牛金币消费记录
type CowGoldRoomPurchaseHistory struct {
	// 主键
	Ref int32 `gorm:"index;primary_key;AUTO_INCREMENT"`
	// 胜者
	Victory Player `gorm:"index"`
	// 败者
	Loser Player `gorm:"index"`
	// 房间 ID
	Room int32 `gorm:"column:room_id"`
	// 金币数
	Number int32
	// 时间
	CreatedAt time.Time
}

// ---------------------------------------------------------------------------------------------------------------------

// 五子棋消费记录
type GomokuPurchaseHistory struct {
	// 主键
	Ref int32 `gorm:"index;primary_key;AUTO_INCREMENT"`
	// 师傅
	Master Player `gorm:"index"`
	// 学生
	Student Player `gorm:"index"`
	// 金币数
	Number int32
	// 时间
	CreatedAt time.Time
}

// ---------------------------------------------------------------------------------------------------------------------

// 代理分成记录
type BonusHistory struct {
	// 主键
	Ref int32 `gorm:"index;primary_key;AUTO_INCREMENT"`
	// 代理
	Supervisor Player `gorm:"index"`
	// 玩家
	Player Player `gorm:"index"`
	// 金币数
	Number int32
	// 原因
	Reason string
	// 消费记录 ID
	Purchase int32
	// 时间
	CreatedAt time.Time
}

func newBonusByLever28(supervisor, player Player, number, purchase int32) *BonusHistory {
	bonus := newBonus(supervisor, player, number, purchase)
	bonus.Reason = "lever28"
	return bonus
}

func newBonusByRedPaperBag(supervisor, player Player, number, purchase int32) *BonusHistory {
	bonus := newBonus(supervisor, player, number, purchase)
	bonus.Reason = "red"
	return bonus
}

func newBonusByRoomCost(supervisor, player Player, number, purchase int32) *BonusHistory {
	bonus := newBonus(supervisor, player, number, purchase)
	bonus.Reason = "cow_cost"
	return bonus
}

func newBonus(supervisor, player Player, number, purchase int32) *BonusHistory {
	return &BonusHistory{
		Supervisor: supervisor,
		Player:     player,
		Number:     number,
		Purchase:   purchase,
		CreatedAt:  time.Now(),
	}
}
