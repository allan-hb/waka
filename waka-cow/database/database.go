package database

import (
	"fmt"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/sirupsen/logrus"

	"github.com/liuhan907/waka/waka-cow/conf"
)

var (
	log = logrus.WithFields(logrus.Fields{
		"pid":    os.Getpid(),
		"module": "cow.database",
	})

	mysql *gorm.DB
)

func init() {
	db, err := gorm.Open("mysql", fmt.Sprintf(`%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local`,
		conf.Option.Database.User, conf.Option.Database.Password, conf.Option.Database.Host, conf.Option.Database.Name))
	if err != nil {
		log.Fatalln(err)
	}

	mysql = db

	mysql.LogMode(true)

	tables := []interface{}{
		new(Configuration),
		new(PlayerData),
		new(FreezeData),
		new(TransactionData),

		new(CowHistory),
		new(GomokuHistory),
		new(Lever28History),
		new(RedHistory),
	}
	if conf.Option.Install.Reset {
		if err := mysql.DropTableIfExists(tables...).Error; err != nil {
			log.Panic(err)
		}
		if err := mysql.CreateTable(tables...).Error; err != nil {
			log.Panic(err)
		}
		if err := mysql.Exec("alter table players AUTO_INCREMENT = 100000;").Error; err != nil {
			log.Panic(err)
		}
	}
	if conf.Option.Install.Update {
		if err := mysql.AutoMigrate(tables...).Error; err != nil {
			log.Panic(err)
		}
	}

	if conf.Option.Install.Reset || conf.Option.Install.Update {
		systemPlayerCount := 0
		if err := mysql.Model(new(PlayerData)).Where("id = ?", DefaultSupervisor).Count(&systemPlayerCount).Error; err != nil {
			log.Panic(err)
		}
		if systemPlayerCount == 0 {
			if err := mysql.Create(&PlayerData{
				Id:            DefaultSupervisor,
				Nickname:      "__system",
				CreatedAt:     time.Now(),
				Vip:           time.Now(),
				Supervisor:    DefaultSupervisor,
				VictoryWeight: DefaultVictoryWeight,
			}).Error; err != nil {
				log.Panic(err)
			}
		}
	}

	RefreshConfiguration()

	recoverFreezeMoneyAfterLast()
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
		player, number, err := recoverFreezeMoney(ts, freeze.Id)
		if err != nil {
			log.WithFields(logrus.Fields{
				"freeze": freeze.Id,
				"player": freeze.Player,
				"number": freeze.Number,
				"err":    err,
			}).Warnln("recover freeze money failed")
			ts.Rollback()
			return
		}
		log.WithFields(logrus.Fields{
			"player": player,
			"number": number,
		}).Warnln("found freeze and recovered")
	}

	ts.Commit()
}
