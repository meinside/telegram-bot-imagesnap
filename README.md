# Telegram Bot for Capturing Images with ImageSnap

With this bot, you can capture images with ImageSnap on your Mac.

## 0. Prepare

Install Go and generate your Telegram bot's API token.

Also, you need to install [ImageSnap](http://iharder.sourceforge.net/current/macosx/imagesnap/):

```bash
$ brew install imagesnap
```

## 1. Install and configure

```bash
$ go get -u github.com/meinside/telegram-bot-imagesnap
$ cd $GOPOATH/src/github.com/meinside/telegram-bot-imagesnap
$ cp config.json.sample config.json
$ vi config.json
```

and edit values to yours:

```json
{
	"api_token": "0123456789:abcdefghijklmnopqrstuvwyz-x-0a1b2c3d4e",
	"available_ids": [
		"telegram_id_1",
		"telegram_id_2",
		"telegram_id_3"
	],
	"monitor_interval": 3,
	"is_verbose": false
}
```

## 2. Build and run

```bash
$ go build -o telegrambot main.go
```

and run it:

```bash
$ ./telegrambot
```

## 3. Run as a service

### a. launchctl

Copy sample .plist file:

```bash
$ sudo cp $GOPOATH/src/github.com/meinside/telegram-bot-imagesnap/launchd/telegram-imagesnap.plist /Library/LaunchDaemons/telegram-imagesnap.plist
```

and edit values:

```
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>telegram-imagesnap</string>
	<key>ProgramArguments</key>
	<array>
		<string>/path/to/telegram-imagesnap/telegrambot</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
</dict>
</plist>
```

Now load it with:

```bash
$ sudo launchctl load /Library/LaunchDaemons/telegram-imagesnap.plist
```

## 998. Trouble shooting

TODO

## 999. License

MIT

