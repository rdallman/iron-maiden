package main

import (
	"log"

	"github.com/iron-io/iron_go3/mq"
)

type IronRunner struct{}

func (ir *IronRunner) Name() string { return "IronMQ" }

func (ir *IronRunner) Produce(name, body string, messages int) {
	q := mq.New(name)
	msgs := make([]mq.Message, messages)
	for i := 0; i < messages; i++ {
		msgs[i] = mq.Message{Body: body}
	}
	//then := time.Now()
	_, err := q.PushMessages(msgs...)
	//go common.PostStatHatValue("travis@iron.io", "nginx-put", float64(time.Since(then)/time.Millisecond))
	if err != nil {
		log.Println(err)
	}
}

func (ir *IronRunner) Consume(name string, messages int) {
	q := mq.New(name)
	//then := time.Now()
	msgs, err := q.GetN(messages)
	//go common.PostStatHatValue("travis@iron.io", "nginx-get", float64(time.Since(then)/time.Millisecond))
	if err != nil {
		log.Println(err)
	}
	for _, msg := range msgs {
		err := msg.Delete()
		if err != nil {
			log.Println(err)
		}
	}
}
