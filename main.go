package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	// Menu texts
	firstMenu  = "<b>Menu 1</b>\n\nA beautiful menu with a shiny inline button."
	secondMenu = "<b>Menu 2</b>\n\nA better menu with even more shiny inline buttons."

	// Button texts
	nextButton     = "Next"
	backButton     = "Back"
	tutorialButton = "Tutorial"

	// Store bot screaming status
	screaming = false
	bot       *tgbotapi.BotAPI
	botscope  *tgbotapi.BotAPI

	// Keyboard layout for the first menu. One button, one row
	firstMenuMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(nextButton, nextButton),
		),
	)

	// Keyboard layout for the second menu. Two buttons, one per row
	secondMenuMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(backButton, backButton),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(tutorialButton, "https://core.telegram.org/bots/api"),
		),
	)
)

func main() {
	var err error
	botToken := os.Getenv("YOUR_BOT_TOKEN")
	bot, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		// Abort if something is wrong
		log.Panic(err)
	}

	// baseURL := "https://api.telegram.org/bot"
	// endpoint := baseURL + botToken + "/sendMessage&Updates"
	// endpoint2 := baseURL + botToken + "/chatJoinRequest"
	// Set this to true to log all interactions with telegram servers
	bot.Debug = false

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	// tes := tgbotapi.NewUpdate(1)
	// ApproveChatJoinRequestConfig()
	// Create a new cancellable background context. Calling `cancel()` leads to the cancellation of the context
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// // `updates` is a golang channel which receives telegram updates
	updates := bot.GetUpdatesChan(u)
	// upd := bot.GetUpdatesChan()

	// Pass cancellable context to goroutine
	go receiveUpdates(ctx, updates)

	// Tell the user the bot is online
	log.Println("Start listening for updates. Press enter to stop")

	// Wait for a newline symbol, then cancel handling updates
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	cancel()

	http.HandleFunc("/hello", helloHandler)
	fmt.Println("Server Start Up........")
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

func receiveUpdates(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	// `for {` means the loop is infinite until we manually stop it
	for {
		select {
		// stop looping if ctx is cancelled
		case <-ctx.Done():
			return
		// receive update from channel and then handle it
		case update := <-updates:
			handleUpdate(update)
		}
	}
}

func handleUpdate(update tgbotapi.Update) {
	switch {
	case update.ChannelPost != nil:
		log.Printf("ChannelPost")
		break
	case update.ChatJoinRequest != nil:
		log.Printf("ChatJoinRequest")
		handleChatJoinRequest(update.ChatJoinRequest)
		break
	// Handle messages
	case update.Message != nil:
		// log.Printf("Message")
		handleMessage(update.Message)
		break

	// Handle button clicks
	case update.CallbackQuery != nil:
		log.Printf("CallbackQuery")
		handleButton(update.CallbackQuery)
		break
	default:
		// log.Printf("%v", update)
		// webhook, _ := bot.GetWebhookInfo()
		// log.Printf("%v", webhook)
		break
	}
}

func handleChatJoinRequest(member *tgbotapi.ChatJoinRequest) {
	// str := fmt.Sprintf("%+v\n", member)
	// msg := tgbotapi.NewMessage(member.Chat.ID, str)
	// _, err = bot.Send(msg)

	// if err != nil {
	// 	log.Printf("An error occured: %s", err.Error())
	// }
}

