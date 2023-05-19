package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
)

func main() {
	config := viper.New()
	config.BindEnv("mongo_url", "LB_MONGO_URL")
	config.BindEnv("mongo_password", "LB_MONGO_PASSWORD")
	config.BindEnv("mongo_user", "LB_MONGO_USER")
	config.BindEnv("depth", "LB_DEPTH")
	config.BindEnv("telegram_apitoken", "LB_TELEGRAM_APITOKEN")

	mongo_url := config.GetString("mongo_url")
	mongo_password := config.GetString("mongo_password")
	mongo_user := config.GetString("mongo_user")
	depth := config.GetInt64("depth")
	telegram_apitoken := config.GetString("telegram_apitoken")

	bot, err := tgbotapi.NewBotAPI(telegram_apitoken)
	if err != nil {
		panic(err)
	}
	client := connection(mongo_url, mongo_user, mongo_password)
	defer Disconnect(client)

	bot.Debug = true

	// Create a new UpdateConfig struct with an offset of 0. Offsets are used
	// to make sure Telegram knows we've handled previous values and we don't
	// need them repeated.
	updateConfig := tgbotapi.NewUpdate(0)

	// Tell Telegram we should wait up to 30 seconds on each request for an
	// update. This way we can get information just as quickly as making many
	// frequent requests without having to send nearly as many.
	updateConfig.Timeout = 30

	// Start polling Telegram for updates.
	updates := bot.GetUpdatesChan(updateConfig)

	// Let's go through each update that we're getting from Telegram.
	for update := range updates {
		// Telegram can send many types of updates depending on what your Bot
		// is up to. We only want to look at messages for now, so we can
		// discard any other updates.
		if update.Message == nil {
			continue
		}

		// Now that we know we've gotten a new message, we can construct a
		// reply! We'll take the Chat ID and Text from the incoming message
		// and use it to create a new message.

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		switch update.Message.Command() {
		case "help":
			msg.Text = "I understand /task"
		case "task":
			// todo: string and int64 update.Message.From.ID
			increaseState(update.Message.From.ID, client, depth)
			msg.Text = askWord(update.Message.From.ID)
		default:
			msg.ParseMode = "HTML"
			msg.Text = checkAnswer(update.Message.From.ID, update.Message.Text)
		}
		// We'll also say that this message is a reply to the previous message.
		// For any other specifications than Chat ID or Text, you'll need to
		// set fields on the `MessageConfig`.
		msg.ReplyToMessageID = update.Message.MessageID

		// Okay, we're sending our message off! We don't care about the message
		// we just sent, so we'll discard it.
		if _, err := bot.Send(msg); err != nil {
			// Note that panics are a bad way to handle errors. Telegram can
			// have service outages or network errors, you should retry sending
			// messages or more gracefully handle failures.
			panic(err)
		}
	}
}
