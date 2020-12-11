package bot

import (
	"CTBTgBot/internal/conf"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type B = tgbotapi.BotAPI
type M = tgbotapi.Message
type Context struct {
	admin   int
	channel int64
}
type C = Context

func setup() *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(conf.ReadToken())
	if err != nil {
		log.Fatal(err)
	}
	return bot
}

func Run() {
	bot := setup()
	bot.Debug = true
	log.Println("Successfully establish connection to bot", bot.Self.UserName)
	//TODO: Load owner

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}
	for update := range updates {
		if update.Message == nil {
			continue
		}
		log.Printf("[%s][%s]", update.Message.From, update.Message.Text)
	}
}

func cmdHandler(bot *B, msg *M) {
	if msg.IsCommand() {

	}
}

func sendT(b *B, chatID int64, text string) (M, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	return b.Send(msg)
}

func cmdStart(b *B, m *M, c *C) {
	_, err := sendT(b, m.Chat.ID, "A bot for helping you manage contribution.\nAuthor: @SaiToAsuka_kksk")
	if err != nil {
		log.Println("[cmdStart]", err)
	}
}

func cmdHelp(b *B, m *M, c *C) {

}

func cmdCTB(b *B, m *M, c *C) {
	if m.ReplyToMessage != nil {
		if len(*m.ReplyToMessage.Photo) != 0 {
			pht := tgbotapi.NewPhotoShare(int64(c.admin), (*m.ReplyToMessage.Photo)[0].FileID)
			b.Send(pht)
			sendT(b, m.Chat.ID, "Successfully forwarding your message, thanks for your contribution.")
		}
	}
}
