package database

import (
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/liuhan907/waka/waka-four/proto"
)

type FourWarHistory struct {
	// 主键
	Id int32 `gorm:"index;primary_key;AUTO_INCREMENT"`
	// 玩家
	Player Player `gorm:"index"`
	// 房间模式
	// 0 约战
	// 1 代开
	Mode int32
	// 记录数据
	Payload []byte `gorm:"type:mediumblob"`
	// 时间
	CreatedAt time.Time
}

// 查询四张战绩
func FourQueryWarHistory(player Player, limit int32) ([]*four_proto.FourWarHistory, error) {
	var d []*FourWarHistory
	if err := mysql.Where("player = ?", player).Order("created_at desc").Limit(limit).Find(&d).Error; err != nil {
		return nil, err
	}
	var x []*four_proto.FourWarHistory
	for _, v := range d {
		k := &four_proto.FourWarHistory{}
		if err := proto.Unmarshal(v.Payload, k); err != nil {
			return nil, err
		}
		x = append(x, k)
	}
	return x, nil
}

// 添加四张约战场战绩
func FourAddOrderRoomWarHistory(player Player, room int32, finally *four_proto.FourFinallySettle) (e error) {
	return fourAddRoomWarHistory(0, player, room, finally)
}

// 添加四张约战场战绩
func FourAddAARoomWarHistory(player Player, room int32, finally *four_proto.FourFinallySettle) (e error) {
	return fourAddRoomWarHistory(0, player, room, finally)
}

// 添加四张代开场战绩
func FourAddPayForAnotherRoomWarHistory(player Player, room int32, finally *four_proto.FourFinallySettle) (e error) {
	return fourAddRoomWarHistory(1, player, room, finally)
}

// 添加四张战绩
func fourAddRoomWarHistory(mode int32, player Player, room int32, finally *four_proto.FourFinallySettle) (e error) {
	record := &four_proto.FourWarHistory{
		RoomId:    room,
		Type:      mode,
		Finally:   finally,
		CreatedAt: time.Now().Unix(),
	}
	d, err := proto.Marshal(record)
	if err != nil {
		return err
	}
	if err := mysql.Create(&FourWarHistory{
		Mode:      mode,
		Player:    player,
		Payload:   d,
		CreatedAt: time.Now(),
	}).Error; err != nil {
		return err
	}
	return nil
}
