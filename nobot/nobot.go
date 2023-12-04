// Copyright (c) 2023. staking Inc. All rights reserved.
// Author icicd
// Create Time 2023/11/8

package nobot

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	defaultMessageSleep         = time.Second * 15
	defaultMessageChanSize      = 2048
	defaultStreamKeyPrefix      = "notice:robot:"
	defaultMessageLabelField    = "nobot_label"
	defaultMessageLabelExpField = "nobot_expiration"
)

type INoticeRobotMessager interface {
	Notice(message []map[string]interface{}) error
}

type NoticeRobot struct {
	ctx                  context.Context
	name                 string                      // redis stream key
	core                 *noticeRobotCore            // send message to redis
	messager             INoticeRobotMessager        // send message to tg chat group
	messageQ             chan map[string]interface{} // message queue
	messageIntervalTime  time.Duration               // interval for reading messages from Redis
	zeroMessageSleepTime time.Duration               // sleep time for no messages read from Redis
	running              bool                        // running
	labelExpirationMap   *sync.Map                   // label life time
}

func NewNoticeRobot(
	ctx context.Context,
	redisCli *redis.Client,
	name string,
	messager INoticeRobotMessager,
	messageIntervalTime time.Duration,
	zeroMessageSleepTime time.Duration,
	running bool,
) *NoticeRobot {
	robot := &NoticeRobot{
		ctx:  ctx,
		name: name,
		core: &noticeRobotCore{
			redisKey: defaultStreamKeyPrefix + name,
			redisCli: redisCli,
		},
		messageQ:             make(chan map[string]interface{}, defaultMessageChanSize),
		messager:             messager,
		messageIntervalTime:  messageIntervalTime,
		zeroMessageSleepTime: zeroMessageSleepTime,
		running:              running,
		labelExpirationMap:   &sync.Map{},
	}

	if running {
		robot.Run()
	}

	return robot
}

// send message to robot
// labels[0] Message label, controlling the frequency of incoming system messages
// labels[1] The interval time between messages entering the system
func (n *NoticeRobot) SendNotice(message map[string]interface{}, labels ...string) {

	if !n.running {
		return
	}

	var (
		label      string
		expiration = defaultMessageSleep
	)

	labelsSize := len(labels)

	if labelsSize > 0 {
		label = strings.ReplaceAll(labels[0], " ", "")

		message[defaultMessageLabelField] = label

		if labelsSize > 1 {

			var err error
			expiration, err = time.ParseDuration(labels[1])
			if err != nil {
				expiration = defaultMessageSleep
			}

			message[defaultMessageLabelExpField] = time.Now().Add(expiration)

			if labelexp, ok := n.labelExpirationMap.Load(label); ok {
				if time.Now().Before(labelexp.(time.Time)) {
					// 已经存在类似同标签的消息，丢弃信息
					return
				}
			}

			n.labelExpirationMap.Store(label, time.Now().Add(expiration))
		}
	}

	message["ctime"] = time.Now()

	select {
	case n.messageQ <- message:
	case <-time.After(1 * time.Second):
		fmt.Println("send robot notice timeout")
	}
}

// send message to tg chat group
func (n *NoticeRobot) Notice() {
	ctx := context.Background()
	key := n.core.redisKey
	groupName := "G1"
	groupExist := false

	groupInfo, err := n.core.redisCli.XInfoGroups(ctx, key).Result()
	for _, g := range groupInfo {
		if g.Name == groupName {
			groupExist = true
		}
	}

	if !groupExist {
		_, err = n.core.redisCli.XGroupCreateMkStream(ctx, key, groupName, "$").Result()
		if err != nil {
			fmt.Println(err)
		}
	}

	args := &redis.XReadGroupArgs{
		Group:    groupName,
		Consumer: "C1",
		Streams:  []string{key, ">"},
		Count:    3,
		Block:    0,
		NoAck:    true,
	}

	msgId := make([]string, 0)

	for {
		xstream, _ := n.core.redisCli.XReadGroup(ctx, args).Result()
		messages := make([]map[string]interface{}, 0)

		for _, stream := range xstream {
			for _, xmessage := range stream.Messages {
				msgId = append(msgId, xmessage.ID)
				messages = append(messages, xmessage.Values)
			}
		}

		if len(messages) == 0 {
			time.Sleep(n.zeroMessageSleepTime)
		} else {
			n.messager.Notice(messages)
			time.Sleep(n.messageIntervalTime)
		}
	}

}

func (n *NoticeRobot) clearLabel() {
	for {
		expTime := time.Now().Add(-24 * time.Hour)
		n.labelExpirationMap.Range(func(key, value any) bool {
			if value.(time.Time).Before(expTime) {
				n.labelExpirationMap.Delete(key)
			}
			return true
		})

		time.Sleep(1 * time.Hour)
	}
}

func (n *NoticeRobot) Run() {

	go func() {
		for {
			select {
			case value := <-n.messageQ:
				n.core.sendNotice(value)
			case <-n.ctx.Done():
				close(n.messageQ)
				return
			}
		}
	}()

	go n.Notice()

	go n.clearLabel()
}

func (n *NoticeRobot) Close() {
	n.core.close()
	close(n.messageQ)
}
