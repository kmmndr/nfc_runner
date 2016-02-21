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
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/tarm/serial"
)

var verbose bool

func readOlimexSerial(serialNumber chan<- string, port string) {
	buf := make([]byte, 128)

	c := &serial.Config{Name: port, Baud: 115200}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	defer s.Close()

	/**
	* MOD-RFID125-USBSTICK
	* Brief command list and usage description:
	*   i - print information
	*   ? - print this help
	*   b - switch to bootloader
	*   r - wait RFID tag and print it, or timeout and print zeros
	*   msXX - Set single read mode with XX (decimal) seconds timeout
	*   mcFR - Set continuous read mode with frequency F (Hz) and repeat report
	*          interval R (seconds).
	*          Note1: F and R are single decimal digits
	*          Note2: If F=0 then RF antenna is constantly switched ON.
	*          Note2: If R=0 then RF tag data is reported only once.
	*   mlE - Set led mode to disabled (E=0), or enabled (E=1)
	* See http://www.olimex.com for complete documentation.
	**/
	n, err := s.Write([]byte("mc91\r\n"))
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(100 * time.Millisecond)
	s.Read(buf)

	for {
		n, err = s.Read(buf)
		if err != nil {
			log.Fatal(err)
		}

		readString := string(buf[:n])

		if verbose {
			fmt.Printf("MOD-RFID125 read %q\n", readString)
		}

		re := regexp.MustCompile("^\r\n-([[:alnum:]]{10})\r\n>[[:space:][:print:]]*")

		serial := re.ReplaceAllString(readString, "$1")
		serialNumber <- serial

		time.Sleep(100 * time.Millisecond)
	}
}

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
	var port string
	var commandsFile string
	var command string
	var commands map[string]string
	serialNumber := make(chan string, 1)

	flag.Usage = Usage
	flag.StringVar(&port, "port", "/dev/ttyACM0", "serial port connected to MOD-RFID125 listener")
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
	go readOlimexSerial(serialNumber, port)

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
