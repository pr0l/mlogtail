package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"os/user"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	//"runtime/pprof"

	"github.com/hpcloud/tail"
	"golang.org/x/sys/unix"
)

// type Config map[string]string
type Config struct {
	cmd           string
	cpuprofile    string
	subCmd        string
	setFlags      string
	listen        string
	lnNetworkType string
	lnAddress     string
	maillog       string
	maillogType   string
	socketOwner   string
	socketMode    int
	httpListen    string
	httpEnabled   bool
	initFromFile  bool
}

const (
	cmdAllowed = "stats|stats_reset|reset|tail"
)

func main() {
	cfg := new(Config)
	readCmdLine(cfg)

	/*
		if cfg.cpuprofile != "" {
			f, err := os.Create(cfg.cpuprofile)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			pprof.StartCPUProfile(f)
			defer pprof.StopCPUProfile()
		}
	*/
	if len(cfg.cmd) > 0 {
		switch cfg.cmd {
		case "tail": // start to tail of the log
			tailLog(cfg)
		default:
			if !strings.Contains(cfg.setFlags, "f") {
				getCurrentStats(cfg)
			}
		}
	}

	if strings.Contains(cfg.setFlags, "f") {
		var err error
		var logFile *os.File
		if cfg.maillog == "-" {
			logFile = os.Stdin
		} else {
			logFile, err = os.Open(cfg.maillog)
			if err != nil {
				fmt.Printf("Canot open logfile: %s\n", err)
				os.Exit(1)
			}
		}

		cfg.cmd = "file" // we are working with a disk saved file of STDIN
		PostfixParserInit(cfg)
		buf := bufio.NewReaderSize(logFile, 64*1024)
		var line string
		for {
			line, err = buf.ReadString('\n')
			if err != nil {
				break
			} else {
				PostfixLineParse(line)
			}
		}
		if err != io.EOF {
			fmt.Println(err)
			os.Exit(1)
		} else {
			fmt.Print(PostfixStats())
		}
		logFile.Close()
	}
}

func closeListener(ln net.Listener, cfg *Config) {
	if cfg.lnNetworkType == "unix" {
		fmt.Printf("Removing socket file %s\n", cfg.lnAddress)
		err := unix.Unlink(cfg.lnAddress)
		if err != nil {
			fmt.Printf("Cannot delete socket file %s: %s\n", cfg.lnAddress, err)
		}
	}
}

func createListener(cfg *Config) net.Listener {
	res, err := net.Listen(cfg.lnNetworkType, cfg.lnAddress)
	if err != nil {
		fmt.Printf("Cannot open %s: %s\n", cfg.listen, err)
		os.Exit(1)
	}

	if cfg.lnNetworkType == "unix" {
		// Set socket access permissions
		mode, _ := strconv.ParseInt(fmt.Sprintf("0%d", cfg.socketMode), 0, 64)
		err := unix.Chmod(cfg.lnAddress, uint32(mode))
		if err != nil {
			fmt.Printf("Cannot chmod: %s\n", err)
		}

		// Set socket owner and group if we are root
		if len(cfg.socketOwner) > 0 {
			if os.Geteuid() == 0 {
				err = setFileOwner(cfg.lnAddress, cfg.socketOwner)
				if err != nil {
					fmt.Printf("Cannot set socket owner: %s", err)
				}
			} else {
				fmt.Printf("You need to be a superuser (root) to set the socket owner\n")
			}
		}
	}
	return res
}

func getCurrentStats(cfg *Config) {
	conn, err := net.Dial(cfg.lnNetworkType, cfg.lnAddress)
	if err != nil {
		fmt.Printf("Cannot connect to log reader process: %s\n", err)
		return
	}

	var cmd string
	if len(cfg.subCmd) > 0 {
		cmd = cfg.subCmd
	} else {
		cmd = cfg.cmd
	}
	//buf := make([]byte, 384)
	buf := make([]byte, 2048)
	conn.Write([]byte(cmd))
	cnt, _ := conn.Read(buf)
	fmt.Printf("%s", string(buf[:cnt]))
	conn.Close()
}

func handleSIGINTKILL(ln net.Listener, cfg *Config) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Printf("\nReceived termination signal\n")
	closeListener(ln, cfg)

	os.Exit(0)
}

