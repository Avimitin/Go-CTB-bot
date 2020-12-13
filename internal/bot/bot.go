package bot

import (
	"CTBTgBot/internal/conf"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// B is a short name of BotAPI
type B = tgbotapi.BotAPI

// M is a short name of Message
type M = tgbotapi.Message

// Context is a struct that contains necessary config
type Context struct {
	admin    int
	channel  int64
	register map[int]*Submit
	callBack chan string
	commands *map[string]func(*B, *M, *C)
}

func NewCTX(admin int, channel int64) *Context {
	return &Context{
		admin:    admin,
		channel:  channel,
		register: make(map[int]*Submit),
		callBack: make(chan string),
	}
}

// C is a short name of Context
type C = Context

type Submit struct {
	photo       string
	photoSource string
	submitter   int
	done        bool
}

func NewSubmit(submitter int) *Submit {
	return &Submit{
		submitter: submitter,
	}
}

func (s *Submit) regisPhoto(photoID string) {
	s.photo = photoID
}

func (s *Submit) regisPhtSrc(sourceURL string) {
	s.photoSource = sourceURL
}

func setup() *tgbotapi.BotAPI {
	bot, err := tgbotapi.NewBotAPI(conf.ReadToken())
	if err != nil {
		log.Fatal(err)
	}
	return bot
}

func sendT(b *B, chatID int64, text string) (M, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	return b.Send(msg)
}

// Run function run the bot
func Run() {
	bot := setup()
	bot.Debug = true
	log.Println("[INFO]Successfully establish connection to bot", bot.Self.UserName)
	admin, channel := conf.ReadUsrInfo()
	log.Printf("[INFO]User information initalized:\n admin: %v, channel: %v", admin, channel)
	ctx := NewCTX(admin, channel)
	ctx.commands = initCommands()
	var cmds string
	for key := range *ctx.commands {
		cmds += key + " "
	}
	log.Printf("[INFO]Command initalized: %+v", cmds)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}
	go callBackHandler(bot, ctx)
	for update := range updates {
		if update.CallbackQuery != nil {
			log.Println("[INFO]Receive new data", update.CallbackQuery.Data)
			ctx.callBack <- update.CallbackQuery.Data
		}
		if update.Message == nil {
			continue
		}
		// If a submit started and not done yet
		if sub, ok := ctx.register[update.Message.From.ID]; ok && !sub.done {
			// If user cancel it
			if update.Message.IsCommand() && update.Message.Command() == "cancel" {
				delete(ctx.register, update.Message.From.ID)
				log.Println("[INFO]A contribution process end: User cancel.")
				sendT(bot, update.Message.Chat.ID, "Successfully cancel process.")
				continue
			}

			// submit photo
			if sub.photo == "" {
				if update.Message.Photo != nil {
					sub.regisPhoto((*update.Message.Photo)[0].FileID)
					sendT(bot, update.Message.Chat.ID,
						"Next can you give me the source URL of this photo?, If you don't have it you can give me"+
							"an abstract source or just send n/N to me.")
					continue
				} else {
					sendT(bot, update.Message.Chat.ID, "Invalid message, cancel process")
					log.Println("[INFO]A contribution process end: User invalid input.")
					delete(ctx.register, update.Message.From.ID)
					continue
				}
			}

			// if submit descriptions is null
			if sub.photoSource == "" {
				// if user give text message
				if update.Message.Text != "" {
					// if user's message is not "N" or "n"
					if update.Message.Text != "N" && update.Message.Text != "n" {
						sub.regisPhtSrc(update.Message.Text)
					}

					sendT(bot, update.Message.Chat.ID, "Thanks for your submission.")
					sub.done = true
					// Admin
					pht := tgbotapi.NewPhotoShare(int64(ctx.admin), sub.photo)
					bot.Send(pht)
					msg := tgbotapi.NewMessage(int64(ctx.admin), "You have a new submission.")
					ikbm := makeButton(sub.submitter)
					msg.ReplyMarkup = ikbm
					bot.Send(msg)

				} else {
					sendT(bot, update.Message.Chat.ID, "Invalid input, process stop.")
					log.Println("[INFO]A contribution process end: User invalid input.")
					delete(ctx.register, update.Message.From.ID)
					continue
				}
			}
		}

		log.Printf("[%s][%s]", update.Message.From, update.Message.Text)
		go cmdHandler(bot, update.Message, ctx)
	}
}

func initCommands() *map[string]func(*B, *M, *C) {
	return &map[string]func(*tgbotapi.BotAPI, *tgbotapi.Message, *Context){
		"start": cmdStart,
		"help":  cmdHelp,
		"ctb":   cmdCTB,
	}
}

// make two button pass and refuse to investigate new contribution
func makeButton(submiiter int) tgbotapi.InlineKeyboardMarkup {
	submiiterStr := strconv.Itoa(submiiter)
	pass := "P" + submiiterStr
	refuse := "R" + submiiterStr
	passBtn := tgbotapi.InlineKeyboardButton{
		Text:         "Pass",
		CallbackData: &pass,
	}
	rfsBtn := tgbotapi.InlineKeyboardButton{
		Text:         "Refuse",
		CallbackData: &refuse,
	}
	return tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{passBtn, rfsBtn})
}

func cmdHandler(bot *B, msg *M, ctx *C) {
	if msg.IsCommand() {
		if cmd, ok := (*ctx.commands)[msg.Command()]; ok {
			cmd(bot, msg, ctx)
		}
	}
}

func callBackHandler(b *B, c *C) {
	for cb := range c.callBack {
		switch cb[0] {
		// If pass ctb
		case 'P':
			cb = cb[1:]
			submitter, err := strconv.Atoi(cb)
			log.Println("[INFO]Make new send photo request about", submitter)
			if err != nil {
				//TODO: report error
				log.Println(err)
				continue
			}
			if sub, ok := c.register[submitter]; ok {
				pht := tgbotapi.NewPhotoShare(c.channel, sub.photo)
				pht.Caption = sub.photoSource
				b.Send(pht)
				sendT(b, int64(c.admin), "Successfully post new contribution.")
				delete(c.register, submitter)
			} else {
				sendT(b, int64(c.admin), "Message has expired.")
			}
		case 'R':
			cb = cb[1:]
			submitter, err := strconv.Atoi(cb)
			if err != nil {
				//TODO: report error
				log.Println(err)
				continue
			}
			if _, ok := c.register[submitter]; ok {
				delete(c.register, submitter)
				sendT(b, int64(c.admin), "This contribution has been refused.")
			} else {
				sendT(b, int64(c.admin), "Message has expired.")
			}
		}
	}
}

func cmdStart(b *B, m *M, c *C) {
	_, err := sendT(b, m.Chat.ID, "A bot for helping you manage contribution.\nAuthor: @SaiToAsuka_kksk")
	if err != nil {
		log.Println("[cmdStart]", err)
	}
}

func cmdHelp(b *B, m *M, c *C) {
	sendT(b, m.Chat.ID, "Use command /ctb to submit a new contribution.")
}

func cmdCTB(b *B, m *M, c *C) {
	if c.register == nil {
		c.register = make(map[int]*Submit)
	}
	c.register[m.From.ID] = NewSubmit(m.From.ID)
	sendT(b, m.Chat.ID,
		"A new submit process has started, you can use /cancel to cancel this process in any time.\n\n"+
			"Now please give me a photo")
}
