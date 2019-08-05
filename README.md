# ChillBot

🤖 A simple [Telegram](https://telegram.org) bot for controlling [SickChill](https://github.com/SickChill/SickChill).

## Dependencies

* [go-sickchill-api](https://github.com/noam09/go-sickchill-api)
* [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)
* [docopt-go](https://github.com/docopt/docopt-go)

## Build

Clone this repository and `go build`:

```console
git clone https://github.com/noam09/chillbot
cd chillbot
go build main.go
```

## Install

Use `go install` to get and build ChillBot, making it available in `$GOPATH/bin/chillbot`:

```console
go get -u github.com/noam09/chillbot
go install github.com/noam09/chillbot
```

## docker-compose

A sample `docker-compose.yml` is included.

```console
# Pull the latest code
git clone https://github.com/noam09/chillbot
cd chillbot
# Modify YAML according to your setup
nano docker-compose.yml
# Run the container and send to background
docker-compose up -d
```

The `docker-compose.yml` is based on the official `golang:1.12.7-alpine` image:

```yaml
---
version: "2"

services:
  chillbot:
    image: golang:1.12.7-alpine
    volumes:
      - .:/go/src/chillbot
    working_dir: /go/src/chillbot
    command: >
      sh -c 'go run main.go
      --token=<bot>
      --key=<apikey>
      -w <chatid>
      --host=<host>
      --port=<port>
      --base=<urlbase>
      --ssl'
```

Modify the `command` section's parameters based on the help-text below.

## Usage

Running the bot:

```console
ChillBot

Usage:
  chillbot --token=<bot> -w <chatid>... [--host=<host>] [--port=<port>] [--base=<urlbase>] [--ssl] [-d]
  chillbot --token=<bot> --key=<apikey> -w <chatid>... [--host=<host>] [--port=<port>] [--base=<base>] [--ssl] [-d]
  chillbot -h | --help

Options:
  -h, --help                Show this screen.
  -t, --token=<bot>         Telegram bot token.
  -k, --key=<apikey>        API key.
  -w, --whitelist=<chatid>  Telegram chat ID(s) allowed to communicate with the bot (contact @myidbot).
  -o, --host=<host>         Hostname or address SickChill runs on [default: 127.0.0.1].
  -r, --port=<port>         Port SickChill runs on [default: 5050].
  -b, --base=<urlbase>      Path which should follow the base URL.
  -s, --ssl                 Use TLS/SSL (HTTPS) [default: false].
  -d, --debug               Debug logging [default: false].
```

Controlling the bot:

```
📺 /q - TV show search
Or simply send a query without any commands
Send a hashtag and TVDB ID if you know exactly what you want (#12345).

❎ /c - Cancel current operation
```

**💡 Protip!** Sending ChillBot a hashtag followed by a TVDB series ID (e.g. `#123456`) will add the series to the snatchlist.

## TODO

* Makefile
* systemd service file
* Add group command support (`/command@bot`, untested)
* Check if exists in library
* On-the-fly user whitelisting?
* Choose quality profile other than the default

## Development

Contributions are always welcome, just create a [pull request](https://github.com/noam09/chillbot/pulls) or [submit an issue](https://github.com/noam09/chillbot/issues).

## License

This is free software under the GPL v3 open source license. Feel free to do with it what you wish, but any modification must be open sourced. A copy of the license is included.
