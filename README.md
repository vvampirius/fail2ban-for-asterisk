# fail2ban-for-asterisk

This tool watch for REGISTER request with "Failed to authenticate" in 'journald' for unit 'asterisk' and ban unwanted IP with iptables + **ipset** (and provides metrics for Prometheus monitoring).

> **This tool is not related to [fail2ban](https://github.com/fail2ban/fail2ban) project.**

It requires Linux with iptables and <u>ipset</u>.

Ban rules:
- IP address will be banned temporary for one hour if it has failed auth only to one user for more than 20 times (in 5 minutes).
- IP address will be banned permanently if it has failed auth to >1 users for >=3 times in summary (in 5 minutes).

# Usage:

```shell
./fail2ban-for-asterisk -l 127.0.0.1:8080 -ipset-name asterisk_ban
```

fail2ban-for-asterisk creates ipset list 'asterisk_ban' on start if not exists and related iptables rule.

```shell
ipset create asterisk_ban hash:ip
iptables -A INPUT -m set --match-set asterisk_ban src -j DROP
```

Further, it just adds/removes IPs to the list for ban.

```shell
ipset add asterisk_ban <IP>
```

Metrics:

```shell
curl -s http://127.0.0.1:8080/metrics | egrep '(ipset|ban|auth)'
# HELP banned Banned counter
# TYPE banned counter
banned{type="permanent"} 250
# HELP failed_to_authenticate Failed to authenticate mentions in log
# TYPE failed_to_authenticate counter
failed_to_authenticate 1021
# HELP ipset_entries Count of entries in ipset list
# TYPE ipset_entries gauge
ipset_entries 251
```

