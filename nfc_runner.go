package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var verbose bool

func execCommand(serial string, command string) {
	fmt.Printf("executing command \"%v\" with serial number: \"%v\"\n", command, serial)
	cmd := strings.Split(strings.Replace(command, "%SERIAL", serial, -1), " ")
	out, err := exec.Command(cmd[0], cmd[1:]...).Output()
	if err != nil {
		fmt.Println("error occured")
		fmt.Printf("%s", err)
	}
	fmt.Printf("%s", out)
}

func execCommandFromMap(serialNumber <-chan string, commands map[string]string) {
	for {
		select {
		case serial := <-serialNumber:
			execCommand(serial, commands[serial])
		default:
			if verbose {
				fmt.Println("no activity")
			}
		}
		time.Sleep(time.Second * 2)
	}
}

func execCommandFromString(serialNumber <-chan string, command string) {
	for {
		select {
		case serial := <-serialNumber:
			execCommand(serial, command)
		default:
			if verbose {
				fmt.Println("no activity")
			}
		}
		time.Sleep(time.Second * 2)
	}
}

func fakeNfcEventHandler(serialNumber chan<- string) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1, syscall.SIGUSR2)

	for {
		select {
		case sig := <-sigs:
			fmt.Println("signal received: ", sig)

			now := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
			switch sig {
			case syscall.SIGUSR1:
				serialNumber <- now
			case syscall.SIGUSR2:
				serialNumber <- "1234"
			}

		default:
			// fmt.Println("no signal detected")
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func readCommandsFile(commandsFile string) map[string]string {
	commands := make(map[string]string)

	if verbose {
		fmt.Println("reading file:", commandsFile)
	}
	file, err := os.Open(commandsFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	r := csv.NewReader(bufio.NewReader(file))
	r.Comma = ';'
	r.Comment = '#'
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if verbose {
			fmt.Println(record)
		}
		commands[record[0]] = record[1]
	}

	return commands
}

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "  (%%SERIAL keyword in commands will be substitued by serial number)\n")
}

func main() {
	// program options
	var command string
	var commandsFile string
	var commands map[string]string
	serialNumber := make(chan string, 1)

	flag.Usage = Usage
	flag.StringVar(&command, "command", "echo %SERIAL", "command to execute")
	flag.StringVar(&commandsFile, "file", "", "commands file with lines having 'serial;command'")
	flag.BoolVar(&verbose, "verbose", false, "verbose output")

	flag.Parse()
	if verbose {
		fmt.Println("unknown arguments:", flag.Args())
	}

	// ENVIRONMENT
	os.Setenv("FOO", "1")
	if verbose {
		fmt.Println("FOO:", os.Getenv("FOO"))
		fmt.Println("BAR:", os.Getenv("BAR"))
		fmt.Println()
		for _, e := range os.Environ() {
			pair := strings.Split(e, "=")
			fmt.Println(pair[0])
		}
	}

	// signals
	if verbose {
		fmt.Println("PID:", os.Getpid())
	}

	go fakeNfcEventHandler(serialNumber)

	if len(commandsFile) > 0 {
		commands = readCommandsFile(commandsFile)
		go execCommandFromMap(serialNumber, commands)

	} else {
		go execCommandFromString(serialNumber, command)
	}

	for {
	}

	fmt.Println("exiting")
}
