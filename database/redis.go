package database

import (
	"fmt"
	"time"

	"github.com/ansible-semaphore/semaphore/util"
	"gopkg.in/redis.v3"
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