func readCmdLine(cfg *Config) {
	var cpuprofile, listen, maillog, maillogType, socketOwner, httpListen string
	var socketMode int
	var initFromFile bool

	//flag.StringVar(&cpuprofile, "cpuprofile", "", "write cpu profile to file")
	flag.StringVar(&maillog, "f", "/var/log/mail.log", "Mail log file path, if the path is \"-\" then read from STDIN")
	flag.Bool("h", false, "Show this help")
	flag.StringVar(&httpListen, "http", "", "HTTP server address (e.g., :8080 or 0.0.0.0:8080) to serve stats as JSON")
	flag.BoolVar(&initFromFile, "init-from-file", false, "Read entire log file on startup to initialize counters, then continue tailing")
	flag.StringVar(&listen, "l", "unix:/var/run/mlogtail.sock", "Log reader process is listening for commands on a socket file, or IPv4:PORT,\nor [IPv6]:PORT")
	flag.StringVar(&socketOwner, "o", "", "Set a socket OWNER[:GROUP] while listening on a socket file")
	flag.IntVar(&socketMode, "p", 666, "Set a socket access permissions while listening on a socket file")
	flag.StringVar(&maillogType, "t", "postfix", "Mail log type. It is \"postfix\" only allowed for now")
	flag.Bool("v", false, "Show version information and exit")
	flag.Parse()

	// create a list of explicitly set flags
	var showHelp, showVersion bool
	fsetFunc := func(f *flag.Flag) {
		cfg.setFlags += f.Name
		if f.Name == "h" {
			showHelp = true
		}
		if f.Name == "v" {
			showVersion = true
		}
	}
	flag.Visit(fsetFunc)
	if len(os.Args) == 1 || showHelp {
		usage()
	}
	if showVersion {
		fmt.Printf("%s v. %s, %s\n", PROGNAME, VERSION, runtime.Version())
		os.Exit(0)
	}

	cfg.cpuprofile = cpuprofile
	cfg.listen = listen
	cfg.maillog = maillog
	cfg.maillogType = maillogType
	cfg.socketOwner = socketOwner
	cfg.httpListen = httpListen
	cfg.httpEnabled = len(httpListen) > 0
	cfg.initFromFile = initFromFile

	// get not options parameter (command)
	if flag.NArg() > 0 {
		cmds := flag.Args()
		if strings.Contains(cmdAllowed, cmds[0]) {
			cfg.cmd = cmds[0]
		} else if maillogType == "postfix" && strArrayLookup(PostfixStatusNames[:], cmds[0]) {
			cfg.cmd = "stats"
			cfg.subCmd = cmds[0]
		} else {
			fmt.Printf("Command can be one of \"%s\"\n", cmdAllowed+"|"+strings.Join(PostfixStatusNames[:], "|"))
			os.Exit(1)
		}
	}

	// some configuratioin of tailing process
	if cfg.cmd == "tail" {
		if socketMode <= 777 {
			cfg.socketMode = socketMode
		} else {
			fmt.Printf("File mode cannot be greater than 777, it is set to 666\n")
			cfg.socketMode = 666
		}
	}

	if listen[:5] == "unix:" {
		cfg.lnNetworkType = "unix"
		cfg.lnAddress = listen[5:]
	} else {
		cfg.lnNetworkType = "tcp"
		cfg.lnAddress = listen
	}
}

// setFileOwner gets file name and OWNER[:GROUP] as owner,
// converts OWNER:GROUP to numeric IDs if required and
// apply it to the file
func setFileOwner(fpath, owner string) error {
	ugAr := strings.Split(owner, ":")
	if len(ugAr) > 2 {
		return fmt.Errorf("Incorrect owner %s", owner)
	}

	var uid, gid int
	numRe := regexp.MustCompile(`^\d+$`)

	if numRe.MatchString(ugAr[0]) { // Get UID
		uid, _ = strconv.Atoi(ugAr[0])
	} else {
		if u, err := user.Lookup(ugAr[0]); err == nil {
			uid, _ = strconv.Atoi(u.Uid)
		} else {
			return err
		}
	}

	if len(ugAr) == 2 {
		if numRe.MatchString(ugAr[1]) { // Get GID
			gid, _ = strconv.Atoi(ugAr[1])
		} else {
			if g, err := user.LookupGroup(ugAr[1]); err == nil {
				gid, _ = strconv.Atoi(g.Gid)
			} else {
				return err
			}
		}
	} else {
		gid = -1 // do not change GID
	}

	// Set OWNER:GROUP
	return os.Chown(fpath, uid, gid)
}

func strArrayLookup(a []string, s string) bool {
	for _, v := range a {
		if s == v {
			return true
		}
	}
	return false
}

// initCountersFromFile читает весь лог-файл и инициализирует счётчики
func initCountersFromFile(filename string) error {
	if filename == "-" {
		return fmt.Errorf("Cannot initialize from STDIN")
	}

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("Cannot open log file for initialization: %v", err)
	}
	defer file.Close()

	fmt.Printf("Initializing counters from log file %s...\n", filename)

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		PostfixLineParse(scanner.Text())
		lineCount++
		if lineCount%10000 == 0 {
			fmt.Printf("Processed %d lines...\n", lineCount)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("Error reading log file: %v", err)
	}

	fmt.Printf("Initialization complete: processed %d lines\n", lineCount)
	return nil
}

func tailLog(cfg *Config) {
	ln := createListener(cfg)
	defer closeListener(ln, cfg)
	go handleSIGINTKILL(ln, cfg)
	go PostfixCmgHandle(ln)

	// Запускаем HTTP сервер, если указан флаг -http
	if cfg.httpEnabled {
		go startHTTPServer(cfg.httpListen)
	}

	tailCfg := tail.Config{
		Location: &tail.SeekInfo{Offset: 0, Whence: 2},
		ReOpen:   true,
		Follow:   true,
		Logger:   tail.DiscardingLogger,
	}
	t, err := tail.TailFile(cfg.maillog, tailCfg)
	if err != nil {
		fmt.Printf("Cannot tail mail log file: %v\n", err)
		closeListener(ln, cfg)
		os.Exit(1)
	}

	PostfixParserInit(cfg)

	// Инициализация счётчиков из всего файла, если указан флаг
	if cfg.initFromFile {
		if err := initCountersFromFile(cfg.maillog); err != nil {
			fmt.Printf("Warning: %v\n", err)
		}
	}

	for line := range t.Lines {
		PostfixLineParse(line.Text)
	}
}

func usage() {
	pname := os.Args[0]
	fmt.Printf("Usage:\n  %s [OPTIONS] tail\n", pname)
	fmt.Printf("  %s [OPTIONS] \"stats | stats_reset | reset\"\n", pname)
	fmt.Printf("  %s [OPTIONS] <COUNTER_NAME>\n", pname)
	fmt.Printf("  %s -f <LOG_FILE_NAME>\n\nOptions:\n", pname)
	flag.PrintDefaults()
	os.Exit(0)
}
