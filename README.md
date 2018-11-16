# nfc_runner - for Olimex MOD-RFID125

Run commands when RFID chip is nearby.

## Requirements

To use this program, you'll need a [MOD-RFID125 from
Olimex](https://www.olimex.com/Products/Modules/RFID/MOD-RFID125/). Or use any
other similar device and start forking this project.

## Usage

```shell
$ go run nfc_runner.go --help
Usage of /tmp/go-build502780512/b001/exe/nfc_runner:
  -command string
    	command to execute (default "echo %SERIAL")
  -continuous
    	exec command until detection stop
  -debug
    	output debug messages
  -file string
    	commands file with lines having 'serial;command'
  -port string
    	serial port connected to MOD-RFID125 listener (default "/dev/ttyACM0")
  (%SERIAL keyword in commands will be substitued by serial number)
```

Try a simple echo command first
```shell
$ go run nfc_runner.go
executing command "echo %SERIAL" with serial number: "FEFEFEFEFE"
FEFEFEFEFE
executing command "echo %SERIAL" with serial number: "FAFAFAFAFA"
FAFAFAFAFA
```

Then use it to open door, start music, turn light on, or anything that might be
done using shell command.
```shell
$ go run nfc_runner.go -command /usr/local/bin/launch_rocket_on_mars
```
