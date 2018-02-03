package database

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	ErrFreezeRecovered = errors.New("freeze recovered")
	ErrMoneyNotEnough  = errors.New("money not enough")
)

// 冻结记录
type Freeze int32

// 冻结记录数据
type FreezeData struct {
	// 主键
	Ref Freeze `gorm:"index;primary_key;AUTO_INCREMENT"`
	// 冻结记录所属玩家
	Player Player `gorm:"index"`
	// 被冻结的钱
	Number int32
	// 已恢复
	Recovered bool
	// 创建时间
	CreatedAt time.Time
}

func (FreezeData) TableName() string {
	return "freezes"
}

func freezeMoney(ts *gorm.DB, ref Player, number int32) (Freeze, error) {
	if err := ts.Model(&PlayerData{}).Where("ref = ?", ref).Updates(
		map[string]interface{}{
			"money": gorm.Expr("money - ?", number),
		},
	).Error; err != nil {
		return 0, err
	}

	player := PlayerData{
		Ref: ref,
	}
	if err := ts.First(&player).Error; err != nil {
		return 0, err
	}

	if player.Money < 0 {
		return 0, ErrMoneyNotEnough
	}

	freezeData := FreezeData{
		Player:    ref,
		Number:    number,
		CreatedAt: time.Now(),
	}
	if err := ts.Create(&freezeData).Error; err != nil {
		return 0, err
	}

	return freezeData.Ref, nil
}

func recoverFreezeMoney(ts *gorm.DB, ref Freeze) (Player, error) {
	var freezeData FreezeData
	if err := ts.Where("ref = ?", ref).First(&freezeData).Error; err != nil {
		return 0, err
	}

	if freezeData.Recovered {
		return 0, ErrFreezeRecovered
	}

	if err := ts.Model(&PlayerData{}).Where("ref = ?", freezeData.Player).Updates(
		map[string]interface{}{
			"money": gorm.Expr("money + ?", freezeData.Number),
		},
	).Error; err != nil {
		return 0, err
	}

	freezeData.Recovered = true
	if err := ts.Save(&freezeData).Error; err != nil {
		return 0, err
	}

	return freezeData.Player, nil
}

type modifyMoneyAction struct {
	Player Player
	Number int32
	Before func(ts *gorm.DB, self *modifyMoneyAction) error
	After  func(ts *gorm.DB, self *modifyMoneyAction) error
}

func applyModifyMoneyAction(ts *gorm.DB, modifies []*modifyMoneyAction) error {
	for _, modify := range modifies {
		zeroCheck := false
		if modify.Number < 0 {
			zeroCheck = true
		}
		if modify.Before != nil {
			if err := modify.Before(ts, modify); err != nil {
				return err
			}
		}
		if err := modifyMoney(ts, modify.Player, modify.Number, zeroCheck); err != nil {
			return err
		}
		if modify.After != nil {
			if err := modify.After(ts, modify); err != nil {
				return err
			}
		}
	}

	return nil
}

func modifyMoney(ts *gorm.DB, player Player, money int32, zeroCheck bool) error {
	if money == 0 {
		return nil
	}
	if player == 0 {
		return nil
	}

	if err := ts.Model(&PlayerData{}).Where("ref = ?", player).Updates(
		map[string]interface{}{
			"money": gorm.Expr("money + ?", money),
		},
	).Error; err != nil {
		return err
	}
	if zeroCheck {
		player := PlayerData{
			Ref: player,
		}
		if err := ts.First(&player).Error; err != nil {
			return err
		}

		if player.Money < 0 {
			return ErrMoneyNotEnough
		}
	}
	return nil
}

func recoverFreezeMoneyAfterLast() {
	var freezes []*FreezeData

	ts := mysql.Begin()

	if err := ts.Where("recovered = ?", false).Find(&freezes).Error; err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Warnln("query last freeze money records failed")
		ts.Rollback()
		return
	}

	for _, freeze := range freezes {
		if _, err := recoverFreezeMoney(ts, freeze.Ref); err != nil {
			log.WithFields(logrus.Fields{
				"freeze": freeze.Ref,
				"player": freeze.Player,
				"number": freeze.Number,
				"err":    err,
			}).Warnln("recover freeze money failed")
			ts.Rollback()
			return
		}
	}

	ts.Commit()
}
