package mq

import (
	"Go_NetDisk/config"
	"log"

	"github.com/streadway/amqp" // Advanced Message Queuing Protocol
)

var conn *amqp.Connection // rabbitmq 连接
var channel *amqp.Channel // 消息发布 channel


// 初始化 channel
func initChannel() bool {
	// 1. 判断消息发布的 channel 是否已经被创建好, 如果已经创建好了, 直接return true
	if channel != nil {
		return true
	}

	// 2. 获取一个 rabbitmq 连接
	conn, err := amqp.Dial(config.RabbitURL)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// 3. 通过rabbitmq 连接获取一个 channel, 用于发布消息
	channel, err = conn.Channel()
	if err != nil {
		log.Println(err.Error())
		return false
	}
	return true
}


// Publish 方法 向exchange发布消息,发送成功返回 true, 发送失败返回 false
func Publish(exchange string, routingKey string, msg []byte) bool {
	// 1. 创建 *amqp.Channel
	if !initChannel() {
		return false
	}

	// 2. 通过 amqp.Channel.Publish 发布消息
	err := channel.Publish(
		exchange,// "uploadserver.trans"
		routingKey, // "oss"
		false, // 如果没有对应的queue, 就会丢弃这条消息
		false, // 被废弃的参数, 不在起作用
		amqp.Publishing{
			ContentType: "text/plain", //明文格式
			Body:  msg}) // 消息本体, 见 define.go
	if err!=nil {
		log.Printf(err.Error())
		return  false
	}
	return  true

}




