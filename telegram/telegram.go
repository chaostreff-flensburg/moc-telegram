package telegram

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/beeker1121/goque"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"golang.org/x/time/rate"

	"github.com/chaostreff-flensburg/moc-go/models"
	"github.com/chaostreff-flensburg/moc-telegram/config"
	tmodels "github.com/chaostreff-flensburg/moc-telegram/models"
)

type Telegram struct {
	Token   string
	Text    config.Text
	Bot     *tgbotapi.BotAPI
	DB      *leveldb.DB
	Queue   *goque.Queue
	Limiter *rate.Limiter
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

	updates, _ := t.Bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore non-Message updates
			continue
		}

		chatID := update.Message.Chat.ID
		text := t.Text.Hello
		keyboard := tmodels.KeyboardTypeNone

		switch update.Message.Text {
		case t.Text.Subscribe:
			keyboard = tmodels.KeyboardTypeUnsubscribe
			text = t.Text.Subscribed

			key := fmt.Sprintf("subscriber.%v", chatID)
			t.DB.Put([]byte(key), []byte(fmt.Sprintf("%v", chatID)), nil)
		case t.Text.Unsubscribe:
			keyboard = tmodels.KeyboardTypeRemove
			text = t.Text.Unsubscribed

			t.DB.Delete([]byte(fmt.Sprintf("subscriber.%v", chatID)), nil)
		default:
			keyboard = tmodels.KeyboardTypeSubscribe
		}

		msg := tmodels.NewQueueEntry(chatID, text, keyboard)

		if _, err := t.Queue.EnqueueObject(msg); err != nil {
			log.Panic(err)
		}
	}
}

func (t *Telegram) SendMessage(chatID int64, message *models.Message) {
	//	msg := tgbotapi.NewMessage(chatID, message.Text)
	msg := tmodels.NewQueueEntry(chatID, message.Text, tmodels.KeyboardTypeNone)

	if _, err := t.Queue.EnqueueObject(msg); err != nil {
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
	_, err := t.Queue.Peek()
	if err != nil {
		return
	}

	if t.Limiter.Allow() == false {
		return
	}
	item, _ := t.Queue.Dequeue()

	var queueEntry tmodels.QueueEntry
	err = item.ToObject(&queueEntry)
	if err != nil {
		log.Println(err)
	}

	msg := tgbotapi.NewMessage(queueEntry.ChatID, queueEntry.Text)

	if queueEntry.Keyboard == tmodels.KeyboardTypeSubscribe {
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(t.Text.Subscribe),
			),
		)
	} else if queueEntry.Keyboard == tmodels.KeyboardTypeUnsubscribe {
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(t.Text.Unsubscribe),
			),
		)
	} else if queueEntry.Keyboard == tmodels.KeyboardTypeRemove {
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	}

	if _, err := t.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}
