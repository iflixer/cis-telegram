package settings

import (
	"cis-telegram/database"
	"log"
	"sync"
	"time"
)

type Service struct {
	mu           sync.RWMutex
	dbService    *database.Service
	updatePeriod time.Duration
	settings     []*Setting
}

type Setting struct {
	ID        int
	BotID     int
	Command   string
	Buttons   string
	Part      string
	Orderby   int
	Content   string
	Link      string
	Published bool
}

func (c *Setting) TableName() string {
	return "telegram_settings"
}

func (s *Service) Get(botId int, command, part string) (res []*Setting) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, g := range s.settings {
		if g.BotID == botId && g.Command == command && g.Part == part {
			res = append(res, g)
		}
	}
	return
}

func NewService(dbService *database.Service, updatePeriod int) (s *Service, err error) {
	s = &Service{
		dbService:    dbService,
		updatePeriod: time.Duration(updatePeriod),
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

func (s *Service) loadData() (err error) {
	var results []*Setting
	if err = s.dbService.DB.Where("published=1").Order("command,part,orderby").Find(&results).Error; err == nil {
		s.mu.Lock()
		s.settings = results
		s.mu.Unlock()
	}
	return
}
