// Copyright (c) 2023. box Inc. All rights reserved.
// Author icicd
// Create Time 2023/11/25

package nobot

import (
	"time"

	jsoniter "github.com/json-iterator/go"
)

type INoticeMessage interface {
	FromJSON(data []byte) error
	ToJSON() ([]byte, error)
	Value() map[string]interface{}
	NoticeTpl(messageType int, data string) string
}

type noticeMessage struct {
	title      string        `json:"title"`
	data       string        `json:"data"`
	label      string        `json:"label"`
	labelSleep time.Duration `json:"sleep"`
}

func (m *noticeMessage) FromJSON(data []byte) error {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Unmarshal(data, m)
}

func (m *noticeMessage) ToJSON() ([]byte, error) {
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Marshal(m)
}
