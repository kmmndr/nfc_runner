package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func execCommand(serialNumber <-chan string, cmd []string) {
	for {
		select {
		case msg := <-serialNumber:
			fmt.Printf("executing command %v with serial number: %v\n", cmd, msg)
			re := regexp.MustCompile("%SERIAL")
			args := strings.Join(cmd[1:], " ")
			fmt.Printf("%q\n", re.FindString(args))
			out, err := exec.Command(cmd[0], cmd[1:]...).Output()
			if err != nil {
				fmt.Println("error occured")
				fmt.Printf("%s", err)
			}
			fmt.Printf("%s", out)
		default:
			//fmt.Println("no activity")
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

			switch sig {
			case syscall.SIGUSR1:
				serialNumber <- strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
			case syscall.SIGUSR2:
				serialNumber <- strconv.FormatInt(-time.Now().UTC().UnixNano(), 10)
			}

		default:
			//fmt.Println("no signal detected")
		}
		time.Sleep(100 * time.Millisecond)
	}
}

//
//serialNumbers = make(chan string, 1)
//go fakeNfcEventHandler(serialNumber)

//

func main() {
	// ENVIRONMENT
	os.Setenv("FOO", "1")
	fmt.Println("FOO:", os.Getenv("FOO"))
	fmt.Println("BAR:", os.Getenv("BAR"))
	fmt.Println()
	//for _, e := range os.Environ() {
	//	pair := strings.Split(e, "=")
	//	fmt.Println(pair[0])
	//}

	// program options
	boolPtr := flag.Bool("verbose", false, "verbose output")
	flag.Parse()
	fmt.Println("verbose:", *boolPtr)
	fmt.Println("cmd:", flag.Args())
	fmt.Println("cmd is a", reflect.TypeOf(flag.Args()))

	// signals
	fmt.Println("PID:", os.Getpid())
	serialNumber := make(chan string, 1)

	go fakeNfcEventHandler(serialNumber)
	go execCommand(serialNumber, flag.Args())

	for {
	}

	fmt.Println("exiting")
}
