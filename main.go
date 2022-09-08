// ipset create asterisk_ban hash:ip
// iptables -A INPUT -m set --match-set asterisk_ban src -j DROP
// ipset add asterisk_ban 40.89.131.67

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const VERSION  = `0.4`

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
	listen := flag.String("l", "127.0.0.1:8080", "Listen HTTP on address")
	ipsetName := flag.String("ipset-name", "asterisk_ban", "List name of ipset")
	flag.Parse()

	if *help {
		helpText()
		os.Exit(0)
	}

	if *ver {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	DebugLog.Printf("Started version %s", VERSION)

	firewall := NewFirewall(*ipsetName)
	if !firewall.IsIpsetListExist() {
		if err := firewall.AddIpsetList(); err != nil {
			os.Exit(1)
		}
	}
	if !firewall.IsIptablesRuleExist() {
		if err := firewall.AddIptablesRule(); err != nil {
			os.Exit(1)
		}
	}

	cmd := exec.Command(`journalctl`, `-u`, `asterisk`, `-f`, `-n`, `0`)
	log, err := cmd.StdoutPipe()
	if err != nil {
		ErrorLog.Fatalln(err)
	}
	if err := cmd.Start(); err != nil {
		ErrorLog.Fatalln(err)
	}

	metrics, err := NewMetrics(firewall.IpsetEntries)
	if err != nil {
		os.Exit(1)
	}

	http.Handle("/metrics", promhttp.Handler())

	core := NewCore(firewall, metrics)
	go core.UnbanTemporaryBannedRoutine(time.Minute, time.Hour)
	go core.ReadLog(log)

	server := http.Server{ Addr: *listen }
	if err := server.ListenAndServe(); err != nil {
		ErrorLog.Fatalln(err.Error())
	}
}
