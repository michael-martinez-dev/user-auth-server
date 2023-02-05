package db

import (
	"context"
	"strconv"
	"fmt"
	"os"

	"github.com/go-redis/redis"
)


type RedisConnection interface {
	Close()
	DB() *redis.Client
	GetClient() *redis.Client
	GetContext() context.Context
}

type redisConn struct {
	client *redis.Client
	ctx   context.Context
}

func NewRedisConnection() RedisConnection {
	var c redisConn
	var err error
	redis_addr := fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))
	db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		panic(err)
	}
	c.client = redis.NewClient(&redis.Options{
		Addr:     redis_addr,
		Password: "",
		DB:       db,
	})
	_, err = c.client.Ping().Result()
	if err != nil {
		panic(err)
	}
	c.ctx = context.Background()
	return &c
}

func (c *redisConn) Close() {
	err := c.client.Close()
	if err != nil {
		panic(err)
	}
}

func (c *redisConn) DB() *redis.Client {
	return c.client
}

func (c *redisConn) GetClient() *redis.Client {
	return NewRedisConnection().DB()
}

func (c *redisConn) GetContext() context.Context {
	return NewRedisConnection().(*redisConn).ctx
}
