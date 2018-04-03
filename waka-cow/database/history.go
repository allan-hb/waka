package database

import (
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/liuhan907/waka/waka-cow/proto"
)

type CowHistory struct {
	// 主键
	Id int32 `gorm:"index;unique;primary_key;AUTO_INCREMENT"`
	// 玩家
	Player Player `gorm:"index"`
	// 记录数据
	Payload []byte `gorm:"type:mediumblob"`
	// 时间
	CreatedAt time.Time
}

func (CowHistory) TableName() string {
	return "history_cow"
}

// 查询牛牛战绩
func CowQueryHistory(player Player, limit int32) ([]*cow_proto.NiuniuHistory, error) {
	var d []*CowHistory
	if err := mysql.Where("player = ?", player).Order("created_at desc").Limit(limit).Find(&d).Error; err != nil {
		return nil, err
	}
	var x []*cow_proto.NiuniuHistory
	for _, v := range d {
		k := &cow_proto.NiuniuHistory{}
		if err := proto.Unmarshal(v.Payload, k); err != nil {
			return nil, err
		}
		x = append(x, k)
	}
	return x, nil
}

// 添加牛牛约战战绩
func CowAddHistory(player Player, roomId int32, roomType cow_proto.NiuniuRoomType, finally *cow_proto.NiuniuGameFinally) (e error) {
	record := &cow_proto.NiuniuHistory{
		Type:   roomType,
		RoomId: roomId,
	}
	for _, player := range finally.Players {
		record.Players = append(record.Players, &cow_proto.NiuniuHistory_PlayerData{
			Id:        player.Player,
			Points:    player.Points,
			Victories: player.Victories,
		})
	}
	d, err := proto.Marshal(record)
	if err != nil {
		return err
	}
	if err := mysql.Create(&CowHistory{
		Player:    player,
		Payload:   d,
		CreatedAt: time.Now(),
	}).Error; err != nil {
		return err
	}
	return nil
}

// 添加牛牛流水战绩
func CowAddFlowingHistory(player Player, roomId int32, finally *cow_proto.NiuniuRoundClear) (e error) {
	record := &cow_proto.NiuniuHistory{
		Type:   cow_proto.NiuniuRoomType_Flowing,
		RoomId: roomId,
	}
	for _, player := range finally.Players {
		record.Players = append(record.Players, &cow_proto.NiuniuHistory_PlayerData{
			Id:     player.Player,
			Points: player.Points,
		})
	}
	d, err := proto.Marshal(record)
	if err != nil {
		return err
	}
	if err := mysql.Create(&CowHistory{
		Player:    player,
		Payload:   d,
		CreatedAt: time.Now(),
	}).Error; err != nil {
		return err
	}
	return nil
}

// ---------------------------------------------------------------------------------------------------------------------

type GomokuHistory struct {
	// 主键
	Id int32 `gorm:"index;unique;primary_key;AUTO_INCREMENT"`
	// 玩家
	Player Player `gorm:"index"`
	// 对手
	Opponent Player
	// 学费
	Cost int32
	// 时间
	CreatedAt time.Time
}

func (GomokuHistory) TableName() string {
	return "history_gomoku"
}

// 查询五子棋战绩
func GomokuQueryHistory(player Player, limit int32) ([]*GomokuHistory, error) {
	var d []*GomokuHistory
	if err := mysql.Where("player = ?", player).Order("created_at desc").Limit(limit).Find(&d).Error; err != nil {
		return nil, err
	}
	return d, nil
}

// 添加五子棋战绩
func GomokuAddWarHistory(master, student Player, cost int32) error {
	ts := mysql.Begin()
	if err := ts.Create(&GomokuHistory{
		Player:    master,
		Opponent:  student,
		Cost:      cost,
		CreatedAt: time.Now(),
	}).Error; err != nil {
		ts.Rollback()
		return err
	}
	if err := ts.Create(&GomokuHistory{
		Player:    student,
		Opponent:  master,
		Cost:      cost * (-1),
		CreatedAt: time.Now(),
	}).Error; err != nil {
		ts.Rollback()
		return err
	}
	ts.Commit()
	return nil
}

// ---------------------------------------------------------------------------------------------------------------------

