package mq

import "log"

var done chan bool // 一个用于控制

// StartConsume : 从amqp.Channel接收消息, 开始消费(由 callback 写入 OSS)
func StartConsume(qName string, cName string, callback func(msg []byte) bool) {
	// 1. 创建 *amqp.Channel
	if !initChannel() {
		panic("cannot init amqp.Channel")
	}

	// 消费者获得消息发送的channel,
	msgChannel, err := channel.Consume(
		qName, // "uploadserver.trans.oss"
		cName, // "transfer_oss"
		true,  //自动应答,收到自动回复ACK
		false, // 非唯一的消费者
		false, // rabbitMQ只能设置为false
		false, // noWait, false表示会阻塞直到有消息过来, true表示无法消费时, 立即关闭channel
		nil)
	if err != nil {
		log.Fatal(err)
		return
	}

	done = make(chan bool)

	// 循环读取channel的数据, 交给 callback function 去处理(存入OSS,更新文件信息)
	go func() {
		for c := range msgChannel {
			processErr := callback(c.Body)
			if processErr {
				// TODO: 将任务写入错误队列，待后续处理
			}
		}
	}()

	// 接收done的信号, 没有信息过来则会一直阻塞在这路,避免该函数退出
	<-done

	// 关闭通道
	channel.Close()
}

// StopConsume , 将 true 送入 done channel, 使得 StartConsume 退出, 停止消费
func StopConsume() {
	done <- true
}