func handleMessage(message *tgbotapi.Message) {
	user := message.From
	text := message.Text
	newChatMembers := message.NewChatMembers

	for _, chatMember := range newChatMembers {
		if chatMember.LanguageCode == "jp" {
			log.Printf("join jp id : %d firstName : %s LastName : %s jp %s", chatMember.ID, chatMember.FirstName, chatMember.LastName, chatMember.LanguageCode)
		} else {
			log.Printf("join not jp id : %d firstName : %s LastName : %s jp %s", chatMember.ID, chatMember.FirstName, chatMember.LastName, chatMember.LanguageCode)
			memberConfig := tgbotapi.ChatMemberConfig{ChatID: message.Chat.ID, ChannelUsername: chatMember.UserName, UserID: chatMember.ID}
			banChatMemberConfig := tgbotapi.BanChatMemberConfig{ChatMemberConfig: memberConfig, RevokeMessages: false}
			_, _ = bot.Request(banChatMemberConfig)
			return
		}
	}

	if user == nil {
		return
	}

	log.Printf("%s (%s) wrote %s", user.FirstName, message.From.LanguageCode, text)
	if isExtraMessage(message) ||
		strings.Contains(message.Text, "given away") ||
		strings.Contains(message.Text, "❗️❗️❗️@") ||
		strings.Contains(message.Text, "reward_bot") ||
		strings.Contains(message.Text, "Take here") {
		deletemsag := tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID)
		_, _ = bot.Send(deletemsag)
		// var lowerName = strings.ToLower(user.FirstName)
		// if strings.Contains(user.FirstName, "АirDrор") || strings.Contains(lowerName, "bitgеt") || strings.Contains(lowerName, "airdrop") || strings.Contains(lowerName, "official") || strings.Contains(lowerName, "bot") {
		log.Printf("del message id : %d firstName : %s LastName : %s BAN! %s", user.ID, user.FirstName, user.LastName, text)
		memberConfig := tgbotapi.ChatMemberConfig{ChatID: message.Chat.ID, ChannelUsername: user.UserName, UserID: user.ID}
		restrictChatMemberConfig := tgbotapi.RestrictChatMemberConfig{ChatMemberConfig: memberConfig,
			Permissions: &tgbotapi.ChatPermissions{CanSendMessages: false,
				CanSendMediaMessages:  false,
				CanSendPolls:          false,
				CanSendOtherMessages:  false,
				CanAddWebPagePreviews: false,
				CanChangeInfo:         false,
				CanInviteUsers:        false,
				CanPinMessages:        false}}
		banChatMemberConfig := tgbotapi.BanChatMemberConfig{ChatMemberConfig: memberConfig, RevokeMessages: false}
		_, _ = bot.Request(restrictChatMemberConfig)
		_, _ = bot.Request(banChatMemberConfig)
		return
		// } else {
		// 	log.Printf("del message id : %d firstName : %s LastName : %s Del only", user.ID, user.FirstName, user.LastName)
		// }
	}

	if user.UserName == "" {
		log.Printf("del message id : %d firstName : %s LastName : %s DELETE! %s", user.ID, user.FirstName, user.LastName, text)
		deletemsag := tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID)
		_, _ = bot.Send(deletemsag)
		return
	}

	var err error
	if message.Text == "/testBotCommand" {
		// botScope := tgbotapi.NewBotCommandScopeChatMember(message.Chat.ID, user.ID)
		// msg := tgbotapi.NewMessage(message.Chat.ID, strings.ToUpper(text))
		// bot.Request
		// _, err = botScope.Send(msg)
	}
	// webhook, err := bot.GetWebhookInfo()
	// log.Printf("%v", webhook)
	// if strings.HasPrefix(text, "/") {
	// 	err = handleCommand(message.Chat.ID, text)
	// }
	//  else if screaming && len(text) > 0 {
	// 	msg := tgbotapi.NewMessage(message.Chat.ID, strings.ToUpper(text))
	// 	// To preserve markdown, we attach entities (bold, italic..)
	// 	msg.Entities = message.Entities
	// 	// _, err = bot.Send(msg)
	// 	log.Printf("%s wrote %s", user.FirstName, msg)
	// } else {
	// 	// This is equivalent to forwarding, without the sender's name
	// 	copyMsg := tgbotapi.NewCopyMessage(message.Chat.ID, message.Chat.ID, message.MessageID)
	// 	// _, err = bot.CopyMessage(copyMsg)
	// 	log.Printf("%s wrote %s", user.FirstName, copyMsg)
	// }

	if err != nil {
		log.Printf("An error occured: %s", err.Error())
	}
}

func isExtraMessage(message *tgbotapi.Message) bool {
	var result = false
	if message.Text == "" &&
		message.ForwardFrom == nil &&
		message.ForwardFromChat == nil &&
		message.ReplyToMessage == nil &&
		message.Entities == nil &&
		message.Animation == nil &&
		message.Audio == nil &&
		message.Document == nil &&
		message.Photo == nil &&
		message.Sticker == nil &&
		message.Video == nil &&
		message.VideoNote == nil &&
		message.Voice == nil &&
		message.CaptionEntities == nil &&
		message.Contact == nil &&
		message.Dice == nil &&
		message.Game == nil &&
		message.Poll == nil &&
		message.Venue == nil &&
		message.Location == nil &&
		message.NewChatMembers == nil &&
		message.LeftChatMember == nil &&
		message.NewChatPhoto == nil &&
		message.MessageAutoDeleteTimerChanged == nil &&
		message.PinnedMessage == nil &&
		message.Invoice == nil &&
		message.SuccessfulPayment == nil &&
		message.PassportData == nil &&
		message.ProximityAlertTriggered == nil &&
		message.VoiceChatScheduled == nil &&
		message.VoiceChatStarted == nil &&
		message.VoiceChatEnded == nil &&
		message.VoiceChatParticipantsInvited == nil &&
		message.ReplyMarkup == nil {
		result = true
	}
	return result
}

// When we get a command, we react accordingly
func handleCommand(chatId int64, command string) error {
	var err error

	// switch command {
	// case "/testBotCommand":
	// 	err = testBotCommand()
	// 	break
	// case "/scream":
	// 	screaming = true
	// 	break

	// case "/whisper":
	// 	screaming = false
	// 	break

	// case "/menu":
	// 	err = sendMenu(chatId)
	// 	break
	// }

	return err
}

func handleButton(query *tgbotapi.CallbackQuery) {
	var text string

	markup := tgbotapi.NewInlineKeyboardMarkup()
	message := query.Message

	if query.Data == nextButton {
		text = secondMenu
		markup = secondMenuMarkup
	} else if query.Data == backButton {
		text = firstMenu
		markup = firstMenuMarkup
	}

	callbackCfg := tgbotapi.NewCallback(query.ID, "")
	bot.Send(callbackCfg)

	// Replace menu text and keyboard
	msg := tgbotapi.NewEditMessageTextAndMarkup(message.Chat.ID, message.MessageID, text, markup)
	msg.ParseMode = tgbotapi.ModeHTML
	bot.Send(msg)
}

//	func removeScam() error {
//		var err : error
//		tgbotapi.
//		return err
//	}
func sendMenu(chatId int64) error {
	msg := tgbotapi.NewMessage(chatId, firstMenu)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = firstMenuMarkup
	_, err := bot.Send(msg)
	return err
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	hello := []byte("Hello World!!!")
	_, err := w.Write(hello)
	if err != nil {
		log.Fatal(err)
	}
}
