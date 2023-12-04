// Copyright (c) 2023. box Inc. All rights reserved.
// Author icicd
// Create Time 2023/11/25

package nobot

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type noticeRobotCore struct {
	redisKey string
	redisCli *redis.Client
}

func (t *noticeRobotCore) sendNotice(message map[string]interface{}) (err error) {
	ctx := context.Background()

	addArgs := &redis.XAddArgs{
		Stream:     t.redisKey,
		NoMkStream: false,
		MaxLen:     4096,
		ID:         "*",
		Values:     message,
	}
	err = t.redisCli.XAdd(ctx, addArgs).Err()

	return
}

func (n *noticeRobotCore) close() error {
	return n.redisCli.Close()
}
