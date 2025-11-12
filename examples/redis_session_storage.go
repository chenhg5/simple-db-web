//go:build ignore
// +build ignore

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chenhg5/simple-db-web/handlers"
	"github.com/redis/go-redis/v9"
)

// RedisSessionStorage Redis会话存储实现
// 适用于多实例部署，通过Redis共享会话数据
type RedisSessionStorage struct {
	client  *redis.Client
	ctx     context.Context
	keyPrefix string // Redis key前缀，默认为 "simpledb:session:"
}

// NewRedisSessionStorage 创建Redis会话存储
// addr: Redis地址，如 "localhost:6379"
// password: Redis密码，如果为空则不需要密码
// db: Redis数据库编号，默认为0
func NewRedisSessionStorage(addr, password string, db int) (*RedisSessionStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()
	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("连接Redis失败: %w", err)
	}

	return &RedisSessionStorage{
		client:    client,
		ctx:       ctx,
		keyPrefix: "simpledb:session:",
	}, nil
}

// Get 获取会话数据
func (r *RedisSessionStorage) Get(connectionID string) (*handlers.SessionData, error) {
	key := r.keyPrefix + connectionID
	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("会话不存在")
	}
	if err != nil {
		return nil, fmt.Errorf("获取会话失败: %w", err)
	}

	var data handlers.SessionData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, fmt.Errorf("解析会话数据失败: %w", err)
	}

	return &data, nil
}

// Set 保存会话数据
func (r *RedisSessionStorage) Set(connectionID string, data *handlers.SessionData, ttl time.Duration) error {
	key := r.keyPrefix + connectionID
	val, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("序列化会话数据失败: %w", err)
	}

	if ttl > 0 {
		if err := r.client.Set(r.ctx, key, val, ttl).Err(); err != nil {
			return fmt.Errorf("保存会话失败: %w", err)
		}
	} else {
		if err := r.client.Set(r.ctx, key, val, 0).Err(); err != nil {
			return fmt.Errorf("保存会话失败: %w", err)
		}
	}

	return nil
}

// Delete 删除会话数据
func (r *RedisSessionStorage) Delete(connectionID string) error {
	key := r.keyPrefix + connectionID
	if err := r.client.Del(r.ctx, key).Err(); err != nil {
		return fmt.Errorf("删除会话失败: %w", err)
	}
	return nil
}

// Close 关闭Redis连接
func (r *RedisSessionStorage) Close() error {
	return r.client.Close()
}

// SetKeyPrefix 设置Redis key前缀
func (r *RedisSessionStorage) SetKeyPrefix(prefix string) {
	r.keyPrefix = prefix
}

// 使用示例
func main() {
	// 创建Redis会话存储
	redisStorage, err := NewRedisSessionStorage("localhost:6379", "", 0)
	if err != nil {
		panic(err)
	}
	defer redisStorage.Close()

	// 创建服务器
	server, err := handlers.NewServer()
	if err != nil {
		panic(err)
	}

	// 设置Redis存储
	server.SetSessionStorage(redisStorage)

	// 注册路由并启动服务器
	router := handlers.NewGinRouter(nil)
	server.RegisterRoutes(router)
	router.Engine().Run(":8080")
}

