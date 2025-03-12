package bot

import (
	telebot "cis-telegram/bots/loginbot"
	"cis-telegram/database"
	"cis-telegram/database/settings"
	"log"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Service struct {
	mu              sync.RWMutex
	dbService       *database.Service
	updatePeriod    time.Duration
	bots            map[int]*Bot
	settingsService *settings.Service
}

type Post struct {
	ID     int
	Title  string
	Slug   string
	URL    string
	Poster string
}

type Bot struct {
	ID              int
	Name            string
	BotURL          string
	AppURL          string
	Description     string
	Token           string
	GaTrackingID    string
	GaSecret        string
	SearchURL       string
	UpdatedAt       string
	Published       bool
	api             *tgbotapi.BotAPI  `gorm:"-"`
	quit            chan bool         `gorm:"-"`
	dbService       *database.Service `gorm:"-"`
	settingsService *settings.Service `gorm:"-"`
}

func (b *Bot) TableName() string {
	return "telegram_bot"
}

func (b *Bot) Register(dbService *database.Service, settingsService *settings.Service) (err error) {

	telebot.NewBot(dbService, b.ID, b.Token)

	b.api, err = tgbotapi.NewBotAPI(b.Token)
	if b.api == nil {
		log.Printf("Failed register bot %d, err: %s", b.ID, err.Error())
		return
	}
	b.dbService = dbService
	b.settingsService = settingsService
	b.quit = make(chan bool)
	b.listen()
	log.Printf("registered bot %d", b.ID)
	return
}

func (b *Bot) Kill() {
	b.quit <- true
}

func NewService(dbService *database.Service, settingsService *settings.Service, updatePeriod int) (s *Service, err error) {

	s = &Service{
		bots:            make(map[int]*Bot),
		dbService:       dbService,
		settingsService: settingsService,
		updatePeriod:    time.Duration(updatePeriod),
	}

	err = s.loadData()

	go s.loadWorker()

	return
}

func (s *Service) loadWorker() {
	for {
		time.Sleep(time.Second * s.updatePeriod)
		if err := s.loadData(); err != nil {
			log.Println(err)
		}
	}
}

func (s *Service) SendStart(botID int, chatID int64) (err error) {
	if bot, ok := s.bots[botID]; ok {
		err = bot.SendStart(chatID)
	} else {
		log.Println("bot not found:", botID)
	}
	return
}

func (s *Service) Send(botID int, chatID int64, msg string) (err error) {
	if bot, ok := s.bots[botID]; ok {
		err = bot.Send(chatID, msg)
	} else {
		log.Println("bot not found:", botID)
	}
	return
}

func (s *Service) loadData() (err error) {
	var results []*Bot
	if err = s.dbService.DB.Where("published=1").Find(&results).Error; err == nil {
		s.mu.Lock()
		for _, botNew := range results {
			if botOld, ok := s.bots[botNew.ID]; ok { // update old bot?
				log.Printf("update bot %d? ", botOld.ID)
				if botOld.UpdatedAt == botNew.UpdatedAt {
					log.Println("no")
					continue // don't need to restart bot
				}
				log.Println("yes")
				botOld.Kill()
			}
			log.Println("new bot")
			botNew.Register(s.dbService, s.settingsService)
			s.bots[botNew.ID] = botNew
		}
		log.Println("bots loaded:", len(s.bots))
		s.mu.Unlock()
	}
	return
}
