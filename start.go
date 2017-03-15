package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"text/template"
)

var burrowConfig = `
[general]
group-blacklist=^(console-consumer-|python-kafka-consumer-).*$

[zookeeper]
{{range .ZOOKEEPER}}hostname={{.}}
{{end}}port=2181
timeout=6
lock-path=/burrow/notifier

[kafka "local"]
{{range .KAFKA}}broker={{.}}
{{end}}broker-port=9092
offsets-topic=__consumer_offsets
{{range .ZOOKEEPER}}zookeeper={{.}}
{{end}}zookeeper-path=/local
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

type servers struct {
	ZOOKEEPER []string
	KAFKA     []string
}

func main() {

	vars := servers{}

	vars.KAFKA = strings.Split(strings.Replace(os.Getenv("MQ_URL"), ":9092", "", -1), ",")
	vars.ZOOKEEPER = strings.Split(strings.Replace(os.Getenv("MQ_ZK_URL"), ":2181", "", -1), ",")

	tmpl, err := template.New("burrowConfig").Parse(burrowConfig)
	if err != nil {
		fmt.Errorf("unable to parse template: %s", burrowConfig)
		panic(err)
	}

	os.Remove("/var/tmp/burrow/burrow.pid")

	f, err := os.Create("burrow-config.ini")
	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(f, vars)
	if err != nil {
		fmt.Errorf("unable to execute template: %s", burrowConfig)
		panic(err)
	}

	binary, lookErr := exec.LookPath("/go/bin/burrow")
	if lookErr != nil {
		panic(lookErr)
	}

	args := []string{"/go/bin/burrow", "--config", "/var/tmp/burrow/burrow-config.ini"}

	execErr := syscall.Exec(binary, args, os.Environ())
	if execErr != nil {
		panic(execErr)
	}
}
