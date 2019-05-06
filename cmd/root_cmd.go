package cmd

import (
	"github.com/chaostreff-flensburg/moc-telegram/config"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// rootCmd will run the log streamer
var rootCmd = cobra.Command{
	Use:  "moc-telegram",
	Long: "A service that will serve a telegram message operation center endpoint",
	Run: func(cmd *cobra.Command, args []string) {
		execWithConfig(cmd, moc2telegram)
	},
}

// RootCmd will add flags and subcommands to the different commands
func RootCmd() *cobra.Command {
	rootCmd.AddCommand(&moc2telegramCmd)
	return &rootCmd
}

// execWithConfig load config from env
func execWithConfig(cmd *cobra.Command, fn func(config *config.Config)) {
	logrus.Info("Read Config...")
	config := config.ReadConfig()

	fn(config)
}
