// Copyright (c) 2023. staking Inc. All rights reserved.
// Author icicd
// Create Time 2023/11/8

package nobot

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

const (
	MsgIdCommon = iota
	MsgIdWarn
	MsgIdFatal
)

type sendMessage struct {
	Ok     bool `json:"ok"`
	Result struct {
		MessageID int `json:"message_id"`
		From      struct {
			ID        int    `json:"id"`
			IsBot     bool   `json:"is_bot"`
			FirstName string `json:"first_name"`
			Username  string `json:"username"`
		} `json:"from"`
		Chat struct {
			ID        int    `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Username  string `json:"username"`
			Type      string `json:"type"`
		} `json:"chat"`
		Date int    `json:"date"`
		Text string `json:"text"`
	} `json:"result"`
}

type tgPayload struct {
	ChatId string `json:"chat_id"`
	Text   string `json:"text"`
}

func (m *tgPayload) ToJSON() ([]byte, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Marshal(m)
}

type Tg struct {
	token  string
	chatId string
}

func NewTg(chatId, token string) INoticeRobotMessager {
	return &Tg{
		chatId: chatId, token: token,
	}
}

func (t *Tg) format(message map[string]interface{}) string {

	msgId := 0
	if v, ok := message["msg_id"]; ok {
		msgId, _ = strconv.Atoi(v.(string))
	}

	switch msgId {
	case MsgIdCommon:
		return t.commoTpl(message)
	case MsgIdWarn:
		return t.warnTpl(message)
	case MsgIdFatal:
		return t.fatalTpl(message)
	default:
		return t.commoTpl(message)
	}

	return ""
}

func (t *Tg) commoTpl(data map[string]interface{}) string {

	tpl := `
<pre>
🔔🔔🔔 Notification ` + time.Now().Format("2006-01-02 15:04:05")

	for k, v := range data {
		tpl += "\n<b>" + k + ":</b>" + v.(string)
	}

	tpl += `
</pre>
`
	return tpl
}

func (t *Tg) warnTpl(data map[string]interface{}) string {

	tpl := `
<pre>
❗❗❗️ Warning` + time.Now().Format("2006-01-02 15:04:05")

	for k, v := range data {
		tpl += "\n<b>" + k + ":</b>" + v.(string)
	}

	tpl += `
</pre>
`
	return tpl
}

func (t *Tg) fatalTpl(data map[string]interface{}) string {

	tpl := `
<pre>
😈😈😈 Exception` + time.Now().Format("2006-01-02 15:04:05")

	for k, v := range data {
		tpl += "\n<b>" + k + ":</b>" + v.(string)
	}

	tpl += `
</pre>
`
	return tpl
}

func (t *Tg) Notice(messageList []map[string]interface{}) (err error) {

	var (
		url     string
		content []byte
	)

	mainContent := make([]string, len(messageList))

	for i, messageData := range messageList {
		mainContent[i] = t.format(messageData)
	}

	data := make(map[string]string)
	data["chat_id"] = t.chatId
	data["text"] = strings.Join(mainContent, "")
	data["parse_mode"] = "html"

	url = fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)
	content, err = httpPost(url, data)
	if err != nil {
		log.Printf("Sending message to TG failed:%v", err)
		return
	}

	var rs = &sendMessage{}

	err = json.Unmarshal(content, &rs)
	if err != nil {
		log.Println("Failed to parse Tg return result,err:%v,content:%v", err, content)
	}

	return
}
