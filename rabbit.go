package main

import (
	"github.com/streadway/amqp"
	"log"
)

var (
	amqpURI      = "amqp://guest:guest@localhost:5672/"
	exchangeType = "fanout"
	key          = "test-key"
)

type RabbitRunner struct{}

func (rr *RabbitRunner) Name() string { return "RabbitMQ" }

func (rr *RabbitRunner) setupQueues(names []string) {}

func (rr *RabbitRunner) Produce(name, body string, messages int) {
	exchange := name
	connection, err := amqp.Dial(amqpURI)
	if err != nil {
		log.Printf("Dial: %s", err)
	}
	defer connection.Close()

	channel, err := connection.Channel()
	if err != nil {
		log.Printf("Channel: %s", err)
	}

	if err := channel.ExchangeDeclare(
		exchange,     // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		log.Printf("Exchange Declare: %s", err)
	}

	// We have to declare a queue to receive BEFORE we can publish anything
	// because AMQP didn't get enough attention from their dad as a teenager.
	queue, err := channel.QueueDeclare(
		name,  // name of the queue
		true,  // durable
		false, // delete when usused
		false, // exclusive
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		log.Printf("Queue Declare: %s", err)
	}

	if err = channel.QueueBind(
		queue.Name, // name of the queue
		key,        // bindingKey
		exchange,   // sourceExchange
		false,      // noWait
		nil,        // arguments
	); err != nil {
		log.Printf("Queue Bind: %s", err)
	}

	// Reliable publisher confirms require confirm.select support from the
	// connection.
	if err := channel.Confirm(false); err != nil {
		log.Printf("Channel could not be put into confirm mode: %s", err)
	}

	ack, nack := channel.NotifyConfirm(make(chan uint64, 1), make(chan uint64, 1))

	for i := 0; i < messages; i++ {
		if err = channel.Publish(
			exchange, // publish to an exchange
			key,      // routing to 0 or more queues
			false,    // mandatory
			false,    // immediate
			amqp.Publishing{
				ContentType:  "text/plain",
				Body:         []byte(body),
				DeliveryMode: amqp.Transient, // 1=non-persistent, 2=persistent
				// a bunch of application/implementation-specific fields
			},
		); err != nil {
			log.Printf("Exchange Publish: %s", err)
		}
	}

	// yes, by default, rabbit does not care whether you published or not. check here
	for i := 0; i < messages; i++ {
		select {
		case <-ack:
		case <-nack:
		}
	}
}

func (rr *RabbitRunner) Consume(name string, messages int) {
	exchange := name

	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		log.Println("Dial: %s", err)
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Println("Channel: %s", err)
	}
	defer channel.Close()

	if err = channel.ExchangeDeclare(
		exchange,     // name of the exchange
		exchangeType, // type
		true,         // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		log.Println("Exchange Declare: %s", err)
	}

	channel.Qos(messages, 0, false)

	queue, err := channel.QueueDeclare(
		name,  // name of the queue
		true,  // durable
		false, // delete when usused
		false, // exclusive
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		log.Printf("Queue Declare: %s", err)
	}

	if err = channel.QueueBind(
		queue.Name, // name of the queue
		key,        // bindingKey
		exchange,   // sourceExchange
		false,      // noWait
		nil,        // arguments
	); err != nil {
		log.Printf("Queue Bind: %s", err)
	}

	deliveries, err := channel.Consume(
		queue.Name, // name
		"",         // consumerTag,
		false,      // noAck
		false,      // exclusive
		false,      // noLocal
		false,      // noWait
		nil,        // arguments
	)
	if err != nil {
		log.Printf("Queue Consume: %s", err)
	}

	i := 0
	// why?
	for d := range deliveries {
		d.Ack(true)
		i++
		if i >= messages {
			break
		}
	}
}
