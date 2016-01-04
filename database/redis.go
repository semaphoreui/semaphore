package database

import (
	"fmt"
	"github.com/castawaylabs/semaphore/util"
	"gopkg.in/redis.v3"
	"time"
)

// Redis pool
var Redis *redis.Client

func init() {
	Redis = redis.NewClient(&redis.Options{
		MaxRetries:  2,
		DialTimeout: 10 * time.Second,
		Addr:        util.Config.SessionDb,
	})
}

func RedisPing() {
	if _, err := Redis.Ping().Result(); err != nil {
		fmt.Println("PING to redis unsuccessful")
		panic(err)
	}
}
