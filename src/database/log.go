package database

type TelegramLog struct {
	dbService *Service
	ID        int
	BotID     int
	TgID      int64
	IsIncome  int
	Message   string
}

func (c *TelegramLog) TableName() string {
	return "telegram_log"
}

func TelegramLogCreate(dbService *Service, botID int, tgID int64, message string, isIncome int) (err error) {
	c := &TelegramLog{
		dbService: dbService,
		BotID:     botID,
		TgID:      tgID,
		IsIncome:  isIncome,
		Message:   message,
	}
	err = dbService.DB.Create(c).Error

	if isIncome == 1 {
		telegramUser := NewTelegramUser(dbService, botID, tgID)
		telegramUser.Upsert(message)
	}

	return
}