type Lever28History struct {
	// 主键
	Id int32 `gorm:"index;unique;primary_key;AUTO_INCREMENT"`
	// 玩家
	Player Player `gorm:"index"`
	// 类型
	// 0 我发的
	// 1 我抢的
	Mode int32
	// 历史数据
	Bag []byte `gorm:"type:mediumblob"`
}

func (Lever28History) TableName() string {
	return "history_lever28"
}

// 查询我发的红包历史
func Lever28QueryHandHistory(player Player, limit int32) ([]*cow_proto.Lever28BagClear, error) {
	return lever28QueryHistory(player, 0, limit)
}

// 查询我抢的红包历史
func Lever28QueryGrabHistory(player Player, limit int32) ([]*cow_proto.Lever28BagClear, error) {
	return lever28QueryHistory(player, 1, limit)
}

func lever28QueryHistory(player Player, mode int32, limit int32) ([]*cow_proto.Lever28BagClear, error) {
	var d []*Lever28History
	if err := mysql.Where("player = ? and mode = ?", player, mode).Order("id desc").Limit(limit).Find(&d).Error; err != nil {
		return nil, err
	}
	var r []*cow_proto.Lever28BagClear
	for _, bag := range d {
		pb := &cow_proto.Lever28BagClear{}
		err := proto.Unmarshal(bag.Bag, pb)
		if err != nil {
			log.WithField("err", err).Warnln("unmarshal lever28 history failed")
		} else {
			r = append(r, pb)
		}
	}
	return r, nil
}

// 添加我发的红包记录
func Lever28AddHandHistory(player Player, bag *cow_proto.Lever28BagClear) error {
	return lever28AddHistory(player, 0, bag)
}

// 添加我抢的红包记录
func Lever28AddGrabHistory(player Player, bag *cow_proto.Lever28BagClear) error {
	return lever28AddHistory(player, 1, bag)
}

func lever28AddHistory(player Player, mode int32, bag *cow_proto.Lever28BagClear) error {
	d, err := proto.Marshal(bag)
	if err != nil {
		return err
	}
	if err := mysql.Create(&Lever28History{
		Player: player,
		Mode:   mode,
		Bag:    d,
	}).Error; err != nil {
		return err
	}
	return nil
}

// ---------------------------------------------------------------------------------------------------------------------

type RedHistory struct {
	// 主键
	Id int32 `gorm:"index;unique;primary_key;AUTO_INCREMENT"`
	// 玩家
	Player Player `gorm:"index"`
	// 类型
	// 0 我发的
	// 1 我抢的
	Mode int32
	// 历史数据
	Bag []byte `gorm:"type:mediumblob"`
}

func (RedHistory) TableName() string {
	return "history_red"
}

// 查询我发的红包历史
func RedQueryHandHistory(player Player, limit int32) ([]*cow_proto.RedBagClear, error) {
	return redQueryHistory(player, 0, limit)
}

// 查询我抢的红包历史
func RedQueryGrabHistory(player Player, limit int32) ([]*cow_proto.RedBagClear, error) {
	return redQueryHistory(player, 1, limit)
}

func redQueryHistory(player Player, mode int32, limit int32) ([]*cow_proto.RedBagClear, error) {
	var d []*RedHistory
	if err := mysql.Where("player = ? and mode = ?", player, mode).Order("id desc").Limit(limit).Find(&d).Error; err != nil {
		return nil, err
	}
	var r []*cow_proto.RedBagClear
	for _, bag := range d {
		pb := &cow_proto.RedBagClear{}
		err := proto.Unmarshal(bag.Bag, pb)
		if err != nil {
			log.WithField("err", err).Warnln("unmarshal red history failed")
		} else {
			r = append(r, pb)
		}
	}
	return r, nil
}

// 添加我发的红包记录
func RedAddHandHistory(player Player, bag *cow_proto.RedBagClear) error {
	return redAddHistory(player, 0, bag)
}

// 添加我抢的红包记录
func RedAddGrabHistory(player Player, bag *cow_proto.RedBagClear) error {
	return redAddHistory(player, 1, bag)
}

func redAddHistory(player Player, mode int32, bag *cow_proto.RedBagClear) error {
	d, err := proto.Marshal(bag)
	if err != nil {
		return err
	}
	if err := mysql.Create(&RedHistory{
		Player: player,
		Mode:   mode,
		Bag:    d,
	}).Error; err != nil {
		return err
	}
	return nil
}
