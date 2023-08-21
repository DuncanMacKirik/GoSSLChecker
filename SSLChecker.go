package main

import (
	"crypto/tls"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"time"
)

const MinDays = 7
const SendDelay = 2
const MaxTries = 5
const Token = "XXXXX:YYYYYYYYYYYYYYYYY"
const ChatId = "-ZZZZZZZ"

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

func chk(url string) string {
	msg := ""
        errMsg := ""
	conn, err := tls.Dial("tcp", url + ":443", nil)
	if err != nil {
		errMsg = errMsg + fmt.Sprintf("ERROR: problem getting certificate for server %s: %s\n", url, err.Error())
	}

	err = conn.VerifyHostname(url)
	if err != nil {
		errMsg = errMsg + fmt.Sprintf("ERROR: name mismatch for server's certificate - %s: %s\n", url, err.Error())
	}
	expiry := conn.ConnectionState().PeerCertificates[0].NotAfter
	currentTime := time.Now()
        diff := expiry.Sub(currentTime)
        daysLeft := int(math.Round(diff.Hours() / 24))
        msg = msg + fmt.Sprintf("Server: %s\n", url)
	msg = msg + fmt.Sprintf("Issuer: %s\n", issuer(fmt.Sprintf("%s", conn.ConnectionState().PeerCertificates[0].Issuer)))
	msg = msg + fmt.Sprintf("Expires: %v\n", expiry.Format(time.RFC1123))
	msg = msg + fmt.Sprintf("%d days left\n=================\n", daysLeft)
	fmt.Printf(msg)
	if daysLeft <= MinDays {
		errMsg = msg
	}
	return errMsg
}


func getUrl() string {
	return fmt.Sprintf("https://api.telegram.org/bot%s", Token)
}

func sendMessage(text string) (bool, error) {
	// Global variables
	var err error
	var response *http.Response

	// Send the message
	url := fmt.Sprintf("%s/sendMessage", getUrl())
	body, _ := json.Marshal(map[string]string{
		"chat_id": ChatId,
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

	// Close the request at the end
	defer response.Body.Close()

	// Body
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return false, err
	}

	// Return
	return true, nil
}

func main() {
	urls := []string{"server1", "server2", "server3"}

	msg := ""
	for _, value := range urls {
		msg = msg + chk(value)
	}

	if msg != "" {
		done := false
		tries := 0
		for !done {
			_, err := sendMessage(msg)
			if err == nil {
				done = true
				fmt.Printf("OK!")
			} else {
				fmt.Printf("ERROR: message sending failed! Pausing for %d s...\n", SendDelay)
                        	duration := time.Second * SendDelay
				time.Sleep(duration)
				tries++
				if tries >= MaxTries {
					panic(fmt.Sprintf("Failed to send message after %d retries!\n", tries))
				}
			}
		}
	}
}
