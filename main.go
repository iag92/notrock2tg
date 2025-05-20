package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var apiClient = &http.Client{Timeout: 5 * time.Second}

var AppConfig = struct {
	RocketApiUrl     string `json:"rocket_api_url"`
	RocketApiUser    string `json:"rocket_api_user"`
	RocketApiToken   string `json:"rocket_api_token"`
	TelegramApiToken string `json:"tg_api_token"`
	TelegramChatID   string `json:"tg_chat_id"`
	RenotifySeconds  int    `json:"renotify_seconds"`
}{}

func loadConfig() {
	raw, err := os.ReadFile("config.json")
	if err != nil {
		log.Println("Error occured while reading config")
		return
	}
	json.Unmarshal(raw, &AppConfig)
}

func getRocketData() (string, error) {
	req, err := http.NewRequest("GET", AppConfig.RocketApiUrl+"/api/v1/subscriptions.get", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", AppConfig.RocketApiUser)
	req.Header.Set("X-Auth-Token", AppConfig.RocketApiToken)
	res, err := apiClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	target := struct {
		Update []struct {
			Id                   string `json:"_id"`
			Name                 string `json:"fname"`
			UpdatedAt            string `json:"_updatedAt"`
			Alert                bool   `json:"alert"`
			DisableNotifications bool   `json:"disableNotifications"`
			Unread               int32  `json:"unread"`
		} `json:"update"`
	}{}
	json.NewDecoder(res.Body).Decode(&target)
	msg := ""
	for _, item := range target.Update {
		if item.Alert == false || item.DisableNotifications == true {
			continue
		}
		msg += "- " + item.Name + "\n"
	}
	if msg != "" {
		msg = AppConfig.RocketApiUrl + "\nNew messages in chats:\n" + msg
	}
	return msg, nil
}

func sendTgMessage(msg string, chatID string) error {
	data := struct {
		ChatID string `json:"chat_id"`
		Text   string `json:"text"`
	}{
		ChatID: chatID,
		Text:   msg,
	}
	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(data)
	req, err := http.NewRequest("POST", "https://api.telegram.org/bot"+AppConfig.TelegramApiToken+"/sendMessage", payloadBuf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := apiClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	fmt.Println("["+time.Now().Format(time.DateTime)+"] Send telegram message status =", res.Status)
	return nil
}

func main() {
	loadConfig()
	var lastAlertTime int64 = 0
	if AppConfig.RenotifySeconds == 0 {
		AppConfig.RenotifySeconds = 3600
	}
	for {
		msg, err := getRocketData()
		if err != nil {
			fmt.Println(err)
		}
		if msg != "" && (lastAlertTime == 0 || time.Now().Unix()-lastAlertTime > int64(AppConfig.RenotifySeconds)) {
			err := sendTgMessage(msg, AppConfig.TelegramChatID)
			if err != nil {
				fmt.Println(err)
				continue
			}
			lastAlertTime = time.Now().Unix()
		}
		time.Sleep(30 * time.Second)
	}

}
