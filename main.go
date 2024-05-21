package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func main() {
	url := "https://mm.sberdevices.ru"
	botToken := "hr86g8dxyfn8bfrn7mtdustm3a"
	channelID := "67t8jfxjb7gdtkbw9o3u336gqh"

	client := &http.Client{}
	var lastPostId string

	// ID последнего сообщения при старте
	lastPostId = fetchLastPostId(client, url, botToken, channelID)
	fmt.Println("Начальный lastPostId:", lastPostId)

	for {
		time.Sleep(5 * time.Second) // Пауза перед следующей проверкой

		newPostId := fetchLastPostId(client, url, botToken, channelID)
		if newPostId != "" && newPostId != lastPostId {
			fmt.Println("Обнаружен новый пост:", newPostId)
			message := fetchMessageById(client, url, botToken, channelID, newPostId)
			if sendMessage(client, url, botToken, channelID, message) {
				lastPostId = newPostId // Обновляем только после успешной отправки
				fmt.Println("Эхо-ответ отправлен на новое сообщение.")
			}
		}
	}
}

func fetchLastPostId(client *http.Client, url, botToken, channelID string) string {
	getUrl := url + "/api/v4/channels/" + channelID + "/posts?page=0&per_page=1"
	getReq, err := http.NewRequest("GET", getUrl, nil)
	if err != nil {
		panic(err)
	}
	getReq.Header.Set("Authorization", "Bearer "+botToken)

	getResp, err := client.Do(getReq)
	if err != nil {
		panic(err)
	}
	defer getResp.Body.Close()

	body, _ := ioutil.ReadAll(getResp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if order, ok := result["order"].([]interface{}); ok && len(order) > 0 {
		return order[0].(string)
	}
	return ""
}

func fetchMessageById(client *http.Client, url, botToken, channelID, postId string) string {
	getUrl := url + "/api/v4/posts/" + postId
	getReq, err := http.NewRequest("GET", getUrl, nil)
	if err != nil {
		panic(err)
	}
	getReq.Header.Set("Authorization", "Bearer "+botToken)

	getResp, err := client.Do(getReq)
	if err != nil {
		panic(err)
	}
	defer getResp.Body.Close()

	body, _ := ioutil.ReadAll(getResp.Body)
	var postDetails map[string]interface{}
	json.Unmarshal(body, &postDetails)
	return postDetails["message"].(string)
}

func sendMessage(client *http.Client, url, botToken, channelID, message string) bool {
	data := map[string]string{
		"channel_id": channelID,
		"message":    message,
	}
	payload, _ := json.Marshal(data)
	postReq, err := http.NewRequest("POST", url+"/api/v4/posts", bytes.NewBuffer(payload))
	if err != nil {
		panic(err)
	}
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Authorization", "Bearer "+botToken)

	resp, err := client.Do(postReq)
	if err != nil {
		fmt.Println("Ошибка при отправке сообщения:", err)
		return false
	}
	defer resp.Body.Close()

	return true
}
