package telebot

import (
	"cis-telegram/database"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	tele "gopkg.in/telebot.v4"
)

type TeleBot struct {
	Token string
}

func NewBot(dbService *database.Service, botId int, token string) {
	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}
	//middleware.Recover()
	//b.Use(middleware.FlixLogger())
	//b.Use(middleware.AutoRespond())

	menu := &tele.ReplyMarkup{ResizeKeyboard: true}
	//selector := &tele.ReplyMarkup{}
	// Reply buttons.
	btnHelp := menu.Text("Помощь")
	btnSubscribe1 := menu.Text("1 мес = 1⭐")
	btnSubscribe2 := menu.Text("6 мес = 2⭐")
	btnSubscribe3 := menu.Text("1 год = 3⭐")
	btnStatus := menu.Text("Статус подписки")
	// btnPrev := selector.Data("⬅", "prev")
	// btnNext := selector.Data("➡", "next")

	menu.Reply(
		menu.Row(btnSubscribe1, btnSubscribe2, btnSubscribe3),
		menu.Row(btnStatus),
		menu.Row(btnHelp),
	)

	// selector.Inline(
	// 	selector.Row(btnPrev, btnNext),
	// )

	b.Handle("/start", func(c tele.Context) error {
		err := c.Send("Приветы. Я Юнга, помощник Пиратки. Хочешь закинуть нам дублонов?", menu)
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, "/start", 1)
		return err
	})

	// On reply button pressed (message)
	b.Handle(&btnHelp, func(c tele.Context) error {
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, "[btnHelp]", 1)
		return c.Send("Этот бот позволяет оплатить премиум доступ к сервису.\nОплата производится в звездах.\nПотрачанные звезды не возвращаются.\nПодарки - не отдарки! :)", menu)
	})

	// On inline button pressed (callback)
	b.Handle(&btnSubscribe1, func(c tele.Context) error {
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, "[btnSubscribe1]", 1)
		log.Println("payment request!")
		invoice := &tele.Invoice{
			Title:       "Премиум доступ на 1 месяц",
			Description: "отключение рекламы",
			Payload:     "subscription1",
			Currency:    "XTR",
			Total:       1,
			Prices: []tele.Price{
				{Label: "adv free", Amount: 1},
			},
		}
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, "invoice1m", 0)
		return c.Send(invoice)
	})

	b.Handle(&btnSubscribe2, func(c tele.Context) error {
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, "[btnSubscribe2]", 1)
		log.Println("payment request!")
		invoice := &tele.Invoice{
			Title:       "Премиум доступ на 6 месяцев",
			Description: "отключение рекламы",
			Payload:     "subscription2",
			Currency:    "XTR",
			Total:       2,
			Prices: []tele.Price{
				{Label: "adv free", Amount: 2},
			},
		}
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, "invoice6m", 0)
		return c.Send(invoice)
	})

	b.Handle(&btnSubscribe3, func(c tele.Context) error {
		log.Println("payment request!")
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, "[btnSubscribe3]", 1)
		invoice := &tele.Invoice{
			Title:       "Премиум доступ на 1 год",
			Description: "отключение рекламы",
			Payload:     "subscription3",
			Currency:    "XTR",
			Total:       3,
			Prices: []tele.Price{
				{Label: "adv free", Amount: 3},
			},
		}
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, "invoice1y", 0)
		return c.Send(invoice)
	})

	b.Handle(&btnStatus, func(c tele.Context) error {
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, "[btnStatus]", 1)

		dt, err := dleRequest("status", c.Sender().ID, "")
		if err != nil {
			log.Println(err)
			database.TelegramLogCreate(dbService, botId, c.Sender().ID, "[btnStatus] ERROR:"+err.Error(), 1)
			return c.Send("Ошибка связи с сервером. Попробуй позже!", menu)
		}

		if dt == "" {
			return c.Send("Статус подписки: не оплачено", menu)
		}

		return c.Send("Статус подписки: оплачено до "+dt, menu)
	})

	b.Handle("/hello", func(c tele.Context) error {
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, "/hello", 1)
		return c.Send("Hello!" + fmt.Sprintf("%d", c.Sender().ID))
	})

	b.Handle("/me", func(c tele.Context) error {
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, "/hello", 1)
		return c.Send("ID:" + fmt.Sprintf("%d", c.Sender().ID))
	})

	// b.Handle("/link", func(c tele.Context) error {
	// 	log.Println("payment request!")
	// 	invoice := tele.Invoice{
	// 		Title:       "Премиум доступ на 1 месяц",
	// 		Description: "без рекламы",
	// 		Payload:     "subscription_1m",
	// 		Currency:    "XTR",
	// 		Total:       1,
	// 		Prices: []tele.Price{
	// 			{Label: "adv free", Amount: 1},
	// 		},
	// 	}
	// 	link, err := b.CreateInvoiceLink(invoice)
	// 	if err != nil {
	// 		log.Println(err)
	// 	}
	// 	return c.Send(link)
	// })

	b.Handle("/refund", func(c tele.Context) error {
		log.Println("payment refund!")
		return b.RefundStars(c.Sender(), "asd")
	})

	b.Handle(tele.OnCheckout, func(c tele.Context) error {
		log.Println("payment started!", c.Message())
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, "payment accepted", 0)
		return c.Accept()
	})

	b.Handle(tele.OnPayment, func(c tele.Context) error {
		log.Println("payment done!")
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, "payment finished", 0)
		subscription := c.Payment().Payload
		dt, err := dleRequest("payment", c.Sender().ID, subscription)
		if err != nil {
			database.TelegramLogCreate(dbService, botId, c.Sender().ID, "payment finish ERROR:"+err.Error(), 1)
			return c.Send("Ошибка связи с сервером. Попробуй позже!", menu)
		}
		return c.Send("Оплата получена.\nПодписка продлена до "+dt+".\nПриятного просмотра!", menu)
	})

	b.Handle(tele.OnText, func(c tele.Context) error {
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, c.Text(), 0)
		return c.Reply("Пардоне муа, не понимаю", menu)
	})

	b.Start()
}

func dleRequest(action string, tgID int64, subscription string) (dt string, err error) {
	q := fmt.Sprintf("https://odminko.printhouse.casa/engine/ajax/controller.php?mod=telegram&?action=%s&tg_id=%d&subscription=%s", action, tgID, subscription)
	//q := fmt.Sprintf("https://proxy.cis-dle.orb.local/engine/ajax/controller.php?mod=telegram&action=%s&tgid=%d&subscription=%s", action, tgID, subscription)
	req, err := http.NewRequest("GET", q, nil)

	type dleResponse struct {
		Status  string `json:"status"`
		Premium string `json:"premium"`
	}

	if err != nil {
		log.Println(err)
		return
	}

	// req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Println(err)
		return
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	log.Println(string(body))

	r := dleResponse{}
	if err = json.Unmarshal(body, &r); err != nil {
		log.Println("error unmarshal:", err)
		return
	}

	if err != nil {
		log.Println(err)
		return
	}

	dt = r.Premium
	return
}
