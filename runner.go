package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/rcrowley/go-metrics"
)

var baseName string //This serves as the base for influx
var args []int

type MQRunner interface {
	Name() string
	// Producer defines a runner that writes x messages with
	// body specifed on queue called name.
	Produce(name, body string, messages int)
	// Consumer defines a runner that gets x messages from
	// queue called name.
	Consume(name string, messages int)

	setupQueues(queues []string)
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var config Config

	f, err := os.Create("errorlog")
	if err != nil {
		log.Println(err)
		os.Exit(2)
	}
	log.SetOutput(f)

	cfgFile, err := os.Open("influx.json")
	if err != nil {
		fmt.Fprintln(os.Stdout, "error: ", err)
		return
	}
	err = json.NewDecoder(cfgFile).Decode(&config)
	if err != nil {
		return
	}

	go Influxdb(metrics.DefaultRegistry, 250*time.Millisecond, &config)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var mqs []MQRunner
	//mqs = append(mqs, new(IronRunner), new(RabbitRunner))
	mqs = append(mqs, new(RabbitRunner))

	if len(os.Args) < 6 {
		fmt.Println("usage: ./iron-maiden [message_count] [messages_per_batch] [threads_per_queue] [amount_of_queues] [message_size]")
		return
	}
	for _, v := range os.Args[1:] {
		i, err := strconv.Atoi(v)
		if err != nil {
			log.Fatalf("couldnt parse string", v)
		}
		args = append(args, i)
	}

	prodAndConsume(mqs, args[0], args[1], args[2], args[3], args[4])
	metrics.WriteOnce(metrics.DefaultRegistry, os.Stdout)
}

func prodThenConsume(mqs []MQRunner, messages, atATime, threadperQ, queues, bytes int) {
	baseName = fmt.Sprintf("%d_%d_%d_%d_ptc", args[0], args[1], args[3], args[4])
	qnames := qnames(queues)
	for _, mq := range mqs {
		fmt.Println(mq.Name()+":", "concurrency benchmark with", messages, "message(s),",
			atATime, "at a time, across", queues, "queue(s) using ", bytes, "bytes")

		mq.setupQueues(qnames)
		dur := produce(mq, messages, atATime, threadperQ, qnames, bytes)
		fmt.Println("producer took", dur)
		dur = consume(mq, messages, atATime, threadperQ, qnames)
		fmt.Println("consumer took", dur)
	}
}

func prodAndConsume(mqs []MQRunner, messages, atATime, threadperQ, queues, bytes int) {
	baseName = fmt.Sprintf("%d_%d_%d_%d", args[0], args[1], args[3], args[4])
	qnames := qnames(queues)
	for _, mq := range mqs {
		fmt.Println(mq.Name()+":", "concurrency benchmark with", messages, "message(s),",
			atATime, "at a time, across", queues, "queue(s) using ", bytes, "bytes")

		mq.setupQueues(qnames)
		var wait sync.WaitGroup
		wait.Add(2)
		then := time.Now()
		go func() {
			defer wait.Done()
			produce(mq, messages, atATime, threadperQ, qnames, bytes)
		}()
		go func() {
			defer wait.Done()
			consume(mq, messages, atATime, threadperQ, qnames)
		}()
		wait.Wait()
		fmt.Println("producer and consumer took", time.Since(then))
	}
}

// for each queue specified, produce x messages y at a time
func produce(mq MQRunner, messages, atATime, threadperQ int, qnames []string, bytes int) time.Duration {

	produceTimerName := fmt.Sprintf("producer_%s_%s", mq.Name(), baseName)
	timer := metrics.GetOrRegisterTimer(produceTimerName, metrics.DefaultRegistry)
	payload := rand_str(bytes)

	var wait sync.WaitGroup
	wait.Add(len(qnames))
	then := time.Now()
	for _, name := range qnames {
		go func(name string) {
			defer wait.Done()
			var waiter sync.WaitGroup
			waiter.Add(threadperQ)
			for i := 0; i < threadperQ; i++ {
				go func() {
					for j := 0; j < messages/atATime/threadperQ; j++ {
						timer.Time(func() {
							mq.Produce(name, payload, atATime)
						})
					}
					waiter.Done()
				}()
			}
			waiter.Wait()
		}(name)
	}
	wait.Wait()
	return time.Since(then)
}

// for each queue specified, consume x messages y at a time
func consume(mq MQRunner, messages, atATime, threadperQ int, qnames []string) time.Duration {
	consumeTimerName := fmt.Sprintf("consumer_%s_%s", mq.Name(), baseName)
	timer := metrics.GetOrRegisterTimer(consumeTimerName, metrics.DefaultRegistry)
	var wait sync.WaitGroup
	wait.Add(len(qnames))
	then := time.Now()
	for _, name := range qnames {
		go func(name string) {
			defer wait.Done()
			var waiter sync.WaitGroup
			waiter.Add(threadperQ)
			for i := 0; i < threadperQ; i++ {
				go func() {
					for j := 0; j < messages/atATime/threadperQ; j++ {
						timer.Time(func() {
							mq.Consume(name, atATime)
						})
					}
					waiter.Done()
				}()
			}
			waiter.Wait()
		}(name)
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
