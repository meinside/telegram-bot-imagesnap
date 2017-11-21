// telegram bot for capturing image with ImageSnap
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	bot "github.com/meinside/telegram-bot-go"

	"github.com/meinside/telegram-bot-imagesnap/conf"
	"github.com/meinside/telegram-bot-imagesnap/helper"
)

type status int16

const (
	statusWaiting status = iota
)

const (
	imageSnapBin = "/usr/local/bin/imagesnap" // XXX - brew install imagesnap
	tempDir      = "/tmp"
)

type Session struct {
	UserId        string
	CurrentStatus status
}

type sessionPool struct {
	Sessions map[string]Session
	sync.Mutex
}

// variables
var apiToken string
var monitorInterval int
var isVerbose bool
var availableIds []string
var pool sessionPool
var launched time.Time

// keyboards
var allKeyboards = [][]bot.KeyboardButton{
	bot.NewKeyboardButtons(conf.CommandCapture),
	bot.NewKeyboardButtons(conf.CommandStatus, conf.CommandHelp),
}
var cancelKeyboard = [][]bot.KeyboardButton{
	bot.NewKeyboardButtons(conf.CommandCancel),
}

// initialization
func init() {
	launched = time.Now()

	// read variables from config file
	if config, err := helper.GetConfig(); err == nil {
		apiToken = config.ApiToken
		availableIds = config.AvailableIds
		monitorInterval = config.MonitorInterval
		if monitorInterval <= 0 {
			monitorInterval = conf.DefaultMonitorIntervalSeconds
		}
		isVerbose = config.IsVerbose

		// initialize variables
		sessions := make(map[string]Session)
		for _, v := range availableIds {
			sessions[v] = Session{
				UserId:        v,
				CurrentStatus: statusWaiting,
			}
		}
		pool = sessionPool{
			Sessions: sessions,
		}
	} else {
		panic(err.Error())
	}
}

// check if given Telegram id is available
func isAvailableId(id string) bool {
	for _, v := range availableIds {
		if v == id {
			return true
		}
	}
	return false
}

// for showing help message
func getHelp() string {
	return `
Following commands are supported:

*For ImageSnap*

/capture : capture an image with ImageSnap

*Others*

/status : show this bot's status
/help : show this help message
`
}

// for showing current status of this bot
func getStatus() string {
	return fmt.Sprintf("Uptime: %s\nMemory Usage: %s", helper.GetUptime(launched), helper.GetMemoryUsage())
}

// process incoming update from Telegram
func processUpdate(b *bot.Bot, update bot.Update) bool {
	// check username
	var userId string
	if update.Message.From.Username == nil {
		log.Printf("*** Not allowed (no user name): %s", update.Message.From.FirstName)
		return false
	}
	userId = *update.Message.From.Username
	if !isAvailableId(userId) {
		log.Printf("*** Id not allowed: %s", userId)
		return false
	}

	// process result
	result := false

	pool.Lock()
	if session, exists := pool.Sessions[userId]; exists {
		// text from message
		var txt string
		if update.Message.HasText() {
			txt = *update.Message.Text
		} else {
			txt = ""
		}

		var message string
		var options = map[string]interface{}{
			"reply_markup": bot.ReplyKeyboardMarkup{
				Keyboard:       allKeyboards,
				ResizeKeyboard: true,
			},
			"parse_mode": bot.ParseModeMarkdown,
		}

		switch session.CurrentStatus {
		case statusWaiting:
			switch {
			// start
			case strings.HasPrefix(txt, conf.CommandStart):
				message = conf.MessageDefault
			// capture
			case strings.HasPrefix(txt, conf.CommandCapture):
				message = ""
			// status
			case strings.HasPrefix(txt, conf.CommandStatus):
				message = getStatus()
			// help
			case strings.HasPrefix(txt, conf.CommandHelp):
				message = getHelp()
			// fallback
			default:
				message = fmt.Sprintf("*%s*: %s", txt, conf.MessageUnknownCommand)
			}
		}

		if len(message) > 0 {
			// send message
			if sent := b.SendMessage(update.Message.Chat.Id, message, options); sent.Ok {
				result = true
			} else {
				log.Printf("*** Failed to send message: %s", *sent.Description)
			}
		} else {
			// typing...
			b.SendChatAction(update.Message.Chat.Id, bot.ChatActionTyping)

			// send photo
			if filepath, err := captureImageSnap(); err == nil {
				if sent := b.SendPhoto(update.Message.Chat.Id, bot.InputFileFromFilepath(filepath), options); sent.Ok {
					if err := os.Remove(filepath); err != nil {
						log.Printf("*** Failed to delete temp file: %s", err)
					}
					result = true
				} else {
					log.Printf("*** Failed to send photo: %s", *sent.Description)
				}
			} else {
				log.Printf("*** Image capture failed: %s", err)
			}
		}
	} else {
		log.Printf("*** Session does not exist for id: %s", userId)
	}
	pool.Unlock()

	return result
}

func captureImageSnap() (filepath string, err error) {
	filepath = fmt.Sprintf("%s/captured_`date +%Y%m%d_%H%M`.jpg", tempDir)
	if _, err := exec.Command(imageSnapBin, filepath).CombinedOutput(); err != nil {
		return "", err
	}
	return filepath, nil
}

func main() {
	client := bot.NewClient(apiToken)
	client.Verbose = isVerbose

	// get info about this bot
	if me := client.GetMe(); me.Ok {
		log.Printf("Launching bot: @%s (%s)", *me.Result.Username, me.Result.FirstName)

		// delete webhook (getting updates will not work when wehbook is set up)
		if unhooked := client.DeleteWebhook(); unhooked.Ok {
			// wait for new updates
			client.StartMonitoringUpdates(0, monitorInterval, func(b *bot.Bot, update bot.Update, err error) {
				if err == nil {
					if update.Message != nil {
						processUpdate(b, update)
					}
				} else {
					log.Printf("*** Error while receiving update (%s)", err.Error())
				}
			})
		} else {
			panic("Failed to delete webhook")
		}
	} else {
		panic("Failed to get info of the bot")
	}
}
