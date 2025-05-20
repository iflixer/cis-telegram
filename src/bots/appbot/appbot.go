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
	btn1 := menu.Text("Сайт Пиратки")
	btn2 := menu.Text("Кино в telegram")
	btn3 := menu.Text("Скачать приложение")
	btn4 := menu.Text("Подписаться")
	btn5 := menu.Text("Связаться с нами")

	menu.Reply(
		menu.Row(btn1, btn2),
		menu.Row(btn3, btn4),
		menu.Row(btn5),
	)

	// selector.Inline(
	// 	selector.Row(btnPrev, btnNext),
	// )

	b.Handle("/start", func(c tele.Context) error {
		err := c.Send("Приветы. Я Кок, помощник Пиратки. Помогу найти что посмотреть", menu)
		//database.TelegramLogCreate(dbService, botId, c.Sender().ID, "/start", 1)
		return err
	})

	b.Handle(&btn1, func(c tele.Context) error {
		return c.Send("Смотри лучшие фильмы и сериалы на сайте https://piratka.me", menu)
	})

	b.Handle(&btn2, func(c tele.Context) error {
		return c.Send("Смотри лучшие фильмы и сериалы в telegram приложении https://t.me/piratka_me_app_bot/app?startapp=default", menu)
	})

	b.Handle(&btn3, func(c tele.Context) error {
		return c.Send("Скачай приложение для android https://apk.piratka.me/engine/ajax/controller.php?mod=download_apk", menu)
	})

	b.Handle(&btn4, func(c tele.Context) error {
		return c.Send("Загляни в канал", menu)
	})

	b.Handle(&btn5, func(c tele.Context) error {
		return c.Send("Напиши сообщение и мы его получим:", menu)
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
