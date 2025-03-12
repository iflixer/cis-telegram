package database

import (
	"context"
	"log"
	"time"

	"gorm.io/gorm/clause"
)

type TelegramUser struct {
	dbService      *Service
	ID             int
	BotID          int
	TgID           int64
	LastCommand    string
	Counter        int
	UpdatedAt      *time.Time
	LastActivityAt *time.Time
	PushID         int
	Disabled       bool
	PushTime       *time.Time
}

func (c *TelegramUser) TableName() string {
	return "telegram_user"
}

func NewTelegramUser(dbService *Service, botID int, tgID int64) (user *TelegramUser) {
	user = &TelegramUser{
		BotID: botID,
		TgID:  tgID,
	}
	if err := dbService.DB.Where("bot_id=? and tg_id=?", botID, tgID).Limit(1).Find(user).Error; err != nil {
		log.Println(err)
	}
	user.dbService = dbService
	return user
}

func (c *TelegramUser) GetPremiumStatus(dbService *Service, botID int, tgID int64) {
	user := &TelegramUser{}
	if err := dbService.DB.Where("bot_id=? and tg_id=?", botID, tgID).Limit(1).Find(user).Error; err != nil {
		log.Println(err)
	}
	c.dbService.DB.Save(c)
}

/*func (c *TelegramUser) Update(lastCommand string) {
	c.LastCommand = lastCommand
	c.dbService.DB.Save(c)
}*/

func (c *TelegramUser) Upsert(lastCommand string) {
	ctx := context.Background()
	c.Counter++
	c.LastCommand = lastCommand
	now := time.Now().UTC()
	c.LastActivityAt = &now
	if len(c.LastCommand) > 50 {
		c.LastCommand = c.LastCommand[:50]
	}
	c.dbService.DB.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(c)
	c.dbService.DB.Where("bot_id=? and tg_id=?", c.BotID, c.TgID).First(c)
}
