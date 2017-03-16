package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"text/template"
)

var burrowConfig = `
[general]
group-blacklist=^(console-consumer-|python-kafka-consumer-).*$

[zookeeper]
{{range .ZOOKEEPER}}hostname={{.}}
{{end}}port={{.ZOOKEEPER_PORT}}
timeout=120
lock-path=/burrow/notifier

[kafka "local"]
{{range .KAFKA}}broker={{.}}
{{end}}broker-port={{.KAFKA_PORT}}
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
	ZOOKEEPER      []string
	KAFKA          []string
	KAFKA_PORT     int
	ZOOKEEPER_PORT int
}

func convertToInt(port string) int {
	i, err := strconv.Atoi(port)
	if err != nil {
		panic(err)
	}
	return i
}

func main() {

	vars := servers{}

	fmt.Println("ZOOKEEPER:", os.Getenv("ZOOKEEPER"))
	fmt.Println("KAFKA:", os.Getenv("KAFKA"))
	fmt.Println("KAFKA_PORT:", os.Getenv("KAFKA_PORT"))
	fmt.Println("ZOOKEEPER_PORT:", os.Getenv("ZOOKEEPER_PORT"))

	vars.KAFKA = strings.Split(os.Getenv("KAFKA"), ",")
	vars.ZOOKEEPER = strings.Split(os.Getenv("ZOOKEEPER"), ",")
	vars.KAFKA_PORT = convertToInt(os.Getenv("KAFKA_PORT"))
	vars.ZOOKEEPER_PORT = convertToInt(os.Getenv("ZOOKEEPER_PORT"))

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

	fmt.Println("_____Reading parsed file_____")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	fmt.Println("_____Reading file done_____")
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	args := []string{"/go/bin/burrow", "--config", "/var/tmp/burrow/burrow-config.ini"}

	execErr := syscall.Exec(binary, args, os.Environ())
	if execErr != nil {
		panic(execErr)
	}
}
