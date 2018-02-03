package database

import "time"

// 四张约战消费记录
type FourOrderRoomPurchaseHistory struct {
	// 主键
	Id int32 `gorm:"index;primary_key;AUTO_INCREMENT"`
	// 谁
	Player Player `gorm:"index"`
	// 房间 ID
	Room int32
	// 金币数
	Number int32
	// 时间
	CreatedAt time.Time
}

// ---------------------------------------------------------------------------------------------------------------------

// 四张代开消费记录
type FourPayForAnotherRoomPurchaseHistory struct {
	// 主键
	Id int32 `gorm:"index;primary_key;AUTO_INCREMENT"`
	// 谁
	Player Player `gorm:"index"`
	// 房间 ID
	Room int32
	// 金币数
	Number int32
	// 时间
	CreatedAt time.Time
}

// ---------------------------------------------------------------------------------------------------------------------
