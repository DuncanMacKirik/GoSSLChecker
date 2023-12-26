package main

import (
	"crypto/tls"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"
	"github.com/cloudfoundry/jibber_jabber"
	"github.com/urfave/cli/v2"
	"golang.org/x/text/message"
	"golang.org/x/text/language"
)

var GetVer bool
var MinDays int
var SendDelay int
var MaxTries int
var VerboseMode bool
var TgmToken string
var TgmChatId string

var p *message.Printer
var matcher language.Matcher

const APP_VERSION           = "0.5.0"

const LNG_PROGRAM_USAGE     = "Check for SSL/TLS certificates that are expiring soon, and report them to the specified Telegram chat"
const LNG_APP_VERSION       = "%s version %s\n"
const LNG_LANG_EN           = "force usage of English language, instead of cheking the OS defaults"
const LNG_LANG_RU           = "force usage of Russian language, instead of cheking the OS defaults"
const LNG_GET_VERSION       = "print program version and exit"
const LNG_GET_HELP          = "print program usage information and exit"
const LNG_VERBOSE_MODE      = "verbose mode"
const LNG_CERT_MIN_DAYS     = "minimal remaining active days for a certificate"
const LNG_DELAY_BTW_SND_ATT = "delay between message sending attempts (in seconds)"
const LNG_MAX_NUM_SND_ATT   = "maximum number of message sending attempts"
const LNG_TGM_TOKEN         = "Telegram token for sending messsages"
const LNG_TGM_CHATID        = "Telegram chat id for sending messsages"
const LNG_SRV_NAMES         = "server name(s) to check (separated by spaces)"
const LNG_ERR_MISSING_PAR   = "ERROR: missing required parameter(s)!\n"
const LNG_ERR_INVALID_FLAG  = "ERROR: unknown flag specified: -" // prefix string
const LNG_ERR_MISSING_URL   = "ERROR: no URL(s) specified for checking!\n"
const LNG_ERR_GETCERT_SRV_S = "ERROR: problem getting certificate for server %s: %s\n"
const LNG_ERR_NAME_MISM_S   = "ERROR: name mismatch for server's certificate - %s: %s\n"
const LNG_ERR_STATUS_N200_D = "ERROR: response status code is not OK (200): %d"
const LNG_ERR_SEND_FAIL_S_D = "ERROR: message sending failed (%s)! Pausing for %d s...\n"
const LNG_ERR_SEND_FAIL_R_D = "Failed to send message after %d retries!\n"
const LNG_ERRORS_FOUND_S    = "ERRORS FOUND:\n%s\n"
const LNG_CHECKING_URL      = "Checking URL: %s\n"
const LNG_SENDING           = "Sending...\n"
const LNG_OK                = "OK!\n"
const LNG_SERVER_S          = "Server: %s\n"
const LNG_ISSUER_S          = "Issuer: %s\n"
const LNG_EXPIRES_V         = "Expires: %v\n"
const LNG_DAYSLEFT_D        = "%d days left\n"

