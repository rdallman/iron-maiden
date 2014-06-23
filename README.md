
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
in transient mode. 

### Configs

To run the IronMQ tests a valid iron.json file will be needed. 
For the RabbitMQ tests, the url was hot configured on the server which it was
run on. Yes, I feel bad. 
