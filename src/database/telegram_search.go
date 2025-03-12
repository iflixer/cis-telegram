package database

type TelegramSearch struct {
	dbService *Service
	ID        int
	BotID     int
	TgID      int64
	Text      string
	Found     int
}

func (c *TelegramSearch) TableName() string {
	return "telegram_search"
}

func NewTelegramSearch(dbService *Service, botID int, tgID int64, text string, found int) (search *TelegramSearch) {
	search = &TelegramSearch{
		dbService: dbService,
		BotID:     botID,
		TgID:      tgID,
		Text:      text,
		Found:     found,
	}
	dbService.DB.Save(search)
	return search
}
