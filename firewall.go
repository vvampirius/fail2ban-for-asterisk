package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
)

var RegexpIpsetEntries = regexp.MustCompile(`Number of entries: (\d+)`)

type Firewall struct {
	IpsetName string
}

func (firewall *Firewall) IsIpsetListExist() bool {
	cmd := exec.Command(`ipset`, `list`, firewall.IpsetName)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func (firewall *Firewall) AddIpsetList() error {
	cmd := exec.Command(`ipset`, `create`, firewall.IpsetName, `hash:ip`)
	if err := cmd.Run(); err != nil {
		ErrorLog.Println(err)
		return err
	}
	DebugLog.Printf("Added ipset list '%s'", firewall.IpsetName)
	return nil
}

func (firewall *Firewall) IsIptablesRuleExist() bool {
	cmd := exec.Command(`iptables`, `-S`, `INPUT`)
	response, err := cmd.Output()
	if err != nil {
		ErrorLog.Println(err)
		return false
	}
	findString := fmt.Sprintf("-A INPUT -m set --match-set %s src -j DROP\n", firewall.IpsetName)
	buffer := bytes.NewBuffer(response)
	reader := bufio.NewReader(buffer)
	for {
		line, err := reader.ReadString('\n')
		if line == findString {
			return true
		}
		if err != nil {
			if err != io.EOF {
				ErrorLog.Println(err)
			}
			break
		}
	}
	return false
}

func (firewall *Firewall) AddIptablesRule() error {
	cmd := exec.Command(`iptables`, `-A`, `INPUT`, `-m`, `set`, `--match-set`, firewall.IpsetName, `src`, `-j`, `DROP`)
	if err := cmd.Run(); err != nil {
		ErrorLog.Println(err)
		return err
	}
	DebugLog.Printf("Added iptables rule '%s'", firewall.IpsetName)
	return nil
}

func (firewall *Firewall) BanIP(ip string, banned *Banned) error {
	if bannedIP := banned.Get(ip); bannedIP != nil {
		return ErrAlreadyBanned
	}
	cmd := exec.Command(`ipset`, `add`, firewall.IpsetName, ip)
	if err := cmd.Run(); err != nil {
		ErrorLog.Println(ip, err.Error())
		return err
	}
	return nil
}

func (firewall *Firewall) UnbanIP(ip string) error {
	cmd := exec.Command(`ipset`, `del`, firewall.IpsetName, ip)
	if err := cmd.Run(); err != nil {
		ErrorLog.Println(ip, err.Error())
		return err
	}
	return nil
}

func (firewall *Firewall) IpsetEntries() float64 {
	cmd := exec.Command(`ipset`, `list`, firewall.IpsetName)
	response, err := cmd.Output()
	if err != nil {
		ErrorLog.Println(err)
		return 0
	}
	buffer := bytes.NewBuffer(response)
	reader := bufio.NewReader(buffer)
	for {
		line, err := reader.ReadString('\n')
		match := RegexpIpsetEntries.FindStringSubmatch(line)
		if len(match) == 2 {
			n, err := strconv.ParseFloat(match[1], 32)
			if err != nil {
				ErrorLog.Println(err)
				return 0
			}
			return n
		}
		if err != nil {
			if err != io.EOF {
				ErrorLog.Println(err)
			}
			break
		}
	}
	return 0
}


func NewFirewall(ipsetName string) *Firewall {
	firewall := Firewall{
		IpsetName: ipsetName,
	}
	return &firewall
}