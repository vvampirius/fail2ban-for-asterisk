package main

import (
	"bufio"
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"regexp"
	"sync"
	"time"
)

var (
	RegexpFailedToAuthenticate = regexp.MustCompile(`Request 'REGISTER' from '<(.*)>' failed for '(\d+\.\d+\.\d+\.\d+):\d+'.*Failed to authenticate`)
	ErrAlreadyBanned = errors.New(`Already banned`)
)

type Core struct {
	Fails []*Fail
	failsMu sync.Mutex
	Banned *Banned
	Firewall *Firewall
	Metrics *Metrics
}

func (core *Core) ReadLog(r io.Reader) {
	reader := bufio.NewReader(r)
	for {
	  line, err := reader.ReadString('\n')
	  core.ReadLine(line)
	  if err != nil {
		  ErrorLog.Fatalln(err.Error())
	  }
	}
}

func (core *Core) ReadLine(line string) {
	if line == `` { return }
	match := RegexpFailedToAuthenticate.FindStringSubmatch(line)
	if len(match) != 3 { return }
	core.AuthFailed(match[1], match[2])
}

func (core *Core) AuthFailed(sip, ip string) {
	core.Metrics.FailedToAuthenticate.Inc()
	fail := core.getFail(ip)
	fail.AddAttempt(sip)
	far := fail.GetAttempts(5 * time.Minute)
	if far.Uniq == 1 {
		if far.Sum >= 20 {
			if err := core.Firewall.BanIP(fail.IP, core.Banned); err != nil {
				if err != ErrAlreadyBanned { ErrorLog.Printf("Can't temporary ban %s (%v)", fail.IP, far.SIP) }
				return
			}
			fail.TemporaryBannedAt = time.Now()
			DebugLog.Printf("Temporary banned %s (%v)", fail.IP, far.SIP)
			core.Banned.Add(fail.IP, true)
			core.Metrics.Banned.With(prometheus.Labels{`type`: `temporary`}).Inc()
		}
	} else {
		if far.Sum >= 3 {
			if err := core.Firewall.BanIP(fail.IP, core.Banned); err != nil {
				if err != ErrAlreadyBanned {
					ErrorLog.Printf("Can't permanently ban %s", fail.IP)
				}
				return
			}
			DebugLog.Printf("Permanently banned %s", fail.IP)
			core.Banned.Add(fail.IP, false)
			core.Metrics.Banned.With(prometheus.Labels{`type`: `permanent`}).Inc()
		}
	}
}

func (core *Core) getFail(ip string) *Fail {
	core.failsMu.Lock()
	defer core.failsMu.Unlock()
	for _, fail := range core.Fails {
		if fail.IP == ip { return fail }
	}
	fail := NewFail(ip)
	core.Fails = append(core.Fails, fail)
	return fail
}

func (core *Core) UnbanTemporaryBannedRoutine(scanInterval, maxAge time.Duration) {
	for {
		time.Sleep(scanInterval)
		core.failsMu.Lock()
		for _, fail := range core.Fails {
			if fail.TemporaryBannedAt.IsZero() { continue }
			if fail.TemporaryBannedAt.Add(maxAge).Before(time.Now()) {
				if err := core.Firewall.UnbanIP(fail.IP); err != nil { continue }
				DebugLog.Printf("Unbanned temporary %s", fail.IP)
				go core.RemoveFail(fail.IP) // must be executed with 'go' otherwise will be deadlock
				core.Banned.Remove(fail.IP)
			}
		}
		core.failsMu.Unlock()
	}
}

func (core *Core) RemoveFail(ip string) {
	core.failsMu.Lock()
	defer core.failsMu.Unlock()
	newFails := make([]*Fail, 0)
	for _, fail := range core.Fails {
		if fail.IP == ip { continue }
		newFails = append(newFails, fail)
	}
	core.Fails = newFails
}


func NewCore(firewall *Firewall, metrics *Metrics) *Core {
	core := Core{
		Fails: make([]*Fail, 0),
		Banned: NewBanned(),
		Firewall: firewall,
		Metrics: metrics,
	}
	return &core
}