package main

import (
	"crypto/tls"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
	"github.com/cloudfoundry/jibber_jabber"
	"golang.org/x/text/message"
	"golang.org/x/text/language"
)

var MinDays *int
var SendDelay *int
var MaxTries *int
var TgmToken = ""
var TgmChatId = ""

var p *message.Printer
var matcher language.Matcher

func initLangs() {
	message.SetString(language.AmericanEnglish, "ERROR: problem getting certificate for server %s: %s\n", "ERROR: problem getting certificate for server %s: %s\n")
	message.SetString(language.Russian, "ERROR: problem getting certificate for server %s: %s\n", "ОШИБКА: невозможно получить сертификат для сервера %s: %s\n")
	message.SetString(language.AmericanEnglish, "ERROR: name mismatch for server's certificate - %s: %s\n", "ERROR: name mismatch for server's certificate - %s: %s\n")
	message.SetString(language.Russian, "ERROR: name mismatch for server's certificate - %s: %s\n", "ОШИБКА: неверное имя в сертификате сервера - %s: %s\n")
	message.SetString(language.AmericanEnglish, "ERROR: response status code is not OK (200): %d", "ERROR: response status code is not OK (200): %d")
	message.SetString(language.Russian, "ERROR: response status code is not OK (200): %d", "ОШИБКА: код статуса в ответе сервера ошибочный (не 200): %d")
	message.SetString(language.AmericanEnglish, "ERROR: message sending failed (%s)! Pausing for %d s...\n", "ERROR: message sending failed (%s)! Pausing for %d s...\n")
	message.SetString(language.Russian, "ERROR: message sending failed (%s)! Pausing for %d s...\n", "ОШИБКА: невозможно послать сообщение (%s)! Делаем паузу (%d с)...\n")
	message.SetString(language.AmericanEnglish, "Failed to send message after %d retries!\n", "Failed to send message after %d retries!\n")
	message.SetString(language.Russian, "Failed to send message after %d retries!\n", "Не удалось отправить сообщение с %d попыток!\n")
	message.SetString(language.AmericanEnglish, "ERRORS FOUND:\n%s\n", "ERRORS FOUND:\n%s\n")
	message.SetString(language.Russian, "ERRORS FOUND:\n%s\n", "НАЙДЕННЫЕ ОШИБКИ:\n%s\n")

	message.SetString(language.AmericanEnglish, "Sending...\n",  "Sending...\n")
	message.SetString(language.Russian, "Sending...\n", "Отправка...\n")
	message.SetString(language.AmericanEnglish, "OK!\n", "OK!\n")
	message.SetString(language.Russian, "OK!\n", "ОК!\n")
	message.SetString(language.AmericanEnglish, "Server: %s\n", "Server: %s\n")
	message.SetString(language.Russian, "Server: %s\n",  "Сервер: %s\n")
	message.SetString(language.AmericanEnglish, "Issuer: %s\n", "Issuer: %s\n")
	message.SetString(language.Russian, "Issuer: %s\n",  "Выдан: %s\n")
	message.SetString(language.AmericanEnglish, "Expires: %v\n", "Expires: %v\n")
	message.SetString(language.Russian, "Expires: %v\n",  "Истекает: %v\n")
	message.SetString(language.AmericanEnglish, "%d days left\n", "%d days left\n")
	message.SetString(language.Russian, "%d days left\n", "Осталось: %d дней\n")
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
		errMsg = errMsg + p.Sprintf("ERROR: problem getting certificate for server %s: %s\n", url, err.Error())
		return errMsg
	}

	err = verifyHostname(conn, url)
	if err != nil {
		errMsg = errMsg + p.Sprintf("ERROR: name mismatch for server's certificate - %s: %s\n", url, err.Error())
		return errMsg
	}

	expiry := conn.ConnectionState().PeerCertificates[0].NotAfter
	currentTime := time.Now()
        diff := expiry.Sub(currentTime)
        daysLeft := int(math.Round(diff.Hours() / 24))
        msg = msg + p.Sprintf("Server: %s\n", url)
	msg = msg + p.Sprintf("Issuer: %s\n", issuer(fmt.Sprintf("%s", conn.ConnectionState().PeerCertificates[0].Issuer)))
	msg = msg + p.Sprintf("Expires: %v\n", expiry.Format(time.RFC1123))
	msg = msg + p.Sprintf("%d days left\n", daysLeft)
	msg = msg + "=================\n"
	fmt.Printf(msg)
	if daysLeft <= *MinDays {
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
		err = errors.New(p.Sprintf("ERROR: response status code is not OK (200): %d", response.StatusCode))
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

	var urls []string

	MinDays = flag.Int("min-days", 5, "minimal remaining active days for a certificate")
	SendDelay = flag.Int("send-delay", 2, "delay between message sending attempts (in seconds)")
	MaxTries = flag.Int("max-tries", 5, "maximum number of message sending attempts")
        flag.StringVar(&TgmToken, "tgm-token", "", "REQUIRED: Telegram token")
        flag.StringVar(&TgmChatId, "tgm-chatid", "", "REQUIRED: Telegram chat id")
        flag.Parse()

        urls = flag.Args()

        argsErr := ""

        if TgmToken == "" {
        	argsErr = argsErr + "ERROR: Telegram token is required\n"
        }

        if TgmChatId == "" {
        	argsErr = argsErr + "ERROR: Telegram chat id is required\n"
        }

        if len(urls) == 0 {
        	argsErr = argsErr + "ERROR: no server name(s) given\n"
        }

        if argsErr != "" {
        	p.Printf(argsErr)
        	os.Exit(-1)
        }

	msg := ""
	for _, value := range urls {
		msg = msg + chk(value)
	}

	duration := time.Duration(*SendDelay) * time.Second
	if msg != "" {
		p.Printf("ERRORS FOUND:\n%s\n", msg)
		p.Printf("Sending...\n")
		done := false
		tries := 0
		for !done {
			_, err = sendMessage(msg)
			if err == nil {
				done = true
				p.Println("OK!")
			} else {
				p.Printf("ERROR: message sending failed (%s)! Pausing for %d s...\n", err, SendDelay)
				time.Sleep(duration)
				tries++
				if tries >= *MaxTries {
					panic(p.Sprintf("Failed to send message after %d retries!\n", tries))
				}
			}
		}
	}
}
