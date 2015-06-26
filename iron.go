package main

import (
	"log"

	"github.com/iron-io/iron_go3/mq"
)

type IronRunner struct{}

func (ir *IronRunner) setupQueues(queues []string) {
	qs, err := mq.List()
	for _, q := range qs {
		log.Println("[INFO] deleting queues")
		err = q.Delete()
		if err != nil {
			log.Println("delete err", err)
		}
	}
	for _, q := range queues {
		_, err := mq.CreateQueue(q, mq.QueueInfo{})
		if err != nil {
			log.Println("err", err)
		}
	}
}

func (ir *IronRunner) Name() string { return "IronMQ" }

func (ir *IronRunner) Produce(name, body string, messages int) {
	q := mq.New(name)
	msgs := make([]string, messages)
	for i := 0; i < messages; i++ {
		msgs[i] = body
	}
	_, err := q.PushStrings(msgs...)
	if err != nil {
		log.Println(err)
	}
}

func (ir *IronRunner) Consume(name string, messages int) {
	q := mq.New(name)
	_, err := q.PopN(messages)
	if err != nil {
		log.Println(err)
	}
}
