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
	"strings"
	"time"
	"github.com/alecthomas/kingpin/v2"
	"github.com/cloudfoundry/jibber_jabber"
	"golang.org/x/text/message"
	"golang.org/x/text/language"
)

var MinDays *int
var SendDelay *time.Duration
var MaxTries *int
var TgmToken *string
var TgmChatId *string

var p *message.Printer
var matcher language.Matcher

const LNG_LANG_EN           = "force usage of English language, instead of cheking the OS defaults"
const LNG_LANG_RU           = "force usage of Russian language, instead of cheking the OS defaults"
const LNG_CERT_MIN_DAYS     = "minimal remaining active days for a certificate"
const LNG_DELAY_BTW_SND_ATT = "delay between message sending attempts (in seconds)"
const LNG_MAX_NUM_SND_ATT   = "maximum number of message sending attempts"
const LNG_TGM_TOKEN         = "Telegram token for sending messsages"
const LNG_TGM_CHATID        = "Telegram chat id for sending messsages"
const LNG_SRV_NAMES         = "server name(s) to check (separated by spaces)"
const LNG_ERR_GETCERT_SRV_S = "ERROR: problem getting certificate for server %s: %s\n"
const LNG_ERR_NAME_MISM_S   = "ERROR: name mismatch for server's certificate - %s: %s\n"
const LNG_ERR_STATUS_N200_D = "ERROR: response status code is not OK (200): %d"
const LNG_ERR_SEND_FAIL_S_D = "ERROR: message sending failed (%s)! Pausing for %d s...\n"
const LNG_ERR_SEND_FAIL_R_D = "Failed to send message after %d retries!\n"
const LNG_ERRORS_FOUND_S    = "ERRORS FOUND:\n%s\n"
const LNG_SENDING           = "Sending...\n"
const LNG_OK                = "OK!\n"
const LNG_SERVER_S          = "Server: %s\n"
const LNG_ISSUER_S          = "Issuer: %s\n"
const LNG_EXPIRES_V         = "Expires: %v\n"
const LNG_DAYSLEFT_D        = "%d days left\n"

func initLangs() {
	message.SetString(language.AmericanEnglish, LNG_LANG_EN, LNG_LANG_EN)
	message.SetString(language.Russian, LNG_LANG_EN, "использовать английский язык (вместо попытки автоопределения языка ОС)")
	message.SetString(language.AmericanEnglish, LNG_LANG_RU, LNG_LANG_RU)
	message.SetString(language.Russian, LNG_LANG_RU, "использовать русский язык (вместо попытки автоопределения языка ОС)")
        message.SetString(language.AmericanEnglish, LNG_CERT_MIN_DAYS, LNG_CERT_MIN_DAYS)
	message.SetString(language.Russian, LNG_CERT_MIN_DAYS, "минимально допустимое время истечения сертификата (в днях)")
	message.SetString(language.AmericanEnglish, LNG_DELAY_BTW_SND_ATT, LNG_DELAY_BTW_SND_ATT)
	message.SetString(language.Russian, LNG_DELAY_BTW_SND_ATT, "длительность задержки между попытками отправки сообщений в Telegram (в секундах)")
	message.SetString(language.AmericanEnglish, LNG_MAX_NUM_SND_ATT, LNG_MAX_NUM_SND_ATT)
	message.SetString(language.Russian, LNG_MAX_NUM_SND_ATT, "максимальное количество попыток отправки сообщений в Telegram")
	message.SetString(language.AmericanEnglish, LNG_TGM_TOKEN, LNG_TGM_TOKEN)
	message.SetString(language.Russian, LNG_TGM_TOKEN, "значение токена Telegram для отправки сообщений")
	message.SetString(language.AmericanEnglish, LNG_TGM_CHATID, LNG_TGM_CHATID)
	message.SetString(language.Russian, LNG_TGM_CHATID, "значение chat id Telegram для отправки сообщений")
	message.SetString(language.AmericanEnglish, LNG_SRV_NAMES, LNG_SRV_NAMES)
	message.SetString(language.Russian, LNG_SRV_NAMES, "имя (имена) серверов для проверки (через пробел)")

	message.SetString(language.AmericanEnglish, LNG_ERR_GETCERT_SRV_S, LNG_ERR_GETCERT_SRV_S)
	message.SetString(language.Russian, LNG_ERR_GETCERT_SRV_S, "ОШИБКА: невозможно получить сертификат для сервера %s: %s\n")
	message.SetString(language.AmericanEnglish, LNG_ERR_NAME_MISM_S, LNG_ERR_NAME_MISM_S)
	message.SetString(language.Russian, LNG_ERR_NAME_MISM_S, "ОШИБКА: неверное имя в сертификате сервера - %s: %s\n")
	message.SetString(language.AmericanEnglish, LNG_ERR_STATUS_N200_D, LNG_ERR_STATUS_N200_D)
	message.SetString(language.Russian, LNG_ERR_STATUS_N200_D, "ОШИБКА: код статуса в ответе сервера ошибочный (не 200): %d")
	message.SetString(language.AmericanEnglish, LNG_ERR_SEND_FAIL_S_D, LNG_ERR_SEND_FAIL_S_D)
	message.SetString(language.Russian, LNG_ERR_SEND_FAIL_S_D, "ОШИБКА: невозможно послать сообщение (%s)! Делаем паузу (%d с)...\n")
	message.SetString(language.AmericanEnglish, LNG_ERR_SEND_FAIL_R_D, LNG_ERR_SEND_FAIL_R_D)
	message.SetString(language.Russian, LNG_ERR_SEND_FAIL_R_D, "Не удалось отправить сообщение с %d попыток!\n")
	message.SetString(language.AmericanEnglish, LNG_ERRORS_FOUND_S, LNG_ERRORS_FOUND_S)
	message.SetString(language.Russian, LNG_ERRORS_FOUND_S, "НАЙДЕННЫЕ ОШИБКИ:\n%s\n")

	message.SetString(language.AmericanEnglish, LNG_SENDING,  LNG_SENDING)
	message.SetString(language.Russian, LNG_SENDING, "Отправка...\n")
	message.SetString(language.AmericanEnglish, LNG_OK, LNG_OK)
	message.SetString(language.Russian, LNG_OK, "ОК!\n")
	message.SetString(language.AmericanEnglish, LNG_SERVER_S, LNG_SERVER_S)
	message.SetString(language.Russian, LNG_SERVER_S,  "Сервер: %s\n")
	message.SetString(language.AmericanEnglish, LNG_ISSUER_S, LNG_ISSUER_S)
	message.SetString(language.Russian, LNG_ISSUER_S,  "Выдан: %s\n")
	message.SetString(language.AmericanEnglish, LNG_EXPIRES_V, LNG_EXPIRES_V)
	message.SetString(language.Russian, LNG_EXPIRES_V,  "Истекает: %v\n")
	message.SetString(language.AmericanEnglish, LNG_DAYSLEFT_D, LNG_DAYSLEFT_D)
	message.SetString(language.Russian, LNG_DAYSLEFT_D, "Осталось: %d дней\n")
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
	if daysLeft <= *MinDays {
		errMsg = msg
	}
	return errMsg
}


