package main

import (
	"fmt"
	"os"
	"text/template"
	"os/exec"
	"syscall"
)

var burrowConfig = `
[general]
group-blacklist=^(console-consumer-|python-kafka-consumer-).*$

[zookeeper]
hostname={{.ZOOKEEPER_NAME}}
port=2181
timeout=6
lock-path=/burrow/notifier

[kafka "local"]
broker={{.KAFKA_NAME}}
broker-port=9092
offsets-topic=__consumer_offsets
zookeeper=zookeeper
zookeeper-path=/local
zookeeper-offsets=true
offsets-topic=__consumer_offsets

[tickers]
broker-offsets=60

[lagcheck]
intervals=10
expire-group=604800

[httpserver]
server=on
port=8000
`

var env = []string{"KAFKA_NAME", "ZOOKEEPER_NAME"}


func main() {

	vars := map[string]string{}

	for _, v := range env {
		vars[v] = os.Getenv(v)
		fmt.Printf("%s=%s\n", v, os.Getenv(v))
	}

	tmpl, err := template.New("burrowConfig").Parse(burrowConfig)
	if err != nil {
		fmt.Errorf("unable to parse template: %s\n", burrowConfig)
		panic(err)
	}

	f, err := os.Create("burrow-config.ini")
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(f, vars)
	if err != nil {
		fmt.Errorf("unable to execute template: %s\n", burrowConfig)
		panic(err)
	}

	binary, lookErr := exec.LookPath("/go/bin/burrow")
	if lookErr != nil {
		panic(lookErr)
	}

	args := []string{"/go/bin/burrow", "--config", "burrow-config.ini"}

	execErr := syscall.Exec(binary, args, os.Environ())
	if execErr != nil {
		panic(execErr)
	}
}
