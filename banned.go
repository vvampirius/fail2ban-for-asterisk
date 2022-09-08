package main

import "time"

type BannedIP struct {
	IP string
	Time time.Time
	Temporary bool
}

type Banned struct {
	IPs []BannedIP
}

func (banned *Banned) Get(ip string) *BannedIP {
	for i, bannedIP := range banned.IPs {
		if bannedIP.IP == ip {
			return &banned.IPs[i]
		}
	}
	return nil
}

func (banned *Banned) Add(ip string, temporary bool) {
	bannedIP := BannedIP{
		IP: ip,
		Temporary: temporary,
		Time: time.Now(),
	}
	banned.IPs = append(banned.IPs, bannedIP)
}


func NewBanned() *Banned {
	banned := Banned{
		IPs: make([]BannedIP, 0),
	}
	return &banned
}