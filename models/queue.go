package models

type KeyboardType int

const (
	KeyboardTypeNone        KeyboardType = 0
	KeyboardTypeSubscribe   KeyboardType = 1
	KeyboardTypeUnsubscribe KeyboardType = 2
	KeyboardTypeRemove      KeyboardType = 3
)

type QueueEntry struct {
	ChatID   int64
	Text     string
	Keyboard KeyboardType
}

func NewQueueEntry(chatID int64, text string, keyboard KeyboardType) QueueEntry {
	return QueueEntry{
		ChatID:   chatID,
		Text:     text,
		Keyboard: keyboard,
	}
}