var lngStrings = []struct {
	en string;
	ru string;
}{
	{en: LNG_PROGRAM_USAGE, ru: "проверить сертификаты SSL/TLS на предмет скорого истечения срока, с отправкой предупреждений в чат Telegram"},
	{en: LNG_APP_VERSION, ru: "%s, версия %s\n"},
	{en: LNG_LANG_EN, ru: "использовать английский язык (вместо попытки автоопределения языка ОС)"},
	{en: LNG_LANG_RU, ru: "использовать русский язык (вместо попытки автоопределения языка ОС)"},
	{en: LNG_GET_VERSION, ru: "показать версию программы и выйти"},
	{en: LNG_GET_HELP, ru: "показать короткую справку об использовании программы и выйти"},
	{en: LNG_VERBOSE_MODE, ru: "включить вывод подробной информации"},
	{en: LNG_CERT_MIN_DAYS, ru: "минимально допустимое время истечения сертификата (в днях)"},
	{en: LNG_DELAY_BTW_SND_ATT, ru: "длительность задержки между попытками отправки сообщений в Telegram (в секундах)"},
	{en: LNG_MAX_NUM_SND_ATT, ru: "максимальное количество попыток отправки сообщений в Telegram"},
	{en: LNG_TGM_TOKEN, ru: "значение токена Telegram для отправки сообщений"},
	{en: LNG_TGM_CHATID, ru: "значение chat id Telegram для отправки сообщений"},
	{en: LNG_SRV_NAMES, ru: "имя (имена) серверов для проверки (через пробел)"},
	{en: LNG_ERR_MISSING_PAR, ru: "ОШИБКА: не задан(ы) один или более обязательных параметров!\n"},
	{en: LNG_ERR_MISSING_URL, ru: "ОШИБКА: не задан(ы) один или более URL сервера(-ов) для проверки!\n"},
	{en: LNG_ERR_GETCERT_SRV_S, ru: "ОШИБКА: невозможно получить сертификат для сервера %s: %s\n"},
	{en: LNG_ERR_NAME_MISM_S, ru: "ОШИБКА: неверное имя в сертификате сервера - %s: %s\n"},
	{en: LNG_ERR_STATUS_N200_D, ru: "ОШИБКА: код статуса в ответе сервера ошибочный (не 200): %d"},
	{en: LNG_ERR_SEND_FAIL_S_D, ru: "ОШИБКА: невозможно послать сообщение (%s)! Делаем паузу (%d с)...\n"},
	{en: LNG_ERR_SEND_FAIL_R_D, ru: "Не удалось отправить сообщение с %d попыток!\n"},
	{en: LNG_ERRORS_FOUND_S, ru: "НАЙДЕННЫЕ ОШИБКИ:\n%s\n"},
	{en: LNG_CHECKING_URL, ru: "Проверка URL: %s\n"},
	{en: LNG_SENDING, ru: "Отправка...\n"},
	{en: LNG_OK, ru: "ОК!\n"},
	{en: LNG_SERVER_S, ru:  "Сервер: %s\n"},
	{en: LNG_ISSUER_S, ru:  "Выдан: %s\n"},
	{en: LNG_EXPIRES_V, ru:  "Истекает: %v\n"},
	{en: LNG_DAYSLEFT_D, ru: "Осталось: %d дней\n"},
}

func initLangs() {
	for _, strs := range lngStrings {
		message.SetString(language.AmericanEnglish, strs.en, strs.en)
		message.SetString(language.Russian, strs.en, strs.ru)
	}
}

func issuer(info string) string {
	params := strings.Split(info, ",")
	cn := "??"
	org := "??"
	for _, value := range params {
		args := strings.Split(value, "=")
		if args[0] == "CN" {
			cn = args[1]
		}
		if args[0] == "O" {
			org = args[1]
		}
	}
	return fmt.Sprintf("%s (%s)", org, cn)
}

func verifyHostname(conn *tls.Conn, url string) (err error) {
	defer func() {
		if r := recover(); r != nil {
		        switch x := r.(type) {
		        	case string:
		        		err = errors.New(x)
		        	case error:
		        		err = x
		        	default:
		        		err = errors.New("Unknown panic")
		        }
		}
	}()
	err = conn.VerifyHostname(url)
	if err != nil {
		return err
	}
	return nil
}

func fail(msg string) {
        fmt.Fprintf(os.Stderr, msg)
        os.Exit(2)
}

func chk(url string) string {
	msg := ""
        errMsg := ""
	conn, err := tls.Dial("tcp", url + ":443", nil)
	if err != nil {
		errMsg = errMsg + p.Sprintf(LNG_ERR_GETCERT_SRV_S, url, err.Error())
		return errMsg
	}

	err = verifyHostname(conn, url)
	if err != nil {
		errMsg = errMsg + p.Sprintf(LNG_ERR_NAME_MISM_S, url, err.Error())
		return errMsg
	}

	expiry := conn.ConnectionState().PeerCertificates[0].NotAfter
	currentTime := time.Now()
        diff := expiry.Sub(currentTime)
        daysLeft := int(math.Round(diff.Hours() / 24))
        msg = msg + p.Sprintf(LNG_SERVER_S, url)
	msg = msg + p.Sprintf(LNG_ISSUER_S, issuer(fmt.Sprintf("%s", conn.ConnectionState().PeerCertificates[0].Issuer)))
	msg = msg + p.Sprintf(LNG_EXPIRES_V, expiry.Format(time.RFC1123))
	msg = msg + p.Sprintf(LNG_DAYSLEFT_D, daysLeft)
	msg = msg + "=================\n"
	fmt.Printf(msg)
	if daysLeft <= MinDays {
		errMsg = msg
	}
	return errMsg
}

