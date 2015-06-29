
### Info

Both RabbitMQ and IronMQ were run on a AWS m3.2xlarge box with specs:

8 vCPU clocked at 2.5GHz, 30 GB RAM, 2 x 80 GB SSD storage

Both servers databases were cleared before each benchmark.

Tests were run on an AWS m1.small box in the same data center.

Tests were all run using the code in this repository, see "runner.go" for a good
idea on how each was run.

RabbitMQ messages were left in transient mode for the tests released, however,
performance degrades significantly when persistence is turned on. IronMQ
defaults to persistent messages (with no transient option) and still outperforms RabbitMQ
in transient mode. One test below shows the performance of RabbitMQ with
persistence turned on, but we tried to play nice.

Each message body was a 639 character phrase (each was the same 639 chars).

### Configs

To run the IronMQ tests a valid iron.json file will be needed.
For the RabbitMQ tests, the url was hot configured on the server which it was
run on. Yes, I feel bad.

### Results (run separately)

```
IronMQ: benchmark with 10000 message(s), 1 at a time, across 100 queue(s)
producer took 9m17.173697857s
consumer took 10m36.681574561s
RabbitMQ: benchmark with 10000 message(s), 1 at a time, across 100 queue(s)
producer took 36m4.893591056s
consumer took 44m8.518818323s
```

```
IronMQ: benchmark with 100000 message(s), 100 at a time, across 1 queue(s)
producer took 4.989562673s
consumer took 19.300769341s
RabbitMQ: benchmark with 100000 message(s), 100 at a time, across 1 queue(s)
producer took 18.580362755s
consumer took 26.615372466s
```

```
IronMQ: concurrency benchmark with 100000 message(s), 100 at a time, across 1
queue(s)
producer and consumer took 18.015285655s
RabbitMQ: concurrency benchmark with 100000 message(s), 100 at a time, across 1
queue(s)
producer and consumer took 40.146439718s
```

```
IronMQ: benchmark with 10000 message(s), 1 at a time, across 1 queue(s)
producer took 8.983370769s
consumer took 17.659504454s
RabbitMQ: benchmark with 10000 message(s), 1 at a time, across 1 queue(s)
producer took 1m1.558754747s
consumer took 1m6.492120839s
```

WITH RabbitMQ persistence turned on:

```
IronMQ: benchmark with 10000 message(s), 1 at a time, across 1 queue(s)
producer took 5.858436882s
consumer took 26.84556203s

RabbitMQ: benchmark with 10000 message(s), 1 at a time, across 1 queue(s)
producer took 2m18.349348198s
consumer took 31.769851235s
```

## InfluxDB

All times written to influx are in milliseconds. Everytime a test is run, it creates a new series called "{consumer/produce}-{mq.Name}-{messages}-{atATime}-{nQueues}-{payloadSize}"
