package appbot

import (
	"cis-telegram/database"
	"fmt"
	"time"

	tele "gopkg.in/telebot.v4"
)

type TeleBot struct {
	Token string
}

func NewBot(dbService *database.Service, botId int, token string) (err error) {
	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return
	}
	//middleware.Recover()
	//b.Use(middleware.FlixLogger())
	//b.Use(middleware.AutoRespond())

	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	//selector := &tele.ReplyMarkup{}
	// Reply buttons.
	btn1 := menu.Text("b1")
	btn2 := menu.Text("b2")
	btn3 := menu.Text("b3")
	btn4 := menu.Text("b4")

	menu.Reply(
		menu.Row(btn1, btn2),
		menu.Row(btn3, btn4),
	)

	// selector.Inline(
	// 	selector.Row(btnPrev, btnNext),
	// )

	b.Handle("/start", func(c tele.Context) error {
		err := c.Send("Приветы. Я Кок, помощник Пиратки. Помогу найти что посмотреть", menu)
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, "/start", 1)
		return err
	})

	b.Handle(&btn1, func(c tele.Context) error {
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, "[btnHelp]", 1)
		return c.Send("т1", menu)
	})

	b.Handle("/hello", func(c tele.Context) error {
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, "/hello", 1)
		return c.Send("Hello!" + fmt.Sprintf("%d", c.Sender().ID))
	})

	b.Handle("/me", func(c tele.Context) error {
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, "/hello", 1)
		return c.Send("ID:" + fmt.Sprintf("%d", c.Sender().ID))
	})

	b.Handle(tele.OnText, func(c tele.Context) error {
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, c.Text(), 0)
		return c.Reply("Пардоне муа, не понимаю", menu)
	})

	go b.Start()
	return
}