func getUrl() string {
	return fmt.Sprintf("https://api.telegram.org/bot%s", TgmToken)
}

func sendMessage(text string) (bool, error) {
	var err error
	var response *http.Response

	url := fmt.Sprintf("%s/sendMessage", getUrl())
	body, _ := json.Marshal(map[string]string{
		"chat_id": TgmChatId,
		"text":    text,
	})
	response, err = http.Post(
		url,
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return false, err
	}

	defer response.Body.Close()

	body, err = ioutil.ReadAll(response.Body)

	if err != nil {
		return false, err
	}

	if response.StatusCode != 200 {
		err = errors.New(p.Sprintf(LNG_ERR_STATUS_N200_D, response.StatusCode))
		return false, err
	}

	return true, nil
}

func printArg(name, val string) {
	str := p.Sprintf(name)
	r := []rune(str)
	r[0] = unicode.ToUpper(r[0])
	str = string(r)
        fmt.Printf("%s: %v\n", str, val)
}

func printArgs() {
	names :=  [...]string{LNG_CERT_MIN_DAYS, LNG_DELAY_BTW_SND_ATT, LNG_MAX_NUM_SND_ATT, 
		LNG_TGM_TOKEN, LNG_TGM_CHATID, LNG_VERBOSE_MODE}
	values := [...]string{fmt.Sprintf("%d", MinDays), fmt.Sprintf("%d", SendDelay), fmt.Sprintf("%d", MaxTries),
		TgmToken, TgmChatId, fmt.Sprintf("%t", VerboseMode)}
	for i, name := range names {
		printArg(name, values[i])
	}
}

func run(Args cli.Args) {
	if VerboseMode {
		//App.Version()
		printArgs()
	}

	msg := ""
	for i := 0; i < Args.Len(); i++ {
		url := Args.Get(i)
		if VerboseMode {
			p.Printf(LNG_CHECKING_URL, url)
		}
		msg = msg + chk(url)
	}

	duration := time.Duration(SendDelay) * time.Second
	if msg != "" {
		p.Printf(LNG_ERRORS_FOUND_S, msg)
		p.Printf(LNG_SENDING)
		done := false
		tries := 0
		for !done {
			_, err := sendMessage(msg)
			if err == nil {
				done = true
				p.Printf(LNG_OK)
			} else {
				p.Fprintf(os.Stderr, LNG_ERR_SEND_FAIL_S_D, err, SendDelay)
				time.Sleep(duration)
				tries++
				if tries >= MaxTries {
					fail(p.Sprintf(LNG_ERR_SEND_FAIL_R_D, tries))
				}
			}
		}
	}
}

