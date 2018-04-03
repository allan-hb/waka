package database

import (
	"time"

	"github.com/jinzhu/gorm"
)

const (
	supervisor1 = 0.3
	supervisor2 = 0.09
	supervisor3 = 0.06
)

// 交易记录
type Transaction int32

// 交易记录数据
type TransactionData struct {
	// 主键
	Id Transaction `gorm:"index;unique;primary_key;AUTO_INCREMENT"`
	// 记录所属玩家
	Player Player
	// 交易对象
	Target Player
	// 数额
	Number int32
	// 类型
	// 0 未知
	// 1 支付
	// 2 收款
	Type int32
	// 原因
	Reason string
	// 创建时间
	CreatedAt time.Time
}

func (TransactionData) TableName() string {
	return "transactions"
}

// 交易
type playerTransaction struct {
	// 原因
	Reason string
	// 支付者
	Payer Player
	// 收款人
	Payee Player
	// 交易数额
	Number int32
	// 折损率
	Loss float64
	// 是否支付提成
	EnableTip bool
}

func buildTransaction(modifies []*modifyMoneyAction, transaction *playerTransaction) []*modifyMoneyAction {
	if transaction.Payer < DefaultSupervisor ||
		transaction.Payee < DefaultSupervisor ||
		transaction.Number <= 0 ||
		(transaction.EnableTip && (transaction.Loss < 0 || transaction.Loss > 1)) {
		return modifies
	}

	modifies = append(modifies, &modifyMoneyAction{
		Player: transaction.Payer,
		Number: transaction.Number * (-1),
		After: func(ts *gorm.DB, self *modifyMoneyAction) error {
			if err := ts.Create(&TransactionData{
				Player:    transaction.Payer,
				Target:    transaction.Payee,
				Number:    transaction.Number,
				Type:      1,
				Reason:    transaction.Reason + ".pay",
				CreatedAt: time.Now(),
			}).Error; err != nil {
				return err
			}
			return nil
		},
	})
	if transaction.EnableTip {
		supervisorPlayer1 := transaction.Payer.PlayerData().Supervisor
		supervisorPlayer2 := supervisorPlayer1.PlayerData().Supervisor
		supervisorPlayer3 := supervisorPlayer2.PlayerData().Supervisor

		number := int32(float64(transaction.Number)*(1-transaction.Loss) + 0.5)
		supervisorNumber1 := int32(float64(transaction.Number-number)*supervisor1 + 0.5)
		supervisorNumber2 := int32(float64(transaction.Number-number-supervisorNumber1)*supervisor2 + 0.5)
		supervisorNumber3 := int32(float64(transaction.Number-number-supervisorNumber1-supervisorNumber2)*supervisor3 + 0.5)
		systemNumber := transaction.Number - number - supervisorNumber1 - supervisorNumber2 - supervisorNumber3

		modifies = append(modifies, &modifyMoneyAction{
			Player: transaction.Payee,
			Number: number,
			After: func(ts *gorm.DB, self *modifyMoneyAction) error {
				if err := ts.Create(&TransactionData{
					Player:    transaction.Payee,
					Target:    transaction.Payer,
					Number:    number,
					Type:      2,
					Reason:    transaction.Reason + ".income",
					CreatedAt: time.Now(),
				}).Error; err != nil {
					return err
				}
				return nil
			},
		})
		modifies = append(modifies, &modifyMoneyAction{
			Player: supervisorPlayer1,
			Number: supervisorNumber1,
			After: func(ts *gorm.DB, self *modifyMoneyAction) error {
				if err := ts.Create(&TransactionData{
					Player:    supervisorPlayer1,
					Target:    transaction.Payer,
					Number:    supervisorNumber1,
					Type:      2,
					Reason:    transaction.Reason + ".tip1",
					CreatedAt: time.Now(),
				}).Error; err != nil {
					return err
				}
				return nil
			},
		})
		modifies = append(modifies, &modifyMoneyAction{
			Player: transaction.Payer.PlayerData().Supervisor.PlayerData().Supervisor,
			Number: supervisorNumber2,
			After: func(ts *gorm.DB, self *modifyMoneyAction) error {
				if err := ts.Create(&TransactionData{
					Player:    supervisorPlayer2,
					Target:    transaction.Payer,
					Number:    supervisorNumber2,
					Type:      2,
					Reason:    transaction.Reason + ".tip2",
					CreatedAt: time.Now(),
				}).Error; err != nil {
					return err
				}
				return nil
			},
		})
		modifies = append(modifies, &modifyMoneyAction{
			Player: transaction.Payer.PlayerData().Supervisor.PlayerData().Supervisor.PlayerData().Supervisor,
			Number: supervisorNumber3,
			After: func(ts *gorm.DB, self *modifyMoneyAction) error {
				if err := ts.Create(&TransactionData{
					Player:    supervisorPlayer3,
					Target:    transaction.Payer,
					Number:    supervisorNumber3,
					Type:      2,
					Reason:    transaction.Reason + ".tip3",
					CreatedAt: time.Now(),
				}).Error; err != nil {
					return err
				}
				return nil
			},
		})
		modifies = append(modifies, &modifyMoneyAction{
			Player: DefaultSupervisor,
			Number: systemNumber,
		})
	} else {
		modifies = append(modifies, &modifyMoneyAction{
			Player: transaction.Payee,
			Number: transaction.Number,
			After: func(ts *gorm.DB, self *modifyMoneyAction) error {
				if err := ts.Create(&TransactionData{
					Player:    transaction.Payee,
					Target:    transaction.Payer,
					Number:    transaction.Number,
					Type:      2,
					Reason:    transaction.Reason + ".income",
					CreatedAt: time.Now(),
				}).Error; err != nil {
					return err
				}
				return nil
			},
		})
	}

	return modifies
}
