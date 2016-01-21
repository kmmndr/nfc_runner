package main

import "os"

// import "strings"
import "fmt"
import "flag"
import "strconv"

import "os/signal"
import "syscall"
import "time"

func execCommand(serialNumber <-chan string) {
	for {
		select {
		case msg := <-serialNumber:
			fmt.Println("executing command with serial number: ", msg)
		default:
			fmt.Println("no activity")
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
	wordPtr := flag.String("word", "foo", "a string")
	numbPtr := flag.Int("numb", 42, "an int")
	boolPtr := flag.Bool("fork", false, "a bool")
	var svar string
	flag.StringVar(&svar, "svar", "bar", "a string var")
	flag.Parse()
	fmt.Println("word:", *wordPtr)
	fmt.Println("numb:", *numbPtr)
	fmt.Println("fork:", *boolPtr)
	fmt.Println("svar:", svar)
	fmt.Println("tail:", flag.Args())

	// signals
	fmt.Println("PID:", os.Getpid())

	//sigs := make(chan os.Signal, 1)
	//done := make(chan bool, 1)
	//signal.Notify(sigs, syscall.SIGUSR1)
	//go func() {
	//	sig := <-sigs
	//	fmt.Println()
	//	fmt.Println(sig)
	//	done <- true
	//}()

	//fmt.Println("awaiting signal")
	//<-done
	serialNumber := make(chan string, 1)

	go fakeNfcEventHandler(serialNumber)
	go execCommand(serialNumber)

	for {
	}

	fmt.Println("exiting")
}
