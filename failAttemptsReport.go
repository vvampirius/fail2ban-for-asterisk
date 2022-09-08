package main

import "time"

type FailAttemptsReport struct {
	Sum int
	Uniq int
	SIP map[string]int
}

func (failAttemptsReport *FailAttemptsReport) Add(sip string) {
	failAttemptsReport.Sum++
	if n, found := failAttemptsReport.SIP[sip]; found {
		failAttemptsReport.SIP[sip] = n + 1
		return
	}
	failAttemptsReport.Uniq++
	failAttemptsReport.SIP[sip] = 1
}


func GetFailAttemptsReport(attempts []FailAttempts, duration time.Duration) FailAttemptsReport {
	far := FailAttemptsReport{
		SIP: make(map[string]int),
	}
	after := time.Now().Add(duration * -1)
	for _, attempt := range attempts {
		if attempt.Time.Before(after) { continue }
		far.Add(attempt.SIP)
	}
	return far
}