func getUrl() string {
	return fmt.Sprintf("https://api.telegram.org/bot%s", *TgmToken)
}

func sendMessage(text string) (bool, error) {
	var err error
	var response *http.Response

	url := fmt.Sprintf("%s/sendMessage", getUrl())
	body, _ := json.Marshal(map[string]string{
		"chat_id": *TgmChatId,
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

func main() {
	initLangs()
	userLanguage, err := jibber_jabber.DetectLanguage()
        matcher = language.NewMatcher(message.DefaultCatalog.Languages())
	tag, _, _ := matcher.Match(language.MustParse(userLanguage))
	p = message.NewPrinter(tag)

	var urls *[]string

        kingpin.Version("0.1.0")
        _ = kingpin.Flag("lang-en", p.Sprintf(LNG_LANG_EN)).Short('e').Bool()
        _ = kingpin.Flag("lang-ru", p.Sprintf(LNG_LANG_RU)).Short('r').Bool()
        MinDays = kingpin.Flag("min-days", p.Sprintf(LNG_CERT_MIN_DAYS)).Default("5").Short('m').Int()
        SendDelay = kingpin.Flag("send-delay", p.Sprintf(LNG_DELAY_BTW_SND_ATT)).Default("3s").Short('d').Duration()
        MaxTries = kingpin.Flag("max-tries", p.Sprintf(LNG_MAX_NUM_SND_ATT)).Default("5").Short('x').Int()
        TgmToken = kingpin.Flag("tgm-token", p.Sprintf(LNG_TGM_TOKEN)).Short('t').Required().String()
        TgmChatId = kingpin.Flag("tgm-chatid", p.Sprintf(LNG_TGM_CHATID)).Short('c').Required().String()
        urls = kingpin.Arg("servers", p.Sprintf(LNG_SRV_NAMES)).Required().Strings()
        kingpin.Parse()

	msg := ""
	for _, value := range *urls {
		msg = msg + chk(value)
	}

	duration := (*SendDelay) * time.Second
	if msg != "" {
		p.Printf(LNG_ERRORS_FOUND_S, msg)
		p.Printf(LNG_SENDING)
		done := false
		tries := 0
		for !done {
			_, err = sendMessage(msg)
			if err == nil {
				done = true
				p.Printf(LNG_OK)
			} else {
				p.Printf(LNG_ERR_SEND_FAIL_S_D, err, SendDelay)
				time.Sleep(duration)
				tries++
				if tries >= *MaxTries {
					panic(p.Sprintf(LNG_ERR_SEND_FAIL_R_D, tries))
				}
			}
		}
	}
}
