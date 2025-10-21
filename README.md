# About mlogtail
The main purpose of the program is monitoring of mail service (MTA) by reading new data appearing in log file and counting the values of some parameters characterizing operation of a mail server. Currently only Postfix logs are supported.

The program has two main usage modes. In the first case (`tail` command), the program reads new data from the log file in background and maintains several counters.

`mlogtail` monitors state of the log file it woking with, so there is no needs to do anything at time of normal logs rotation.

In the second mode, `mlogtail` is used to call to a log reading process and get (and/or reset) current values of the counters.

## New in version 1.2.0

✨ **HTTP API with JSON responses** - statistics are now available via REST API for easy integration with modern monitoring systems (Prometheus, Grafana, etc.)

## Usage

```none
# mlogtail -h
Usage:
  mlogtail [OPTIONS] tail
  mlogtail [OPTIONS] "stats | stats_reset | reset"
  mlogtail [OPTIONS] <COUNTER_NAME>
  mlogtail -f <LOG_FILE_NAME>

Options:
  -f string
        Mail log file path, if the path is "-" then read from STDIN (default "/var/log/mail.log")
  -h    Show this help
  -http string
        HTTP server address (e.g., :8080 or 0.0.0.0:8080) to serve stats as JSON
  -init-from-file
        Read entire log file on startup to initialize counters, then continue tailing
  -l string
        Log reader process is listening for commands on a socket file, or IPv4:PORT,
        or [IPv6]:PORT (default "unix:/var/run/mlogtail.sock")
  -o string
        Set a socket OWNER[:GROUP] while listening on a socket file
  -p int
        Set a socket access permissions while listening on a socket file (default 666)
  -t string
        Mail log type. It is "postfix" only allowed for now (default "postfix")
  -v    Show version information and exit
```

### Log tailing mode

Unfortunately, in Go, the process has no good ways to become a daemon, so we launch the "reader" just in background:

```none
# mlogtail tail &
```

or by `systemctl`. If a log reading process have to listen to a socket then it is required to specify a netwoirking type. For example: `unix:/tmp/some.sock`.

### HTTP API mode

To access statistics via HTTP API (JSON format):

```bash
# HTTP API only (recommended: use TCP socket instead of Unix)
# HTTP API + specific log file
mlogtail -f /var/log/mail.log -http :8080 -l 127.0.0.1:3333 tail

# HTTP API + initialize counters from entire log file
mlogtail -f /var/log/mail.log -http :8080 -l 127.0.0.1:3333 -init-from-file tail

# HTTP API on all interfaces
mlogtail -f /var/log/mail.log -http 0.0.0.0:8080 -l 127.0.0.1:3333 tail

# HTTP API + Unix socket (for Zabbix, requires permissions for /var/run)
mlogtail -f /var/log/mail.log -http :8080 -l unix:/var/run/mlogtail.sock tail

# HTTP API + alternative Unix socket in /tmp
mlogtail -f /var/log/mail.log -http :8080 -l unix:/tmp/mlogtail.sock tail
```

> **⚠️ Important:** By default, the program tries to create Unix socket `/var/run/mlogtail.sock`. 
> If you get "address already in use" or "permission denied" error, use TCP socket: `-l 127.0.0.1:3333`

#### HTTP endpoints:

```bash
# Health check
curl http://localhost:8080/health
# {"status":"ok","version":"1.2.0"}

# Get all statistics
curl http://localhost:8080/stats
# {"bytes_received":1059498852,"bytes_delivered":1039967394,"received":2733,...,"queue_size":42}

# Get specific counter
curl http://localhost:8080/counter/received
# {"counter":"received","value":2733}

# Reset counters (POST)
curl -X POST http://localhost:8080/reset
# {"status":"ok","message":"Counters reset"}

# Get stats and reset (POST)
curl -X POST http://localhost:8080/stats_reset
# {"bytes_received":1234,"bytes_delivered":5678,...,"queue_size":15}

# Note: queue_size shows current Postfix queue size (mailq)

### ⚡ Flag -init-from-file

**Problem:** When mlogtail restarts, all counters reset to 0, causing "jumps" in monitoring.

**Solution:** Use the `-init-from-file` flag:

```bash
# First read entire file, then monitor new entries
mlogtail -f /var/log/mail.log -http :8080 -init-from-file tail
```

**How it works:**
1. On startup, reads the **entire** log file and counts statistics  
2. Then switches to `tail` mode and tracks only new entries
3. Counters already contain history, no jumps on restart

**Perfect for:**
- ✅ Monitoring systems (Zabbix, Prometheus)
- ✅ Servers with large log files  
- ✅ Frequent restart scenarios

### 🔄 Automatic Reset on Log Rotation

For synchronization with Postfix log rotation, you can set up automatic counter reset daily at 00:00:

```bash
# Install systemd timer
sudo cp mlogtail-reset.timer mlogtail-reset.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable mlogtail-reset.timer
sudo systemctl start mlogtail-reset.timer
```

Details: [LOGROTATION.md](LOGROTATION.md)
```

### Counters' values

```none
# mlogtail stats
bytes-received  1059498852
bytes-delivered 1039967394
received        2733
delivered       2944
forwarded       4
deferred        121
bounced         105
rejected        4
held            0
discarded       0
```

It should be noted that if the "reader" is started with the `-l` option, setting the socket or IP address and port on which the process is listening for requests, then the same command line parameters should be used for getting counter values.

Probably a more frequent case of addressing the counters is to get the current value of one of them, for example:

```none
# mlogtail bytes-received
1059498852
```
```none
# mlogtail rejected
4
```

### Log file statistics

In addition to working in real time, mlogtail can be used with a mail log file:

```none
# mlogtail -f /var/log/mail.log
```
or STDIN:
```none
# grep '^Apr  1' /var/log/mail.log | mlogtail -f -
```
for example, to get counters for some defined date.

## Installation

```none
go get -u github.com/hpcloud/tail golang.org/x/sys/unix &&
  go build && strip mlogtail &&
  cp mlogtail /usr/local/sbin &&
  chown root:bin /usr/local/sbin/mlogtail &&
  chmod 0711 /usr/local/sbin/mlogtail
```
