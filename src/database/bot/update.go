package bot

import (
	"cis-telegram/database"
	"cis-telegram/database/settings"
	"cis-telegram/helper"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Update struct {
	tgUpdate             tgbotapi.Update
	telegramUser         *database.TelegramUser
	bot                  *Bot
	userID               int64
	userNameFull         string
	chatID               int64
	chatType             string
	text                 string
	dbService            *database.Service
	settingsService      *settings.Service
	startCommand         string
	sendMeBotLinkCommand string
}

func (u *Update) Prepare(bot *Bot, dbService *database.Service, settingsService *settings.Service, tgUpdate tgbotapi.Update) {
	u.bot = bot
	u.tgUpdate = tgUpdate
	u.dbService = dbService
	u.settingsService = settingsService
	u.chatID = u.tgUpdate.Message.Chat.ID
	u.chatType = u.tgUpdate.Message.Chat.Type // private
	u.text = u.tgUpdate.Message.Text

	//helper.P(u.tgUpdate)
	//helper.P(u.tgUpdate.Message.Text)
	//helper.P(u.tgUpdate.Message.ReplyToMessage)

	log.Println("update age:", time.Since(time.Unix(int64(u.tgUpdate.Message.Date), 0)))

	// ignore ALL replies
	if u.tgUpdate.Message.ReplyToMessage != nil {
		u.text = ""
		log.Println("ignore reply message")
		return
	}

	// ignore long requests
	if len(u.text) > 200 {
		u.text = ""
		log.Println("ignore too long message")
		return
	}

	user := u.tgUpdate.Message.From
	u.userID = user.ID
	u.userNameFull = fmt.Sprintf("%s (%s %s)", user.UserName, user.FirstName, user.LastName)
	d := strings.Split(u.text, "@")
	if len(d) > 1 {
		u.text = d[0]
	}

	telegramUser := database.NewTelegramUser(dbService, u.bot.ID, u.userID)
	if u.chatType == "private" { // save only user who has a bot
		telegramUser.Upsert(u.text)
	}
	u.telegramUser = telegramUser
	//log.Println("telegramUserID:", telegramUser.ID)

	// /start__source_medium_campaign
	if strings.HasPrefix(u.text, "/start") {
		parts := strings.Split(u.text, " ")
		if len(parts) > 0 {
			u.text = parts[0]
		}
		if len(parts) > 1 {
			u.startCommand = parts[1]
		}
	}

	u.sendMeBotLinkCommand = "OPEN APP"
	bottomButton := u.settingsService.Get(u.bot.ID, "bottom", "button")
	if len(bottomButton) > 0 {
		u.sendMeBotLinkCommand = bottomButton[0].Content
	}

}

func (u *Update) CmdStart() {
	u.bot.SendStart(u.chatID)
	// ga.SendEvent(fmt.Sprintf("%d", userID), "start", startCommand)
}

func (u *Update) CmdDebug() {
	u.bot.Send(u.chatID, fmt.Sprintf("This chatID: %d", u.chatID))
	u.bot.Send(u.chatID, fmt.Sprintf("Your ID: %d", u.userID))
	u.bot.Send(u.chatID, "Your Username: "+u.userNameFull)
}

func (u *Update) CmdPing() {
	u.bot.Send(u.chatID, "pong")
	// s.gaService.Track(fmt.Sprintf("%d", userID), "ping", "")
}

func (u *Update) CmdPers() {
	u.bot.Send(u.userID, "hello from bot!")
}

func (u *Update) CmdInvoice(size int) {
	u.bot.SendInvoice(u.userID, "1 month subscription")
}

func (u *Update) CmdSearch() {
	// search in app
	var posts []*Post
	limit := 5
	foundNumber := u.settingsService.Get(u.bot.ID, "search", "found_number")
	if len(foundNumber) > 0 {
		limit = helper.StrToInt(foundNumber[0].Content)
	}

	// hack to avoid our text
	if strings.Contains(u.text, "Open Bot App") {
		return
	}

	params := url.Values{}
	params.Add("limit", fmt.Sprintf("%d", limit))
	params.Add("q", u.text)

	log.Println(">>>", u.text)

	body, err := helper.GetURL(fmt.Sprintf("https://api.mymoviesonline.in/api/1.0/search?%s", params.Encode()))
	if err == nil {
		err = json.Unmarshal(body, &posts)
		if err != nil {
			log.Printf("api fetched but cant be unmarshalled: %s", err)
		}
	}
	_ = database.NewTelegramSearch(u.dbService, u.bot.ID, u.userID, u.text, len(posts))
	// not found
	if len(posts) == 0 {
		txt := "Not found"
		notFoundText := u.settingsService.Get(u.bot.ID, "search", "notfound_text")
		if len(notFoundText) > 0 {
			txt = notFoundText[0].Content
		}
		msg := tgbotapi.NewMessage(u.chatID, txt)
		var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("OPEN APP", u.bot.AppURL),
			),
		)
		notFoundButtons := u.settingsService.Get(u.bot.ID, "search", "notfound_button")
		if len(notFoundButtons) > 0 {
			var rows [][]tgbotapi.InlineKeyboardButton
			for _, sb := range notFoundButtons {
				link := sb.Link
				if u.telegramUser.ID == 0 { // override with bot link for not registered users
					link = u.bot.BotURL
				}
				if !strings.HasPrefix(link, "https://") {
					link = "https://" + sb.Link
				}
				b := tgbotapi.NewInlineKeyboardButtonURL(sb.Content, link)
				row := []tgbotapi.InlineKeyboardButton{b}
				rows = append(rows, row)
			}
			numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(rows...)
		}
		msg.ReplyMarkup = numericKeyboard
		u.bot.SendMsg(msg)
		// s.Send(echoChatID, fmt.Sprintf(`[NOT FOUND]: "%s"`, text))
		// s.gaService.Track(fmt.Sprintf("%d", userID), "search_miss", text)
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, post := range posts {
		link := u.bot.AppURL + "?startapp=" + post.Slug
		if u.telegramUser == nil || u.telegramUser.ID == 0 { // override with bot link for not registered users
			link = u.bot.BotURL
		}

		b := tgbotapi.NewInlineKeyboardButtonURL(post.Title, link)
		row := []tgbotapi.InlineKeyboardButton{b}
		rows = append(rows, row)
	}
	txt := "Found"
	foundText := u.settingsService.Get(u.bot.ID, "search", "found_text")
	if len(foundText) > 0 {
		txt = foundText[0].Content
	}
	msg := tgbotapi.NewMessage(u.chatID, txt)
	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg.ReplyMarkup = numericKeyboard
	u.bot.SendMsg(msg)
	// s.Send(echoChatID, fmt.Sprintf(`[OK]: "%s"`, text))
	// s.gaService.Track(fmt.Sprintf("%d", userID), "search_ok", text)
}

