package bot

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) SendStart(chatID int64) (err error) {
	b.SendBottomButtons(chatID, "default")
	t := "Hi! Send me a movie name or choose one of the most pospular movies\n\n⬇️⬇️⬇️\n"
	settings := b.settingsService.Get(b.ID, "start", "text")
	if len(settings) > 0 {
		t = settings[0].Content
	}
	msg := tgbotapi.NewMessage(chatID, t)
	var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("OPEN APP", "https://t.me/netflix_brazilbot/brazilbotapp?startapp=default"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("House of dragons", "https://t.me/netflix_brazilbot/brazilbotapp?startapp=a-casa-do-dragao-noads"),
		),
	)
	settingsButtons := b.settingsService.Get(b.ID, "start", "button")
	if len(settingsButtons) > 0 {
		var rows [][]tgbotapi.InlineKeyboardButton
		for _, sb := range settingsButtons {
			link := sb.Link
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
	return b.SendMsg(msg)
}
