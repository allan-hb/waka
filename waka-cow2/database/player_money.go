package database

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

var (
	ErrDiamondsNotEnough = errors.New("diamonds not enough")
)

type modifyDiamondsAction struct {
	Player Player
	Number int32
	Before func(ts *gorm.DB, self *modifyDiamondsAction) error
	After  func(ts *gorm.DB, self *modifyDiamondsAction) error
}

func applyModifyDiamondsAction(ts *gorm.DB, modifies []*modifyDiamondsAction) error {
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
		if err := modifyDiamonds(ts, modify.Player, modify.Number, zeroCheck); err != nil {
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

func modifyDiamonds(ts *gorm.DB, player Player, diamonds int32, zeroCheck bool) error {
	if diamonds == 0 {
		return nil
	}
	if player == 0 {
		return nil
	}

	if err := ts.Model(&PlayerData{}).Where("id = ?", player).Updates(
		map[string]interface{}{
			"diamonds": gorm.Expr("diamonds + ?", diamonds),
		},
	).Error; err != nil {
		return err
	}
	if zeroCheck {
		player := PlayerData{
			Id: player,
		}
		if err := ts.First(&player).Error; err != nil {
			return err
		}

		if player.Diamonds < 0 {
			return ErrDiamondsNotEnough
		}
	}
	return nil
}