func (u *Update) CmdTest() {
	msg := tgbotapi.NewPhoto(u.chatID, tgbotapi.FileURL("https://cinemasound.ua/test/fav/td.jpg"))
	var numericKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🎬 Filmes e Series 🎬"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/test"),
		),
	)
	msg.ReplyMarkup = numericKeyboard
	msg.Caption = "Test notification about new episode - added new serie"
	u.bot.SendMsg(msg)
}

func (u *Update) CmdSendMeBotLink() {
	msg := tgbotapi.NewMessage(u.chatID, "⬇️⬇️⬇️")
	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("OPEN APP", u.bot.AppURL+"?startapp=default"),
		),
	)
	// Bad Request: can't parse inline keyboard button: Text buttons are unallowed in the inline keyboard
	bottomButton := u.settingsService.Get(u.bot.ID, "bottom", "button")
	if len(bottomButton) > 0 {
		link := bottomButton[0].Link
		if !strings.HasPrefix(link, "https://") {
			link = "https://" + link
		}
		numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL(bottomButton[0].Content, link),
			),
		)
	}
	msg.ReplyMarkup = numericKeyboard
	u.bot.SendMsg(msg)
}

func (b *Bot) update(tgUpdate tgbotapi.Update) {
	// log.Printf("TgUpdate: %+v\n", tgUpdate.Message)
	// Check if we've gotten a message tgUpdate.
	if tgUpdate.Message != nil {
		update := &Update{}
		update.Prepare(b, b.dbService, b.settingsService, tgUpdate)
		//log.Println("update short:", update.text)
		switch update.text {
		case "/start":
			update.CmdStart()
		case "/debug":
			update.CmdDebug()
		case update.sendMeBotLinkCommand:
			update.CmdSendMeBotLink()
		case "/test":
			update.CmdTest()
		case "/ping":
			update.CmdPing()
		case "/pers":
			update.CmdPers()
		case "/invoice_1m":
			update.CmdInvoice(1)
		case "":
			// do nothing
		default:
			update.CmdSearch()
		}
	} else if tgUpdate.CallbackQuery != nil {
		callback := tgbotapi.NewCallback(tgUpdate.CallbackQuery.ID, tgUpdate.CallbackQuery.Data)
		if _, err := b.api.Request(callback); err != nil {
			log.Println(err)
		}
		switch tgUpdate.CallbackQuery.Data {
		case "news":
			b.Send(tgUpdate.CallbackQuery.Message.Chat.ID, "news!")
		default:
			b.Send(tgUpdate.CallbackQuery.Message.Chat.ID, "??? "+tgUpdate.CallbackQuery.Data)
		}
	}

}
