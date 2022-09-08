# fail2ban-for-asterisk

This tool watch for REGISTER request with "Failed to authenticate" in 'journald' for unit 'asterisk' and ban unwanted IP with iptables + **ipset** (and provides metrics for Prometheus monitoring about it).

> **This tool is not related to [fail2ban](https://github.com/fail2ban/fail2ban) project.**

It requires Linux with iptables and <u>ipset</u>.

Ban rules:
- IP address will be banned temporary for one hour if it has failed auth only to one user for more than 20 times.
- IP address will be banned permanently if it has failed auth to >1 users for >=3 times in summary.

fail2ban-for-asterisk creates ipset list 'asterisk_ban' (can be changed with commandline parameter) if not exists and related iptabled rule.

```shell
ipset create asterisk_ban hash:ip
iptables -A INPUT -m set --match-set asterisk_ban src -j DROP
```

Further, it just adds IPs to the list for ban.

```shell
ipset add asterisk_ban <IP>
```