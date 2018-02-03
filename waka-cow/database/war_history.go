package database

import (
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/liuhan907/waka/waka-cow/proto"
)

type CowWarHistory struct {
	// 主键
	Ref int32 `gorm:"index;primary_key;AUTO_INCREMENT"`
	// 房间模式
	// 0 约战
	// 1 金币
	Mode int32
	// 玩家 ID
	Player Player `gorm:"index"`
	// 记录数据
	Payload []byte `gorm:"mediumblob"`
	// 时间
	CreatedAt time.Time
}

// 查询牛牛战绩
func CowQueryWarHistory(player Player, limit int32) ([]*waka.NiuniuRecord, error) {
	var d []*CowWarHistory
	if err := mysql.Where("player = ?", player).Order("created_at desc").Limit(limit).Find(&d).Error; err != nil {
		return nil, err
	}
	var x []*waka.NiuniuRecord
	for _, v := range d {
		k := &waka.NiuniuRecord{}
		if err := proto.Unmarshal(v.Payload, k); err != nil {
			return nil, err
		}
		x = append(x, k)
	}
	return x, nil
}

// 添加牛牛战绩
func CowAddPlayerWarHistory(player Player, roomId int32, finally *waka.NiuniuRoundFinally) (e error) {
	record := &waka.NiuniuRecord{
		RoomId:    roomId,
		Option:    0,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	for _, player := range finally.Players {
		record.Players = append(record.Players, &waka.NiuniuRecord_Player{
			Id:        player.Player.Id,
			Nickname:  player.Player.Nickname,
			Points:    player.Points,
			Victories: player.Victories,
		})
	}
	d, err := proto.Marshal(record)
	if err != nil {
		return err
	}
	if err := mysql.Create(&CowWarHistory{
		Mode:      0,
		Player:    player,
		Payload:   d,
		CreatedAt: time.Now(),
	}).Error; err != nil {
		return err
	}
	return nil
}

// 添加牛牛金币场战绩
func CowAddGoldWarHistory(player Player, roomId int32, finally *waka.NiuniuRoundClear) (e error) {
	record := &waka.NiuniuRecord{
		RoomId:    roomId,
		Option:    1,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	for _, player := range finally.Players {
		record.Players = append(record.Players, &waka.NiuniuRecord_Player{
			Id:       player.Player.Id,
			Nickname: player.Player.Nickname,
			Points:   player.ThisPoints,
		})
	}
	d, err := proto.Marshal(record)
	if err != nil {
		return err
	}
	if err := mysql.Create(&CowWarHistory{
		Mode:      1,
		Player:    player,
		Payload:   d,
		CreatedAt: time.Now(),
	}).Error; err != nil {
		return err
	}
	return nil
}

// ---------------------------------------------------------------------------------------------------------------------

type GomokuWarHistory struct {
	// 主键
	Ref int32 `gorm:"index;primary_key;AUTO_INCREMENT"`
	// 玩家
	Player Player `gorm:"index"`
	// 对手
	Opponent Player
	// 学费
	Cost int32
	// 时间
	CreatedAt time.Time
}

// 查询五子棋战绩
func GomokuQueryWarHistory(player Player, limit int32) ([]*GomokuWarHistory, error) {
	var d []*GomokuWarHistory
	if err := mysql.Where("player = ?", player).Order("created_at desc").Limit(limit).Find(&d).Error; err != nil {
		return nil, err
	}
	return d, nil
}

// 添加五子棋战绩
func GomokuAddWarHistory(master, student Player, cost int32) error {
	if err := mysql.Create(&GomokuWarHistory{
		Player:    master,
		Opponent:  student,
		Cost:      cost,
		CreatedAt: time.Now(),
	}).Error; err != nil {
		return err
	}
	if err := mysql.Create(&GomokuWarHistory{
		Player:    student,
		Opponent:  master,
		Cost:      cost * (-1),
		CreatedAt: time.Now(),
	}).Error; err != nil {
		return err
	}
	return nil
}

// ---------------------------------------------------------------------------------------------------------------------

type Lever28WarHistory struct {
	// 主键
	Ref int32 `gorm:"index;primary_key;AUTO_INCREMENT"`
	// 玩家
	Player Player `gorm:"index"`
	// 类型
	// 0 我发的
	// 1 我抢的
	Mode int32
	// 历史数据
	Bag []byte `gorm:"type:mediumblob"`
}

// 查询我发的红包历史
func Lever28QueryHandWarHistory(player Player, limit int32) ([]*waka.Lever28RedPaperBag3, error) {
	return lever28QueryWarHistory(player, 0, limit)
}

// 查询我抢的红包历史
func Lever28QueryGrabWarHistory(player Player, limit int32) ([]*waka.Lever28RedPaperBag3, error) {
	return lever28QueryWarHistory(player, 1, limit)
}

func lever28QueryWarHistory(player Player, mode int32, limit int32) ([]*waka.Lever28RedPaperBag3, error) {
	var d []*Lever28WarHistory
	if err := mysql.Where("player = ? and mode = ?", player, mode).Order("ref desc").Limit(limit).Find(&d).Error; err != nil {
		return nil, err
	}
	var r []*waka.Lever28RedPaperBag3
	for _, bag := range d {
		pb := &waka.Lever28RedPaperBag3{}
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
func Lever28AddHandWarHistory(player Player, bag *waka.Lever28RedPaperBag3) error {
	return lever28AddWarHistory(player, 0, bag)
}

// 添加我抢的红包记录
func Lever28AddGrabWarHistory(player Player, bag *waka.Lever28RedPaperBag3) error {
	return lever28AddWarHistory(player, 1, bag)
}

func lever28AddWarHistory(player Player, mode int32, bag *waka.Lever28RedPaperBag3) error {
	d, err := proto.Marshal(bag)
	if err != nil {
		return err
	}
	if err := mysql.Create(&Lever28WarHistory{
		Player: player,
		Mode:   mode,
		Bag:    d,
	}).Error; err != nil {
		return err
	}
	return nil
}

// ---------------------------------------------------------------------------------------------------------------------

type RedWarHistory struct {
	// 主键
	Ref int32 `gorm:"index;primary_key;AUTO_INCREMENT"`
	// 玩家
	Player Player `gorm:"index"`
	// 类型
	// 0 我发的
	// 1 我抢的
	Mode int32
	// 历史数据
	Bag []byte `gorm:"type:mediumblob"`
}

// 查询我发的红包历史
func RedQueryHandWarHistory(player Player, limit int32) ([]*waka.RedRedPaperBag3, error) {
	return redQueryWarHistory(player, 0, limit)
}

// 查询我抢的红包历史
func RedQueryGrabWarHistory(player Player, limit int32) ([]*waka.RedRedPaperBag3, error) {
	return redQueryWarHistory(player, 1, limit)
}

func redQueryWarHistory(player Player, mode int32, limit int32) ([]*waka.RedRedPaperBag3, error) {
	var d []*RedWarHistory
	if err := mysql.Where("player = ? and mode = ?", player, mode).Order("ref desc").Limit(limit).Find(&d).Error; err != nil {
		return nil, err
	}
	var r []*waka.RedRedPaperBag3
	for _, bag := range d {
		pb := &waka.RedRedPaperBag3{}
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
func RedAddHandWarHistory(player Player, bag *waka.RedRedPaperBag3) error {
	return redAddWarHistory(player, 0, bag)
}

// 添加我抢的红包记录
func RedAddGrabWarHistory(player Player, bag *waka.RedRedPaperBag3) error {
	return redAddWarHistory(player, 1, bag)
}

func redAddWarHistory(player Player, mode int32, bag *waka.RedRedPaperBag3) error {
	d, err := proto.Marshal(bag)
	if err != nil {
		return err
	}
	if err := mysql.Create(&RedWarHistory{
		Player: player,
		Mode:   mode,
		Bag:    d,
	}).Error; err != nil {
		return err
	}
	return nil
}
