package loginbot

import (
	"bytes"
	"cis-telegram/database"
	"encoding/json"
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
	// btnHelp := menu.Text("Помощь")
	// btnSubscribe1 := menu.Text("1 мес = 1⭐")
	// btnSubscribe2 := menu.Text("6 мес = 2⭐")
	// btnSubscribe3 := menu.Text("1 год = 3⭐")
	// btnStatus := menu.Text("Статус подписки")
	// btnPrev := selector.Data("⬅", "prev")
	// btnNext := selector.Data("➡", "next")

	btn1 := menu.Text("Трюм - канал новинок")
	btn2 := menu.Text("База в Телеграм")
	btn3 := menu.Text("Тортуга - сайт")
	btn4 := menu.Text("Contact us")

	menu.Reply(
		// menu.Row(btnSubscribe1, btnSubscribe2, btnSubscribe3),
		// menu.Row(btnStatus),
		// menu.Row(btnHelp),
		menu.Row(btn1),
		menu.Row(btn2),
		menu.Row(btn3),
		menu.Row(btn4),
	)

	inline := &tele.ReplyMarkup{}
	btn1inline := inline.URL("Источник вечной молодости", "https://piratka.me/movies/128180-istochnik-vechnoj-molodosti.html")
	btn2inline := inline.URL("Микки 17", "https://piratka.me/movies/126760-mikki-17.html")
	btn3inline := inline.URL("Любовь, смерть и роботы", "https://piratka.me/series/106955-ljubov-smert-i-roboty.html")
	btn4inline := inline.URL("Настоящие детективы", "https://piratka.me/movies/127901-nastojaschie-detektivy.html")

	inline.Inline(
		inline.Row(btn1inline),
		inline.Row(btn2inline),
		inline.Row(btn3inline),
		inline.Row(btn4inline),
	)
	// selector.Inline(
	// 	selector.Row(btnPrev, btnNext),
	// )

	b.Handle("/start", func(c tele.Context) (err error) {
		c.Send("<> Добро пожаловать на борт!\n\n"+
			"Для тебя любое кино и сериалы без ограничений!\n"+
			"Смотри в любом формате и месте - Пиратка добудет все!\n\n"+
			"В этом боте я сообщу тебе о всех новинках сериалов и анонсов, которые ты поместил в Избранное.\n\n"+
			"На моем корабле ты найдешь:\n"+
			"Трюм - канал в котором вся информация о моей самой свежей добыче - все последние новинки\n"+
			"База в Телеграм - не выходя из Телеграма проводи время с удовольствием - смотри кино с удовольствием\n"+
			"Пиратка APК - приложение на твоем телефоне - твои любимые фильмы всегда под рукой\n"+
			"Тортуга - наш сайт - главная база, где тусуются все настоящие Пираты\n\n"+
			"А сейчас вот твоя доля в добыче - 5 самых горячих новинок в отличном качестве:", menu)
		err = c.Send("А сейчас вот твоя доля в добыче - 5 самых горячих новинок в отличном качестве:", inline)
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, "/start", 1)
		return err
	})

	b.Handle(&btn1, func(c tele.Context) error {
		return c.Send("Трюм @new_movie_hd_4k_h_bot", menu)
	})

	b.Handle(&btn2, func(c tele.Context) error {
		return c.Send("Смотри лучшие фильмы и сериалы в telegram приложении https://t.me/piratka_me_app_bot/app?startapp=default", menu)
	})

	b.Handle(&btn3, func(c tele.Context) error {
		return c.Send("Смотри лучшие фильмы и сериалы на сайте https://piratka.me", menu)
		// return c.Send("Скачай приложение для android https://apk.piratka.me/engine/ajax/controller.php?mod=download_apk", menu)
	})

	b.Handle(&btn4, func(c tele.Context) error {
		return c.Send("Напиши сообщение больше 10 символов и мы его получим:", menu)
	})

	// On reply button pressed (message)
	// b.Handle(&btnHelp, func(c tele.Context) error {
	// 	database.TelegramLogCreate(dbService, botId, c.Sender().ID, "[btnHelp]", 1)
	// 	return c.Send("Этот бот позволяет оплатить премиум доступ к сервису.\nОплата производится в звездах.\nПотрачанные звезды не возвращаются.\nПодарки - не отдарки! :)", menu)
	// })

	// On inline button pressed (callback)
	// b.Handle("/test10m", func(c tele.Context) error {
	// 	database.TelegramLogCreate(dbService, botId, c.Sender().ID, "[test10min]", 1)
	// 	log.Println("payment request!")
	// 	invoice := &tele.Invoice{
	// 		Title:       "Премиум доступ на 10 минут",
	// 		Description: "отключение рекламы",
	// 		Payload:     "subscription_test_1",
	// 		Currency:    "XTR",
	// 		Total:       1,
	// 		Prices: []tele.Price{
	// 			{Label: "adv free", Amount: 1},
	// 		},
	// 	}
	// 	database.TelegramLogCreate(dbService, botId, c.Sender().ID, "invoice_test_1", 0)
	// 	return c.Send(invoice)
	// })
	// b.Handle("/test24h", func(c tele.Context) error {
	// 	database.TelegramLogCreate(dbService, botId, c.Sender().ID, "[test24h]", 1)
	// 	log.Println("payment request!")
	// 	invoice := &tele.Invoice{
	// 		Title:       "Премиум доступ на 24 часа",
	// 		Description: "отключение рекламы",
	// 		Payload:     "subscription_test_2",
	// 		Currency:    "XTR",
	// 		Total:       1,
	// 		Prices: []tele.Price{
	// 			{Label: "adv free", Amount: 1},
	// 		},
	// 	}
	// 	database.TelegramLogCreate(dbService, botId, c.Sender().ID, "invoice_test_2", 0)
	// 	return c.Send(invoice)
	// })

	// b.Handle(&btnSubscribe1, func(c tele.Context) error {
	// 	database.TelegramLogCreate(dbService, botId, c.Sender().ID, "[btnSubscribe1]", 1)
	// 	log.Println("payment request!")
	// 	invoice := &tele.Invoice{
	// 		Title:       "Премиум доступ на 1 месяц",
	// 		Description: "отключение рекламы",
	// 		Payload:     "subscription1",
	// 		Currency:    "XTR",
	// 		Total:       1,
	// 		Prices: []tele.Price{
	// 			{Label: "adv free", Amount: 1},
	// 		},
	// 	}
	// 	database.TelegramLogCreate(dbService, botId, c.Sender().ID, "invoice1m", 0)
	// 	return c.Send(invoice)
	// })

	// b.Handle(&btnSubscribe2, func(c tele.Context) error {
	// 	database.TelegramLogCreate(dbService, botId, c.Sender().ID, "[btnSubscribe2]", 1)
	// 	log.Println("payment request!")
	// 	invoice := &tele.Invoice{
	// 		Title:       "Премиум доступ на 6 месяцев",
	// 		Description: "отключение рекламы",
	// 		Payload:     "subscription2",
	// 		Currency:    "XTR",
	// 		Total:       2,
	// 		Prices: []tele.Price{
	// 			{Label: "adv free", Amount: 2},
	// 		},
	// 	}
	// 	database.TelegramLogCreate(dbService, botId, c.Sender().ID, "invoice6m", 0)
	// 	return c.Send(invoice)
	// })

	// b.Handle(&btnSubscribe3, func(c tele.Context) error {
	// 	log.Println("payment request!")
	// 	database.TelegramLogCreate(dbService, botId, c.Sender().ID, "[btnSubscribe3]", 1)
	// 	invoice := &tele.Invoice{
	// 		Title:       "Премиум доступ на 1 год",
	// 		Description: "отключение рекламы",
	// 		Payload:     "subscription3",
	// 		Currency:    "XTR",
	// 		Total:       3,
	// 		Prices: []tele.Price{
	// 			{Label: "adv free", Amount: 3},
	// 		},
	// 	}
	// 	database.TelegramLogCreate(dbService, botId, c.Sender().ID, "invoice1y", 0)
	// 	return c.Send(invoice)
	// })

	// b.Handle(&btnStatus, func(c tele.Context) error {
	// 	database.TelegramLogCreate(dbService, botId, c.Sender().ID, "[btnStatus]", 1)

	// 	dt, err := dleRequest("status", c.Sender().ID, "")
	// 	if err != nil {
	// 		log.Println(err)
	// 		database.TelegramLogCreate(dbService, botId, c.Sender().ID, "[btnStatus] ERROR:"+err.Error(), 1)
	// 		return c.Send("Ошибка связи с сервером. Попробуй позже!", menu)
	// 	}

	// 	if dt == "" {
	// 		return c.Send("Статус подписки: не оплачено", menu)
	// 	}

	// 	return c.Send("Статус подписки: оплачено до "+dt, menu)
	// })

	b.Handle("/hello", func(c tele.Context) error {
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, "/hello", 1)
		return c.Send("Hello!" + fmt.Sprintf("%d", c.Sender().ID))
	})

	b.Handle("/me", func(c tele.Context) error {
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, "/hello", 1)
		return c.Send("ID:" + fmt.Sprintf("%d", c.Sender().ID))
	})

	// b.Handle("/remove_premium", func(c tele.Context) error {
	// 	database.TelegramLogCreate(dbService, botId, c.Sender().ID, "/remove_premium", 1)
	// 	dleRequest("remove_premium", c.Sender().ID, "")
	// 	return c.Send("все, у тебя нет премиума!")
	// })

	// b.Handle("/refund", func(c tele.Context) error {
	// 	log.Println("payment refund!")
	// 	return b.RefundStars(c.Sender(), "asd")
	// })

	// b.Handle(tele.OnCheckout, func(c tele.Context) error {
	// 	log.Println("payment started!", c.Message())
	// 	database.TelegramLogCreate(dbService, botId, c.Sender().ID, "payment accepted", 0)
	// 	return c.Accept()
	// })

	// b.Handle(tele.OnPayment, func(c tele.Context) error {
	// 	log.Println("payment done!")
	// 	database.TelegramLogCreate(dbService, botId, c.Sender().ID, "payment finished", 0)
	// 	subscription := c.Payment().Payload
	// 	dt, err := dleRequest("payment", c.Sender().ID, subscription)
	// 	if err != nil {
	// 		database.TelegramLogCreate(dbService, botId, c.Sender().ID, "payment finish ERROR:"+err.Error(), 1)
	// 		return c.Send("Ошибка связи с сервером. Попробуй позже!", menu)
	// 	}
	// 	return c.Send("Оплата получена.\nПодписка продлена до "+dt+".\nПриятного просмотра!", menu)
	// })

	b.Handle(tele.OnText, func(c tele.Context) error {
		database.TelegramLogCreate(dbService, botId, c.Sender().ID, c.Text(), 0)
		text := c.Text()
		if len(text) > 10 {
			dleComplain(text, c.Sender().ID)
			return c.Reply("Сообщение отправлено, мы его обязательно прочитаем!", inline)
		}
		return c.Reply("Пардоне муа, не понимаю", menu)
	})

	go b.Start()
	return
}

func dleRequest(action string, tgID int64, subscription string) (dt string, err error) {
	q := fmt.Sprintf("https://odminko.printhouse.casa/engine/ajax/controller.php?mod=telegram&action=%s&tgid=%d&subscription=%s", action, tgID, subscription)
	//q := fmt.Sprintf("https://proxy.cis-dle.orb.local/engine/ajax/controller.php?mod=telegram&action=%s&tgid=%d&subscription=%s", action, tgID, subscription)
	log.Println(q)
	req, err := http.NewRequest("POST", q, nil)

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
