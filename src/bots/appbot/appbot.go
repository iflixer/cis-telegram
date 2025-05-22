package appbot

import (
	"bytes"
	"cis-telegram/database"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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
		// URL:    &url.URL{Path: "/webhook"},
		// Poller: &webhook.Webhook{Listen: ":443"},
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
	btn4 := menu.Text("Подписаться на обновления")
	btn5 := menu.Text("Связаться с нами")

	menu.Reply(
		menu.Row(btn1, btn2),
		menu.Row(btn3, btn4),
		menu.Row(btn5),
	)

	// selector.Inline(
	// 	selector.Row(btnPrev, btnNext),
	// )

	inline := &tele.ReplyMarkup{}
	btn1inline := inline.URL("Сайт Пиратки", "https://piratka.me")
	btn2inline := inline.URL("Кино в telegram", "https://t.me/piratka_me_app_bot/app?startapp=default")
	btn3inline := inline.URL("Скачать приложение", "https://apk.piratka.me/engine/ajax/controller.php?mod=download_apk")
	btn4inline := inline.URL("Подписаться на обновления", "https://t.me/piratka_me")

	inline.Inline(
		inline.Row(btn1inline, btn2inline),
		inline.Row(btn3inline, btn4inline),
		inline.Row(btn5),
	)

	b.Handle("/start", func(c tele.Context) error {
		err := c.Send("Приветы. Я Кок, помощник Пиратки. Помогу найти что посмотреть", inline)
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
		return c.Send("Загляни в канал https://t.me/piratka_me", menu)
	})

	b.Handle(&btn5, func(c tele.Context) error {
		return c.Send("Напиши сообщение больше 10 символов и мы его получим:", menu)
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

		text := c.Text()
		if len(text) > 10 {
			dleComplain(text, c.Sender().ID)
			return c.Reply("Сообщение отправлено, мы его обязательно прочитаем", inline)
		}
		return c.Reply("Пардоне муа, не понимаю", inline)
	})

	go b.Start()
	return
}

func dleComplain(message string, tgID int64) (err error) {
	q := "https://odminko.printhouse.casa/engine/ajax/controller.php?mod=feedback&skip_captcha=fhduwiebu4377rdgegt"
	// q := "https://proxy.cis-dle.orb.local/engine/ajax/controller.php?mod=feedback&skip_captcha=fhduwiebu4377rdgegt"
	log.Println(q)

	data := url.Values{}
	data.Set("email", fmt.Sprintf("%d@telegram.me", tgID))
	data.Set("recip", "1")
	data.Set("subject", "tg bot")
	data.Set("message", message)

	req, err := http.NewRequest("POST", q, bytes.NewBufferString(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	log.Println(string(body))

	return
}
