# MOC to Telegram [![](https://images.microbadger.com/badges/image/ctfl/moc-telegram.svg)](https://hub.docker.com/r/ctfl/moc-telegram "DockerHub Image")

Sends MOC Messages to Telegram Users. All messages to be sent are queued. Thus, the telegram Ratelimit is not violated.

## Configuration

```bash
API_ENDPOINT=http://moc
TELEGRAM_TOKEN=

TEXT_SUBSCRIBE=Abonnieren
TEXT_UNSUBSCRIBE=Deabonnieren
TEXT_HELLO=Willkommen
TEXT_SUBSCRIBED=Abonniert!
TEXT_UNSUBSCRIBED=Deabonniert!
```

## Usage

```
moc-telegram
```

## Docker Compose

```yaml
version: '3'
services:
  moc:
    image: ctfl/moc
    env_file:
      - .env
    ports:
      - 80:80
  moc-telegram:
    image: ctfl/moc-telegram
    env_file:
      - .env
```
