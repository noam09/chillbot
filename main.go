package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"

	"github.com/docopt/docopt-go"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/noam09/go-sickchill-api"
)

func main() {
	// Set up signal catching
	sigs := make(chan os.Signal, 1)
	// Catch all signals
	signal.Notify(sigs)
	// signal.Notify(sigs,syscall.SIGQUIT)
	// Method invoked on signal receive
	go func() {
		s := <-sigs
		log.Printf("Received signal: %s", s)
		AppCleanup()
		os.Exit(1)
	}()

	usage := `ChillBot

Usage:
  chillbot --token=<bot> --key=<apikey> -w <chatid>... [--host=<host>] [--port=<port>] [--base=<base>] [--ssl] [-d]
  chillbot -h | --help

Options:
  -h, --help                Show this screen.
  -t, --token=<bot>         Telegram bot token.
  -k, --key=<apikey>        API key.
  -w, --whitelist=<chatid>  Telegram chat ID(s) allowed to communicate with the bot (contact @myidbot).
  -o, --host=<host>         Hostname or address SickChill runs on [default: 127.0.0.1].
  -r, --port=<port>         Port SickChill runs on [default: 8081].
  -b, --base=<urlbase>      Path which should follow the base URL.
  -s, --ssl                 Use TLS/SSL (HTTPS) [default: false].
  -d, --debug               Debug logging [default: false].`

	opts, _ := docopt.ParseDoc(usage)
	log.Println(opts)

	// Get whitelist
	var whitelist []int
	w := opts["--whitelist"].([]string)
	for _, v := range w {
		i, err := strconv.Atoi(v)
		if err != nil {
			continue
		}
		whitelist = append(whitelist, i)
	}

	// Get arguments
	token, _ := opts.String("--token")
	hostname, _ := opts.String("--host")
	port, _ := opts.Int("--port")
	apiKey, _ := opts.String("--key")
	urlBase, _ := opts.String("--base")
	ssl, _ := opts.Bool("--ssl")
	debug, _ := opts.Bool("--debug")

	// Initialize SC client
	sr := sickchill.NewClient(hostname, port, apiKey, urlBase, ssl)

	BOT_TOKEN := token

	bot, err := tgbotapi.NewBotAPI(BOT_TOKEN)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = debug

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	tvdb, _ := regexp.Compile("tvdb ([0-9]{4,10})")
	tvdbOnly, _ := regexp.Compile("^#([0-9]{4,10})$")
	// emojiStar := "\u2b50\ufe0f"
	emojiFilm := "\U0001f4fa"
	// emojiSearch := "\U0001f50d"
	emojiCancel := "\u274e"

	for update := range updates {
		if update.Message == nil {
			continue
		}
		// Check if user ID in whitelist
		if !intInSlice(update.Message.From.ID, whitelist) {
			// log.Println("not me")
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		if tvdbOnly.MatchString(strings.ToLower(update.Message.Text)) {
			i := 0
			_, err := bot.Send(tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping))
			if err != nil {
				log.Println(err)
			}
			log.Println("Found #TVDBID")
			tvdbId := tvdbOnly.FindStringSubmatch(strings.ToLower(update.Message.Text))
			if len(tvdbId) == 2 {
				i, err = strconv.Atoi(tvdbId[1])
			} else {
				continue
			}
			// i, err := strconv.Atoi(tvdbId)
			// fmt.Println("i: ", i)
			if err != nil {
				fmt.Println(err)
				continue
			}
			val, err := sr.AddNewShow(i, "")
			if err != nil {
				log.Println(val)
			}
			msg.Text = fmt.Sprintf("Adding _#%d_ to snatchlist", i)
			msg.ParseMode = "markdown"
		} else if tvdb.MatchString(strings.ToLower(update.Message.Text)) {

			_, err := bot.Send(tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping))
			if err != nil {
				log.Println(err)
			}

			log.Println("Found TVDBID")
			tvdbId := tvdb.FindString(strings.ToLower(update.Message.Text))
			title, _ := regexp.Compile(`(.*?) \[.*?\]`)
			movieTitle := title.FindStringSubmatch(update.Message.Text)
			if len(movieTitle) > 1 {
				t := movieTitle[1]
				log.Println(t)
				words := strings.Fields(tvdbId)
				if len(words) == 2 {
					tvdbId = words[1]
				}
				i, err := strconv.Atoi(tvdbId)
				fmt.Println("i: ", i)
				if err != nil {
					fmt.Println(err)
					continue
					// os.Exit(2)
				}
				val, err := sr.AddNewShow(i, "")
				if err != nil {
					log.Println(val)
				}
				msg.Text = fmt.Sprintf("Adding _%s_ (%s) to snatchlist", t, tvdbId)
				msg.ParseMode = "markdown"
			} else {
				msg.Text = "Failed adding show to snatchlist"
			}
			// Remove the custom keyboard if it's still there
			msg.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{
				RemoveKeyboard: true,
				Selective:      false,
			}
		} else if strings.HasPrefix(update.Message.Text, "/q ") || !strings.HasPrefix(update.Message.Text, "/") {

			_, err := bot.Send(tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping))
			if err != nil {
				log.Println(err)
			}

			split := strings.Split(update.Message.Text, "/q ")
			q := ""
			if len(split) == 2 {
				_, q = split[0], split[1]
			} else {
				q = update.Message.Text
			}
			res, err := sr.SearchTVDB(strings.TrimSpace(q))
			resultCount := 0
			if res.Result == "success" {
				fmt.Println("Success")

				msgBody := ""
				// Make 2D slice with enough rows for the number of search results
				// Add one more row for the "/cancel" button
				rows := make([][]tgbotapi.KeyboardButton, len(res.Data.Results)+1)
				// Loop through results
				for _, a := range res.Data.Results {
					// Create button text for each search result
					button := fmt.Sprintf("%s [Aired: %s] [TVDB %d]", a.Name, a.FirstAired, a.TVDBID)
					// In each row, append one column containing a KeyboardButton with the button text
					rows[resultCount] = append(rows[resultCount], tgbotapi.NewKeyboardButton(button))
					log.Printf("%s [Aired: %s] [TVDBID %d]\n", a.Name, a.FirstAired, a.TVDBID)
					// Tally the number of results
					resultCount += 1
					msgBody += fmt.Sprintf("*%d)* [%s](https://www.tvtime.com/en/show/%d) _(Aired: %s)_ (TVDBID %d)\n",
						resultCount, a.Name, a.TVDBID, a.FirstAired, a.TVDBID)
				}
				// If there is at least one result ready, create the custom keyboard
				if resultCount > 0 {
					button := fmt.Sprintf("/cancel")
					rows[resultCount] = append(rows[resultCount], tgbotapi.NewKeyboardButton(button))
					// Init keyboard variable
					var kb tgbotapi.ReplyKeyboardMarkup
					kb = tgbotapi.ReplyKeyboardMarkup{
						ResizeKeyboard:  true,
						Keyboard:        rows,
						OneTimeKeyboard: true,
					}
					// kb.OneTimeKeyboard = true
					// Append the custom keyboard to the reply message
					msg.ReplyMarkup = kb
					// Append a hint to the results
					msgBody += "\n\nIf you were expecting more results, try running the same search again"
					// Append the list of results to the reply message
					msg.Text = msgBody
					msg.ParseMode = "markdown"
					// Avoid the first TVDB link being resolved for preview
					msg.DisableWebPagePreview = true
				} else {
					msgBody += "Looks like no results were found. "
					msgBody += "You might try running the same search again, "
					msgBody += "sometimes they get lost on the way back!"
					msg.Text = msgBody
				}
			} else {
				log.Println(err)
				msgBody := "Looks like no results were found."
				msg.Text = msgBody
			}
		} else {

			_, err := bot.Send(tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping))
			if err != nil {
				log.Println(err)
			}

			switch strings.TrimSpace(strings.ToLower(update.Message.Text)) {
			case "/start":
				msg.Text = "Hi! Use /q to start searching"
			case "/help", "/h":
				msg.Text = fmt.Sprintf("%s /q - TV show search\nOr simply send a query without any commands", emojiFilm)
				msg.Text += fmt.Sprintf("\nSend a hashtag and TVDB ID if you know exactly what you want (`#12345`).")
				msg.Text += fmt.Sprintf("\n\n%s /c - Cancel current operation", emojiCancel)
				msg.ParseMode = "markdown"
			case "/q":
				msg.Text = fmt.Sprintf("/q should be followed by a search query. Example:\n`/q Who Is America`\n")
				msg.ParseMode = "markdown"
			case "/cancel", "/c":
				// Cancel
				msg.Text = "Cancelling"
				msg.ReplyMarkup = tgbotapi.ReplyKeyboardRemove{
					RemoveKeyboard: true,
					Selective:      false,
				}

			default:
				msg.Text = ""
				msg.ParseMode = "markdown"
			}
		}

		if msg.Text != "" {
			bot.Send(msg)
		}
		continue
	} // end updates
} // end main

func intInSlice(a int, list []int) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func AppCleanup() {
	log.Println("Exiting...")
}
