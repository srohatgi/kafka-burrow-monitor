package main

import (
	"bufio"
	"fmt"
	"log"
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

	fmt.Println("MQ_URL:", os.Getenv("MQ_URL"))
	fmt.Println("MQ_ZK_URL:", os.Getenv("MQ_ZK_URL"))
	fmt.Println("KAFKA_PORT:", os.Getenv("KAFKA_PORT"))
	fmt.Println("ZOOKEEPER_PORT:", os.Getenv("ZOOKEEPER_PORT"))

	vars.KAFKA = strings.Split(strings.Replace(os.Getenv("MQ_URL"), ":"+os.Getenv("KAFKA_PORT"), "", -1), ",")
	vars.ZOOKEEPER = strings.Split(strings.Replace(os.Getenv("MQ_ZK_URL"), ":"+os.Getenv("ZOOKEEPER_PORT"), "", -1), ",")

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

	file, err := os.Open("/var/tmp/burrow/burrow-config.ini")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fmt.Println("_____Start file_____")
	fmt.Println("")

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	fmt.Println("")
	fmt.Println("_____End File_____")
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	args := []string{"/go/bin/burrow", "--config", "/var/tmp/burrow/burrow-config.ini"}

	execErr := syscall.Exec(binary, args, os.Environ())
	if execErr != nil {
		panic(execErr)
	}

	fmt.Println("_____Started successfully_____")
}
