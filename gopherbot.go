// go-telegram-bot v1.0 2016
// Author Dmitri Korchemkin
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// This struct for message from chat
// Inline mode is not support
type Updates struct {
	Ok     bool `json:"ok"`
	Result []struct {
		UpdateID int `json:"update_id"`
		Message  struct {
			MessageID int `json:"message_id"`
			From      struct {
				ID        int    `json:"id"`
				FirstName string `json:"first_name"`
				LastName  string `json:"last_name"`
				Username  string `json:"username"`
			} `json:"from"`
			Chat struct {
				ID        int64    `json:"id"`
				FirstName string `json:"first_name"`
				LastName  string `json:"last_name"`
				Username  string `json:"username"`
				Type      string `json:"type"`
			} `json:"chat"`
			Date     int    `json:"date"`
			Text     string `json:"text"`
			Entities []struct {
				Type   string `json:"type"`
				Offset int    `json:"offset"`
				Length int    `json:"length"`
			} `json:"entities"`
		} `json:"message"`
	} `json:"result"`
}

type TeleAPI struct {
	apiUrl  string
	token   string
	getMsg  string
	sendMsg string
	sendPhoto string
	offset  int
	timeout int
	limit   int
}

// Important: Long polling will work if offset param is send to bot api
// 			and it equal message json field "update_id" + 1
func (t *TeleAPI) GetUpdates() {

	url := t.apiUrl + t.token + t.getMsg +
		"?offset=" + strconv.Itoa(t.offset) +
		"&timeout=" + strconv.Itoa(t.timeout) +
		"&limit=" + strconv.Itoa(t.limit)

	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
		time.Sleep(3 * time.Second)
		t.GetUpdates()
	} else {
		defer resp.Body.Close()

		updates := new(Updates)
		json.NewDecoder(resp.Body).Decode(&updates)

		if updates.Ok {
			for _, val := range updates.Result {
				t.SendMessage(
					val.Message.Chat.ID,
					val.Message.Chat.FirstName,
					val.Message.Text,
				)
				t.offset = val.UpdateID + 1
			}

			if t.offset > 0 {
				t.GetUpdates()
			} else {
				time.Sleep(3 * time.Second)
				t.GetUpdates()
			}
		}
	}
}

func (t *TeleAPI) SendMessage(chatID int64, name string, text string) {

	send := func(botMethod string, botMsg string) {
		jsonStr := []byte(`{"chat_id": ` + strconv.FormatInt(chatID, 10) +
			`, ` + botMsg + `}`)

		req, _ := http.NewRequest(
			"POST",
			t.apiUrl+t.token+botMethod,
			bytes.NewBuffer(jsonStr),
		)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
		} //else {
			// Log sent message status
			// var sentMsgStatus map[string]interface{}
			// json.NewDecoder(resp.Body).Decode(&sentMsgStatus)
			// Handle it if you want
			// log.Println(sentMsgStatus)
		//}
		defer resp.Body.Close()
	}

	// Handle message from user
	switch strings.ToLower(text) {
	case "/start":
		send(t.sendMsg, `"text": "Hello, *` + name +
			`*. I'm *GopherBot*.\nLet's play ping-pong",` +
			`"parse_mode": "markdown"`)
		send(t.sendMsg, `"text": "ping"`)
	case "ping":
		send(t.sendMsg, `"text": "pong"`)
	case "pong":
		send(t.sendMsg, `"text": "ping"`)
	case "hi", "hello":
		send(t.sendMsg, `"text": "Hello"`)
	default:
		send(t.sendPhoto, `"photo": ` +
			`"AgADAgADqacxG2ILjA4QcjhJtLpIW8c6Sw0ABOCfmywocrQrnooBAAEC",` +
			`"caption": "WAT?"`)
	}
}

func main() {
	fmt.Println("Running...")

	var teleApi = &TeleAPI{
		apiUrl:  "https://api.telegram.org/bot",
		token:   "<YOUR TOKEN>",
		getMsg:  "/getUpdates",
		sendMsg: "/sendMessage",
		sendPhoto:  "/sendPhoto",
		offset:  0,  // initial offset
		timeout: 25, // any number, but I'm recommend you set 20-30 seconds
		limit:   1,  // get message one by one
	}

	teleApi.GetUpdates()
}
