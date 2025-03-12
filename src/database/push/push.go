package push

import (
	"cis-telegram/database"
	"cis-telegram/database/bot"
	"log"
	"strings"
	"sync"
	"time"
)

type Service struct {
	mu         sync.RWMutex
	dbService  *database.Service
	botService *bot.Service
	state      string
}

type Push struct {
	ID         int
	BotID      int
	StartedAt  *time.Time
	EndAt      *time.Time
	AudienceID int
	Affected   int
	Command    string
	Status     string
}

func (c *Push) TableName() string {
	return "telegram_push"
}

type Log struct {
	ID            int
	PushID        int
	AudienceID    int
	AudienceName  string
	AudienceQuery string
	Affected      int
}

func (c *Log) TableName() string {
	return "telegram_push_log"
}

type Audience struct {
	ID    int
	Name  string
	Query string
}

func (c *Audience) TableName() string {
	return "telegram_audience"
}

func NewService(dbService *database.Service, botService *bot.Service) (s *Service, err error) {
	s = &Service{
		dbService:  dbService,
		botService: botService,
		state:      "sleep",
	}
	go s.workerSender()
	return
}

func (s *Service) workerSender() {
	for {
		//log.Println("push worker start")
		affected := 0
		var err error
		if affected, err = s.send(); err != nil {
			log.Println(err)
		}
		if affected > 0 {
			time.Sleep(time.Second)
		} else {
			time.Sleep(time.Second * 10)
		}
	}
}

func (s *Service) send() (affected int, err error) {
	var push *Push
	var audience *Audience
	var users []*database.TelegramUser
	if err = s.dbService.DB.Where("started_at<NOW() AND status='ready'").Limit(1).Find(&push).Error; err == nil {

		if push.ID == 0 {
			return
		}

		//helper.P(push)
		// get audience
		if err = s.dbService.DB.Where("id=?", push.AudienceID).Limit(1).First(&audience).Error; err != nil {
			return
		}
		// helper.P(audience)
		// get users
		if err = s.dbService.DB.Where("bot_id=? AND push_id != ? AND disabled=0", push.BotID, push.ID).Where(audience.Query).Limit(20).Find(&users).Error; err != nil {
			return
		}

		log.Println("Push found users:", len(users))

		if len(users) > 0 {
			for _, user := range users {
				log.Println("Push to telegram user:", user.TgID)
				switch push.Command {
				case "/start":
					err = s.botService.SendStart(push.BotID, user.TgID)
				default:
					err = s.botService.Send(push.BotID, user.TgID, push.Command)
				}
				//log.Println("sending error:", err)

				if err != nil && strings.Contains(err.Error(), "bot was blocked by the user") {
					log.Println("disable user in DB")
					user.Disabled = true
				}

				user.PushID = push.ID
				now := time.Now().UTC()
				user.PushTime = &now
				if err = s.dbService.DB.Save(user).Error; err != nil {
					return
				}
			}
			// save stats
			push.Affected += len(users)
			affected = push.Affected
			if err = s.dbService.DB.Save(push).Error; err != nil {
				return
			}
		} else {
			log.Println("push end")
			log := &Log{
				PushID:        push.ID,
				AudienceID:    audience.ID,
				AudienceName:  audience.Name,
				AudienceQuery: audience.Query,
				Affected:      push.Affected,
			}
			if err = s.dbService.DB.Create(log).Error; err != nil {
				return
			}
			push.Status = "done"
			now := time.Now().UTC()
			push.EndAt = &now
			if err = s.dbService.DB.Save(push).Error; err != nil {
				return
			}

		}

		// if we do not have additional buckets, save to log and update push status
		/*var userCountRest int64
		if err = s.dbService.DB.Table("telegram_user").Where("push_id != ?", push.ID).Count(&userCountRest).Error; err != nil {
			return
		}
		*/

	}
	return
}