func main() {
	initLangs()
	userLanguage, _ := jibber_jabber.DetectLanguage()

        // dirty hack to switch language before cli analyser starts up
	for _, value := range os.Args {
	        value = strings.ToLower(strings.TrimSpace(value))
	        if (value == "--lang-en") || (value == "-e") ||
			((len(value) >= 2) && 
				strings.HasPrefix(value, "-") && !strings.HasPrefix(value, "--") && 
				strings.Contains(value, "e")) {
	        	userLanguage = "en"
	        }
	        if (value == "--lang-ru") || (value == "-r") ||
			((len(value) >= 2) &&
				strings.HasPrefix(value, "-") && !strings.HasPrefix(value, "--") && 
				strings.Contains(value, "r")) {
	        	userLanguage = "ru"
	        }
	}

        matcher = language.NewMatcher(message.DefaultCatalog.Languages())
	tag, _, _ := matcher.Match(language.MustParse(userLanguage))
	p = message.NewPrinter(tag)

	cli.HelpFlag = &cli.BoolFlag {
		Name:    "help",
		Aliases: []string{"h"},
		Usage:   p.Sprintf(LNG_GET_HELP),
		DisableDefaultText: true,
	}
	cli.VersionFlag = &cli.BoolFlag {
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   p.Sprintf(LNG_GET_VERSION), 
		DisableDefaultText: true,
	}
	cli.VersionPrinter = func(cCtx *cli.Context) {
		fmt.Printf(p.Sprintf(LNG_APP_VERSION, cCtx.App.Name, cCtx.App.Version))
	}

	if userLanguage == "ru" {
		cli.AppHelpTemplate = `НАЗВАНИЕ:
   {{.Name}} - {{.Usage}}

ИСПОЛЬЗОВАНИЕ:
   {{.HelpName}} опции url1 [url2 ...]
   {{if len .Authors}}
АВТОР:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
ОПЦИИ:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}{{if .Copyright }}
АВТОРСКИЕ ПРАВА:
   {{.Copyright}}
   {{end}}{{if .Version}}
ВЕРСИЯ:
   {{.Version}}
   {{end}}
`
	} else {
		cli.AppHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}

USAGE:
   {{.HelpName}} options url1 [url2 ...]
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}{{if .Copyright }}
COPYRIGHT:
   {{.Copyright}}
   {{end}}{{if .Version}}
VERSION:
   {{.Version}}
   {{end}}
`
	}

	app := &cli.App {
		Name:    "SSLChecker",
		Version: APP_VERSION,
		Usage:   p.Sprintf(LNG_PROGRAM_USAGE),
		UseShortOptionHandling: true,
		Commands: nil,
		Flags:  []cli.Flag {
			&cli.IntFlag {
				Name:  "min-days",
                                Aliases: []string{"m"},
				Value: 5,
				Usage: p.Sprintf(LNG_CERT_MIN_DAYS),
				Destination: &MinDays,
			},            
			&cli.IntFlag {
				Name:  "send-delay",
                                Aliases: []string{"d"},
				Value: 3,
				Usage: p.Sprintf(LNG_DELAY_BTW_SND_ATT),
				Destination: &SendDelay,
			},            
			&cli.IntFlag {
				Name:  "max-tries",
                                Aliases: []string{"x"},
				Value: 5,
				Usage: p.Sprintf(LNG_MAX_NUM_SND_ATT),
				Destination: &MaxTries,
			},            
			&cli.StringFlag {
				Name:  "tgm-token",
                                Aliases: []string{"t"},
				Usage: p.Sprintf(LNG_TGM_TOKEN),
				Required: true,
				Destination: &TgmToken,
			},
			&cli.StringFlag {
				Name:  "tgm-chatid",
                                Aliases: []string{"c"},
				Usage: p.Sprintf(LNG_TGM_CHATID),
                                Required: true,
                                Destination: &TgmChatId,
			},
			&cli.BoolFlag {
				Name:  "lang-en",
                                Aliases: []string{"e"},
				Usage: p.Sprintf(LNG_LANG_EN),
				DisableDefaultText: true,
			},
			&cli.BoolFlag {
				Name:  "lang-ru",
                                Aliases: []string{"r"},
				Usage: p.Sprintf(LNG_LANG_RU),
				DisableDefaultText: true,
			},
			&cli.BoolFlag {
				Name:  "verbose",
                                Aliases: []string{"v"},
				Usage: p.Sprintf(LNG_VERBOSE_MODE),
				DisableDefaultText: true,
				Destination: &VerboseMode,
			},
		},
		Action: func(cCtx *cli.Context) error {
			if cCtx.NArg() == 0 {
				fail(p.Sprintf(LNG_ERR_MISSING_URL))
			}
			run(cCtx.Args())
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fail(err.Error())
	}
}
