package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

const version = "0.1"

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "tbot - IRC bot\n\n")
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: tbot [options] <config-file>\n\n")
	flag.PrintDefaults()
}

func defaultPidFileForConfig(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".toml" || ext != "" {
		return strings.TrimSuffix(path, ext) + ".pid"
	}
	return path + ".pid"
}

func stopBot(pidFile string) error {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return err
	}
	pidString := strings.TrimSpace(string(data))
	if pidString == "" {
		return fmt.Errorf("pid file %s is empty", pidFile)
	}
	pid, err := strconv.Atoi(pidString)
	if err != nil {
		return err
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	if runtime.GOOS == "windows" {
		return proc.Kill()
	}
	return proc.Signal(syscall.SIGTERM)
}

func main() {
	var (
		configPath   string
		logLevel     string
		quiet        bool
		daemon       bool
		verbose      bool
		debug        bool
		writeExample string
		stop         bool
		pidFile      string
		showVersion  bool
	)

	flag.StringVar(&writeExample, "example", "", "Write example config to path and exit")
	flag.StringVar(&writeExample, "e", "", "Write example config to path and exit")
	flag.StringVar(&configPath, "config", "", "Configuration file path")
	flag.StringVar(&configPath, "c", "", "Configuration file path")
	flag.StringVar(&logLevel, "loglevel", "", "Logging level: debug, verbose, info, warn, error")
	flag.StringVar(&logLevel, "L", "", "Logging level: debug, verbose, info, warn, error")
	flag.BoolVar(&quiet, "quiet", false, "No output when running")
	flag.BoolVar(&quiet, "Q", false, "No output when running")
	flag.BoolVar(&daemon, "daemon", false, "Run in background and quiet")
	flag.BoolVar(&daemon, "D", false, "Run in background and quiet")
	flag.BoolVar(&verbose, "verbose", false, "Verbose logging, alias of --loglevel=verbose")
	flag.BoolVar(&debug, "debug", false, "Debug logging, alias of --loglevel=debug")
	flag.BoolVar(&debug, "d", false, "Debug logging, alias of --loglevel=debug")
	flag.BoolVar(&stop, "stop", false, "Stop the running bot and exit")
	flag.BoolVar(&stop, "S", false, "Stop the running bot and exit")
	flag.StringVar(&pidFile, "pidfile", "", "PID file path")
	flag.StringVar(&pidFile, "P", "", "PID file path")
	flag.BoolVar(&showVersion, "version", false, "Show version and exit")
	flag.BoolVar(&showVersion, "V", false, "Show version and exit")
	flag.Usage = usage
	flag.Parse()

	if showVersion {
		fmt.Println(version)
		return
	}

	if stop {
		if flag.NArg() > 0 && configPath == "" {
			configPath = flag.Arg(0)
		}
		if pidFile == "" {
			if configPath == "" {
				fmt.Fprintln(os.Stderr, "error: configuration file path or --pidfile is required to stop the bot")
				usage()
				os.Exit(1)
			}
			pidFile = defaultPidFileForConfig(filepath.Clean(configPath))
		}
		if err := stopBot(pidFile); err != nil {
			fmt.Fprintf(os.Stderr, "failed to stop bot: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "stop request sent using pid file %s\n", pidFile)
		return
	}

	if writeExample != "" {
		example := ExampleConfig()
		if err := example.Save(writeExample); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write example config: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "example config written to %s\n", writeExample)
		return
	}

	if flag.NArg() > 0 && configPath == "" {
		configPath = flag.Arg(0)
	}

	if configPath == "" {
		fmt.Fprintln(os.Stderr, "error: configuration file path is required")
		usage()
		os.Exit(1)
	}

	if daemon {
		quiet = true
	}

	if verbose {
		logLevel = "verbose"
	}
	if debug {
		logLevel = "debug"
	}
	if logLevel == "" {
		logLevel = ""
	}

	configPath = filepath.Clean(configPath)
	config, err := LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed loading config %s: %v\n", configPath, err)
		os.Exit(1)
	}

	if pidFile == "" {
		if config.Bot.PidFile != "" {
			pidFile = filepath.Clean(config.Bot.PidFile)
		} else {
			pidFile = defaultPidFileForConfig(configPath)
		}
	}

	bot, err := NewBot(config, configPath, pidFile, quiet, logLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize bot: %v\n", err)
		os.Exit(1)
	}

	if err := bot.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "bot terminated with error: %v\n", err)
		os.Exit(1)
	}
}
