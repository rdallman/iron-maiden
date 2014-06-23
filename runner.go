package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type MQRunner interface {
	Name() string
	// Producer defines a runner that writes x messages with
	// body specifed on queue called name.
	Produce(name, body string, messages int)
	// Consumer defines a runner that gets x messages from
	// queue called name.
	Consume(name string, messages int)
}

func init() {
	f, err := os.Create("errorlog")
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}
	log.SetOutput(f)
}

func main() {
	var mqs []MQRunner
	mqs = append(mqs, new(IronRunner), new(RabbitRunner))

	// test single producer single queue single message
	prodThenConsume(mqs, 10000, 1, 1)

	//prodThenConsume(mqs, 100000, 100, 1)

	//prodAndConsume(mqs, 100000, 1, 1)
	//prodAndConsume(mqs, 100000, 100, 1)

	//prodThenConsume(mqs, 100000, 1, 100)
	//prodThenConsume(mqs, 100000, 100, 100)

	//prodAndConsume(mqs, 100000, 1, 100)
	//prodAndConsume(mqs, 100000, 100, 100)
}

func prodThenConsume(mqs []MQRunner, messages, atATime, queues int) {
	qnames := qnames(queues)
	for _, mq := range mqs {
		fmt.Println(mq.Name()+":", "benchmark with", messages, "message(s),",
			atATime, "at a time, across", queues, "queue(s)")

		dur := produce(mq, messages, atATime, qnames)
		fmt.Println("producer took", dur)
		dur = consume(mq, messages, atATime, qnames)
		fmt.Println("consumer took", dur)
	}
}

func prodAndConsume(mqs []MQRunner, messages, atATime, queues int) {
	qnames := qnames(queues)
	for _, mq := range mqs {
		fmt.Println(mq.Name()+":", "concurrency benchmark with", messages, "message(s),",
			atATime, "at a time, across", queues, "queue(s)")

		var wait sync.WaitGroup
		wait.Add(2)
		then := time.Now()
		go func() {
			defer wait.Done()
			produce(mq, messages, atATime, qnames)
		}()
		go func() {
			defer wait.Done()
			consume(mq, messages, atATime, qnames)
		}()
		wait.Wait()
		fmt.Println("producer and consumer took", time.Since(then))
	}
}

// for each queue specified, produce x messages y at a time
func produce(mq MQRunner, messages, atATime int, qnames []string) time.Duration {
	var wait sync.WaitGroup
	wait.Add(len(qnames))
	then := time.Now()
	for _, name := range qnames {
		go func(name string) {
			defer wait.Done()
			for i := 0; i < messages/atATime; i++ {
				mq.Produce(name, "con ipsum dolor sit amet shank ground round ribeye t-bone, biltong fatback frankfurter bresaola spare ribs cow turducken landjaeger turkey andouille swine. Ribeye pork venison ball tip pork belly leberkas doner beef beef ribs pig fatback. Filet mignon pork chop corned beef tri-tip boudin strip steak shank spare ribs pork belly ground round shankle short ribs. Tri-tip kielbasa cow tail tongue, turducken jowl doner bacon brisket venison swine. Ribeye chicken pancetta, venison biltong chuck ground round capicola swine andouille. Porchetta pastrami fatback, leberkas capicola drumstick tenderloin meatball frankfurter tail pork tri-tip.",
					atATime)
			}
		}(name)
	}
	wait.Wait()
	return time.Since(then)
}

// for each queue specified, consume x messages y at a time
func consume(mq MQRunner, messages, atATime int, qnames []string) time.Duration {
	var wait sync.WaitGroup
	wait.Add(len(qnames))
	then := time.Now()
	for _, qname := range qnames {
		go func(qname string) {
			defer wait.Done()
			for i := 0; i < messages/atATime; i++ {
				mq.Consume(qname, atATime)
			}
		}(qname)
	}
	wait.Wait()
	return time.Since(then)
}

func qnames(numQ int) []string {
	qnames := make([]string, numQ)
	for i := 0; i < numQ; i++ {
		qnames[i] = rand_str(12)
	}
	return qnames
}

func rand_str(str_size int) string {
	alphanum := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" // rabbit doesn't do unicode so hot :(
	var bytes = make([]byte, str_size)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}
