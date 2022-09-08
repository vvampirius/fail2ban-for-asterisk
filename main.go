// ipset create asterisk_ban hash:ip
// iptables -A INPUT -m set --match-set asterisk_ban src -j DROP
// ipset add asterisk_ban 40.89.131.67

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
)

const VERSION  = `0.1`

var (
	ErrorLog = log.New(os.Stderr, `error#`, log.Lshortfile)
	DebugLog = log.New(os.Stdout, `debug#`, log.Lshortfile)
)

func helpText() {
	fmt.Println(`bla-bla-bla`)
	flag.PrintDefaults()
}

func main() {
	help := flag.Bool("h", false, "print this help")
	ver := flag.Bool("v", false, "Show version")
	flag.Parse()

	if *help {
		helpText()
		os.Exit(0)
	}

	if *ver {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	cmd := exec.Command(`journalctl`, `-u`, `asterisk`, `-f`)
	log, err := cmd.StdoutPipe()
	if err != nil {
		ErrorLog.Fatalln(err)
	}
	if err := cmd.Start(); err != nil {
		ErrorLog.Fatalln(err)
	}

	core := NewCore()
	core.ReadLog(log)
}
