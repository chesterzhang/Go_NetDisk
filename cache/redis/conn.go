package redis

import (
	"fmt"
	"time"

	//"github.com/garyburd/redigo/redis"
	"github.com/gomodule/redigo/redis"

)

var (
	pool      *redis.Pool
	redisHost = "localhost:6379" // host
	redisPass = "testupload" // 登录密码
)

// newRedisPool : 创建redis连接池 并 返回连接池
func newRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50, // 最大可用连接
		MaxActive:   30, // 同时可用连接数
		IdleTimeout: 300 * time.Second, //连接超时, 被回收
		// 创建连接, 并返回连接
		Dial: func() (redis.Conn, error) {
			// 1. 打开连接
			//c, err := redis.Dial("tcp", redisHost)
			c, err := redis.Dial("tcp", "localhost:6379")
			if err != nil {
				fmt.Println(err)
				return nil, err
			}

			// 2. 访问认证
			if _, err = c.Do("AUTH", redisPass); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
		// 每隔1分钟 检测一下是否连接池断线
		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := conn.Do("PING")
			return err
		},
	}
}

func init() {
	pool = newRedisPool()
}

func RedisPool() *redis.Pool {

	return pool
}
