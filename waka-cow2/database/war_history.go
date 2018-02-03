package database

import (
	"time"

	"github.com/gogo/protobuf/proto"

	"github.com/liuhan907/waka/waka-cow2/proto"
)

type CowWarHistory struct {
	// 主键
	Id int32 `gorm:"index;primary_key;AUTO_INCREMENT"`
	// 玩家 ID
	PlayerId Player `gorm:"index"`
	// 房间模式
	// 0 约战
	// 1 代开
	Mode int32

	// 记录数据
	Payload []byte `gorm:"mediumblob"`

	// 时间
	CreatedAt time.Time
}

// 查询牛牛战绩
func CowQueryWarHistory(player Player, limit int32) ([]*cow_proto.NiuniuWarHistory, error) {
	var d []*CowWarHistory
	if err := mysql.Where("player_id = ?", player).Order("created_at desc").Limit(limit).Find(&d).Error; err != nil {
		return nil, err
	}
	var x []*cow_proto.NiuniuWarHistory
	for _, v := range d {
		k := &cow_proto.NiuniuWarHistory{}
		if err := proto.Unmarshal(v.Payload, k); err != nil {
			return nil, err
		}
		x = append(x, k)
	}
	return x, nil
}

// 添加牛牛约战战绩
func CowAddOrderWarHistory(player Player, roomId int32, finally *cow_proto.NiuniuRoundFinally) (e error) {
	return cow_protoAddWarHistory(0, player, roomId, finally)
}

// 添加牛牛代开战绩
func CowAddPayForAnotherWarHistory(player Player, roomId int32, finally *cow_proto.NiuniuRoundFinally) (e error) {
	return cow_protoAddWarHistory(1, player, roomId, finally)
}

// 添加牛牛战绩
func cow_protoAddWarHistory(mode int32, player Player, roomId int32, finally *cow_proto.NiuniuRoundFinally) (e error) {
	record := &cow_proto.NiuniuWarHistory{
		RoomId:    roomId,
		Mode:      mode,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	for _, player := range finally.Players {
		record.Players = append(record.Players, &cow_proto.NiuniuWarHistory_PlayerData{
			Player:    player.Player,
			Points:    player.Points,
			Victories: player.Victories,
		})
	}
	d, err := proto.Marshal(record)
	if err != nil {
		return err
	}
	if err := mysql.Create(&CowWarHistory{
		Mode:      mode,
		PlayerId:  player,
		Payload:   d,
		CreatedAt: time.Now(),
	}).Error; err != nil {
		return err
	}
	return nil
}
