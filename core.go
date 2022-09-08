package main

import (
	"bufio"
	"errors"
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


func NewCore(firewall *Firewall) *Core {
	core := Core{
		Fails: make([]*Fail, 0),
		Banned: NewBanned(),
		Firewall: firewall,
	}
	return &core
}