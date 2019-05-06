package cmd

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/syndtr/goleveldb/leveldb"

	api "github.com/chaostreff-flensburg/moc-go"
	"github.com/chaostreff-flensburg/moc-telegram/config"
	"github.com/chaostreff-flensburg/moc-telegram/telegram"
)

var moc2telegramCmd = cobra.Command{
	Use:   "moc2telegram",
	Short: "MOC2Telegram",
	Long:  "Transfer moc messages to telegram.",
	Run: func(cmd *cobra.Command, args []string) {
		execWithConfig(cmd, moc2telegram)
	},
}

// moc2telegram start moc2telegram
func moc2telegram(config *config.Config) {
	log.Info("Start")

	log.Info("Init LevelDB")
	db, err := leveldb.OpenFile("/data/db.leveldb", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Info("Init Telegram")
	telegramClient := telegram.NewTelegram(config, db)
	telegramClient.Connect()
	go telegramClient.Loop()

	log.Info("Telegram Running!")

	go telegramClient.SendLoop(0.1 * time.Second)

	apiClient := api.NewClient(config.Endpoint)
	apiClient.Loop(20 * time.Second)

	for message := range apiClient.NewMessages {
		log.Info(message.ID)
		telegramClient.SendAll(message)
	}
}
