package config

import (
	env "github.com/Netflix/go-env"
	log "github.com/sirupsen/logrus"
)

type Text struct {
	Subscribe    string `env:"TEXT_SUBSCRIBE"`
	Unsubscribe  string `env:"TEXT_UNSUBSCRIBE"`
	Hello        string `env:"TEXT_HELLO"`
	Subscribed   string `env:"TEXT_SUBSCRIBED"`
	Unsubscribed string `env:"TEXT_UNSUBSCRIBED"`
}

type Config struct {
	Endpoint string `env:"API_ENDPOINT"`

	TelegramToken string `env:"TELEGRAM_TOKEN"`

	Text Text
}

// ReadConfig from env
func ReadConfig() *Config {
	var config Config
	_, err := env.UnmarshalFromEnviron(&config)
	if err != nil {
		log.Fatal(err)
	}

	if config.Endpoint == "" {
		log.Fatal("Need API_ENDPOINT env var")
	}

	if config.TelegramToken == "" {
		log.Fatal("Need TELEGRAM_TOKEN env var")
	}

	if config.Text.Subscribe == "" {
		config.Text.Subscribe = "Subscribe"
	}

	if config.Text.Unsubscribe == "" {
		config.Text.Unsubscribe = "Unsubscribe"
	}

	if config.Text.Hello == "" {
		config.Text.Hello = "Hi."
	}

	if config.Text.Subscribed == "" {
		config.Text.Subscribed = "You have subscribed."
	}

	if config.Text.Unsubscribed == "" {
		config.Text.Unsubscribed = "Bye."
	}

	return &config
}
