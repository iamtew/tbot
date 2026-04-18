// Package main is the executable entrypoint for tbot.
// It parses command-line options, loads configuration, and starts or stops the bot.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

const version = "0.1"

var buildInfo = "dev"
var buildTime = "unknown"
var gitCommit = "unknown"
var gitStatus = "unknown"
var gitBranch = "unknown"
var githubLink = "unknown"

func displayBuildInfo() {
	fmt.Printf("tbot version %s\n", version)
	fmt.Printf("Build: %s\n", buildInfo)
	fmt.Printf("Build time: %s\n", buildTime)
	fmt.Printf("Git commit: %s\n", gitCommit)
	fmt.Printf("Git status: %s\n", gitStatus)
	fmt.Printf("Git branch: %s\n", gitBranch)
	fmt.Printf("GitHub: %s\n", normalizeGitHubLink(githubLink))
}

func normalizeGitHubLink(link string) string {
	if strings.HasPrefix(link, "git@github.com:") {
		link = "https://github.com/" + strings.TrimPrefix(link, "git@github.com:")
	}
	if strings.HasPrefix(link, "https://github.com/") {
		link = strings.TrimSuffix(link, ".git")
	}
	return link
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "tbot - IRC bot\n\n")
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: tbot [options] <config-file>\n\n")
	flag.PrintDefaults()
}

// defaultPidFileForConfig returns the default pid file path for a given config file.
// If the config file is a TOML file, the PID file uses the same name with a .pid extension.
func defaultPidFileForConfig(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".toml" || ext != "" {
		return strings.TrimSuffix(path, ext) + ".pid"
	}
	return path + ".pid"
}

// stopBot reads the pid file and sends a stop signal to the running tbot process.
// On Windows it kills the process, and on POSIX platforms it sends SIGTERM.
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
		configPath    string
		logLevel      string
		quiet         bool
		daemon        bool
		verbose       bool
		debug         bool
		writeExample  string
		stop          bool
		pidFile       string
		showVersion   bool
		showBuildInfo bool
	)

	flag.StringVar(&writeExample, "example", "", "Write example config to path and exit")
	flag.StringVar(&writeExample, "e", "", "Write example config to path and exit")
	flag.StringVar(&configPath, "config", "", "Configuration file path")
	flag.StringVar(&configPath, "c", "", "Configuration file path")
	flag.StringVar(&logLevel, "loglevel", "", "Logging level: debug, verbose, info, warn, error")
	flag.StringVar(&logLevel, "L", "", "Logging level: debug, verbose, info, warn, error")
	flag.StringVar(&logFile, "logfile", "", "Log file path (default: <config-dir>/tbot.log)")
	flag.StringVar(&logFile, "l", "", "Log file path (default: <config-dir>/tbot.log)")
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
	flag.BoolVar(&showBuildInfo, "build-info", false, "Show detailed build information and exit")
	flag.Usage = usage
	flag.Parse()

	if showBuildInfo {
		displayBuildInfo()
		return
	}

	if showVersion {
		fmt.Printf("%s %s\n", version, buildInfo)
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

	// Set default log file if not specified
	if logFile == "" {
		logFile = filepath.Join(filepath.Dir(configPath), "tbot.log")
	}
	config.Bot.LogFile = logFile

	if !quiet {
		displayBuildInfo()
		fmt.Printf("tbot %s starting with config %s\n", version, configPath)
	}
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

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigc
		if !quiet {
			fmt.Fprintln(os.Stdout, "shutdown requested")
		}
		bot.requestShutdown()
	}()

	if err := bot.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "bot terminated with error: %v\n", err)
		os.Exit(1)
	}
}
