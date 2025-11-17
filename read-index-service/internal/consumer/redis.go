package consumer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mattermost/mattermost-read-index-service/internal/index"
)

type RedisConsumer struct {
	client       *redis.Client
	indexService *index.Service
	streamName   string
	groupName    string
	consumerName string
}

func NewRedisConsumer(redisURL string, indexService *index.Service) *RedisConsumer {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}

	client := redis.NewClient(opt)

	return &RedisConsumer{
		client:       client,
		indexService: indexService,
		streamName:   "read_cursor_events",
		groupName:    "read-index-service",
		consumerName: "consumer-1",
	}
}

func (c *RedisConsumer) Start(ctx context.Context) error {
	// 测试连接
	if err := c.client.Ping(ctx).Err(); err != nil {
		log.Printf("Redis connection failed: %v", err)
		return err
	}

	log.Println("Connected to Redis")

	// 创建消费者组（如果不存在）
	c.client.XGroupCreateMkStream(ctx, c.streamName, c.groupName, "0")

	log.Printf("Starting Redis consumer: stream=%s, group=%s", c.streamName, c.groupName)

	for {
		select {
		case <-ctx.Done():
			log.Println("Redis consumer stopped")
			return nil
		default:
			if err := c.consumeBatch(ctx); err != nil {
				log.Printf("Error consuming batch: %v", err)
				time.Sleep(time.Second) // 错误后等待
			}
		}
	}
}

func (c *RedisConsumer) consumeBatch(ctx context.Context) error {
	streams, err := c.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    c.groupName,
		Consumer: c.consumerName,
		Streams:  []string{c.streamName, ">"},
		Count:    100,
		Block:    5 * time.Second,
	}).Result()

	if err != nil {
		if err == redis.Nil {
			return nil // 没有新消息
		}
		return err
	}

	for _, stream := range streams {
		for _, message := range stream.Messages {
			if err := c.processMessage(ctx, message); err != nil {
				log.Printf("Error processing message %s: %v", message.ID, err)
				// 继续处理其他消息
			} else {
				// ACK 消息
				c.client.XAck(ctx, c.streamName, c.groupName, message.ID)
			}
		}
	}

	return nil
}

func (c *RedisConsumer) processMessage(ctx context.Context, msg redis.XMessage) error {
	data, ok := msg.Values["data"].(string)
	if !ok {
		return nil // 跳过无效消息
	}

	var event index.ReadCursorEvent
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		return err
	}

	// 处理事件
	if err := c.indexService.HandleEvent(&event); err != nil {
		return err
	}

	log.Printf("Processed event: channel=%s, user=%s, seq=%d", 
		event.ChannelID, event.UserID, event.NewLastSeq)

	return nil
}
