package telegram

import (
	"fmt"
	"log"
	"strconv"

	"github.com/beeker1121/goque"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"golang.org/x/time/rate"

	"github.com/chaostreff-flensburg/moc-go/models"
	"github.com/chaostreff-flensburg/moc-telegram/config"
)

type Telegram struct {
	Token   string
	Text    config.Text
	Bot     *tgbotapi.BotAPI
	DB      *leveldb.DB
	Queue   *goque.Queue
	Limiter rate.Limiter
}

func NewTelegram(config *config.Config, db *leveldb.DB) *Telegram {
	return &Telegram{
		Token:   config.TelegramToken,
		DB:      db,
		Text:    config.Text,
		Limiter: rate.NewLimiter(20, 30),
	}
}

func (t *Telegram) Connect() {
	q, err := goque.OpenQueue("/data/queue.db")
	if err != nil {
		log.Panic(err)
	}
	t.Queue = q

	bot, err := tgbotapi.NewBotAPI(t.Token)
	if err != nil {
		log.Panic(err)
	}

	t.Bot = bot

	t.Bot.Debug = false

	log.Printf("Authorized on account %s", t.Bot.Self.UserName)
}

func (t *Telegram) Loop() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	var subscribeKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(t.Text.Subscribe),
		),
	)

	var unsubscribeKeyboard = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(t.Text.Unsubscribe),
		),
	)

	updates, _ := t.Bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore non-Message updates
			continue
		}

		chatID := update.Message.Chat.ID
		msg := tgbotapi.NewMessage(chatID, t.Text.Hello)

		switch update.Message.Text {
		case t.Text.Subscribe:
			msg.ReplyMarkup = unsubscribeKeyboard
			msg.Text = t.Text.Subscribed

			key := fmt.Sprintf("subscriber.%v", chatID)
			t.DB.Put([]byte(key), []byte(fmt.Sprintf("%v", chatID)), nil)
		case t.Text.Unsubscribe:
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
			msg.Text = t.Text.Unsubscribed

			t.DB.Delete([]byte(fmt.Sprintf("subscriber.%v", chatID)), nil)
		default:
			msg.ReplyMarkup = subscribeKeyboard
		}

		if _, err := q.EnqueueObject(msg); err != nil {
			log.Panic(err)
		}
	}
}

func (t *Telegram) SendMessage(chatID int64, message *models.Message) {
	msg := tgbotapi.NewMessage(chatID, message.Text)

	if _, err := q.EnqueueObject(msg); err != nil {
		log.Panic(err)
	}
}

func (t *Telegram) SendAll(message *models.Message) {
	log.Println("Start sending to all...")

	iter := t.DB.NewIterator(util.BytesPrefix([]byte("subscriber.")), nil)
	for iter.Next() {
		value := iter.Value()
		i, _ := strconv.ParseInt(string(value), 10, 64)
		t.SendMessage(i, message)
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		log.Println(err)
	}

	log.Println("Finish sending to all.")
}

func (t *Telegram) SendLoop(tick time.Duration) {
	go func() {
		for range time.Tick(tick) {
			t.SendFromQueue()
		}
	}()
}

func (t *Telegram) SendFromQueue() {
	_, err := q.Peek()
	if err != nil {
		return
	}

	if t.Limiter.Allow() == false {
		return
	}
	item, _ := q.Dequeue()

	var msg tgbotapi.MessageConfig
	err = item.ToObject(&msg)
	if err != nil {
		log.Println(err)
	}

	if _, err := t.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}
