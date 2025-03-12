package bot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) Send(ch int64, message string) (err error) {
	msgConfig := tgbotapi.NewMessage(ch, message)
	// msg.ParseMode = tgbotapi.ModeHTML
	// msg.DisableWebPagePreview = true
	err = b.SendMsg(msgConfig)
	return
}

func (b *Bot) SendInvoice(ch int64, message string) (err error) {
	prices := []tgbotapi.LabeledPrice{
		{Label: "adv free", Amount: 100},
	}
	msgConfig := tgbotapi.NewInvoice(ch, "pay 1 m", "to get adv free", "pay", "pay", "XTR", "10", prices)
	// msg.ParseMode = tgbotapi.ModeHTML
	// msg.DisableWebPagePreview = true
	err = b.SendMsg(msgConfig)
	return
}

func (b *Bot) SendMsg(msg tgbotapi.Chattable) (err error) {
	if b.api == nil {
		log.Println("error sendMsg: api is nil")
		return
	}
	if _, err = b.api.Send(msg); err != nil {
		log.Printf("(bot %d) send message error: %s\n", b.ID, err)
	}
	return
}

func (b *Bot) SendBottomButtons(chatID int64, setName string) {
	sendMeBotLinkCommand := "OPEN APP"
	bottomButton := b.settingsService.Get(b.ID, "bottom", "button")
	bottomMessage := b.settingsService.Get(b.ID, "bottom", "message")
	bottomTxt := "🎬"
	if len(bottomMessage) > 0 {
		bottomTxt = bottomMessage[0].Content
	}
	msg := tgbotapi.NewMessage(chatID, bottomTxt)
	var numericKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(sendMeBotLinkCommand),
		),
	)
	if len(bottomButton) > 0 {
		numericKeyboard = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(bottomButton[0].Content),
			),
		)
	}
	msg.ReplyMarkup = numericKeyboard
	b.SendMsg(msg)
}

func (b *Bot) listen() {

	// Create a new UpdateConfig struct with an offset of 0. Offsets are used
	// to make sure Telegram knows we've handled previous values and we don't
	// need them repeated.
	updateConfig := tgbotapi.NewUpdate(0)

	// Tell Telegram we should wait up to 30 seconds on each request for an
	// update. This way we can get information just as quickly as making many
	// frequent requests without having to send nearly as many.
	updateConfig.Timeout = 30

	// Start polling Telegram for updates.
	updates := b.api.GetUpdatesChan(updateConfig)

	// Let's go through each update that we're getting from Telegram.
	go func(quit chan bool, botId int) {
		for {
			select {
			case <-quit:
				log.Printf("kill listener for bot %d", botId)
				return
			case update := <-updates:
				b.update(update)
			}
		}
	}(b.quit, b.ID)

}
