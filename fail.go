package main

import "time"

type FailAttempts struct {
	Time time.Time
	SIP string
}

type Fail struct {
	IP string
	TemporaryBannedAt time.Time
	Attempts []FailAttempts
}

func (fail *Fail) AddAttempt(sip string) {
	fail.Attempts = append(fail.Attempts, FailAttempts{
		Time: time.Now(),
		SIP: sip,
	})
}

func (fail *Fail) GetAttempts(duration time.Duration) FailAttemptsReport {
	return GetFailAttemptsReport(fail.Attempts, duration)
}

func NewFail(ip string) *Fail {
	fail := Fail{
		IP: ip,
		Attempts: make([]FailAttempts, 0),
	}
	return &fail
}