package main

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelVerbose
	LevelInfo
	LevelWarn
	LevelError
)

func parseLogLevel(value string) LogLevel {
	switch strings.ToLower(value) {
	case "debug":
		return LevelDebug
	case "verbose":
		return LevelVerbose
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}

type Bot struct {
	config        *Config
	configPath    string
	pidFile       string
	conn          net.Conn
	reader        *bufio.Reader
	writer        *bufio.Writer
	logger        *log.Logger
	logLevel      LogLevel
	quiet         bool
	quit          chan struct{}
	restart       chan struct{}
	barrels       map[string]Barrel
	barrelConfigs map[string]*BarrelConfig
	channelModes  map[string]map[string]rune
	adminMasks    map[string]struct{}
	lastURL       map[string]*URLMetadata
	mu            sync.Mutex
}

type URLMetadata struct {
	URL    string
	Title  string
	Detail string
}

func NewBot(cfg *Config, configPath, pidFile string, quiet bool, overrideLevel string) (*Bot, error) {
	output := io.Discard
	if !quiet {
		output = os.Stdout
	}
	logger := log.New(output, "[tbot] ", log.LstdFlags)
	level := cfg.Bot.LogLevel
	if overrideLevel != "" {
		level = overrideLevel
	}
	bot := &Bot{
		config:        cfg,
		configPath:    configPath,
		pidFile:       pidFile,
		logger:        logger,
		logLevel:      parseLogLevel(level),
		quiet:         quiet,
		quit:          make(chan struct{}),
		restart:       make(chan struct{}, 1),
		barrels:       make(map[string]Barrel),
		barrelConfigs: cfg.Barrels,
		channelModes:  make(map[string]map[string]rune),
		adminMasks:    make(map[string]struct{}),
		lastURL:       make(map[string]*URLMetadata),
	}

	if cfg.Bot.LogFile != "" {
		file, err := os.OpenFile(cfg.Bot.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, err
		}
		bot.logger.SetOutput(io.MultiWriter(output, file))
	}

	for _, mask := range cfg.Bot.Admins {
		bot.adminMasks[mask] = struct{}{}
	}

	rand.Seed(time.Now().UnixNano())
	bot.registerBarrel(NewURLBarrel())
	bot.registerBarrel(NewFishBarrel())
	bot.applyBarrelConfig()

	return bot, nil
}

func (b *Bot) writePidFile() error {
	if b.pidFile == "" {
		return nil
	}
	data := []byte(fmt.Sprintf("%d\n", os.Getpid()))
	return os.WriteFile(b.pidFile, data, 0o644)
}

func (b *Bot) removePidFile() {
	if b.pidFile == "" {
		return
	}
	_ = os.Remove(b.pidFile)
}

func (b *Bot) Run() error {
	if err := b.writePidFile(); err != nil {
		return err
	}
	defer b.removePidFile()

	for {
		if err := b.connect(); err != nil {
			return err
		}
		b.logInfo("connected to %s:%d", b.config.Network.Server, b.config.Network.Port)
		err := b.readLoop()
		if err != nil && !errors.Is(err, io.EOF) {
			b.logWarn("connection loop ended: %v", err)
		}
		select {
		case <-b.quit:
			b.logInfo("shutdown requested")
			return nil
		case <-b.restart:
			b.logInfo("reconnecting to server")
			continue
		default:
			return nil
		}
	}
}

func (b *Bot) connect() error {
	address := fmt.Sprintf("%s:%d", b.config.Network.Server, b.config.Network.Port)
	var conn net.Conn
	var err error
	if b.config.Network.TLS {
		conn, err = tls.Dial("tcp", address, &tls.Config{InsecureSkipVerify: true})
	} else {
		conn, err = net.Dial("tcp", address)
	}
	if err != nil {
		return err
	}
	b.conn = conn
	b.reader = bufio.NewReader(conn)
	b.writer = bufio.NewWriter(conn)
	if b.config.Network.Password != "" {
		b.sendRaw("PASS %s", b.config.Network.Password)
	}
	b.sendRaw("NICK %s", b.config.Bot.Nick)
	b.sendRaw("USER %s 0 * :%s", b.config.Bot.User, b.config.Bot.RealName)
	return nil
}

func (b *Bot) readLoop() error {
	for {
		line, err := b.reader.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimRight(line, "\r\n")
		b.logDebug("<= %s", line)
		msg := ParseIRCLine(line)
		b.handleMessage(msg)
	}
}

func (b *Bot) handleMessage(msg IRCMessage) {
	switch msg.Command {
	case "PING":
		if msg.Trailing == "" && len(msg.Params) > 0 {
			msg.Trailing = msg.Params[0]
		}
		b.sendRaw("PONG :%s", msg.Trailing)
	case "001":
		b.joinChannels()
	case "353":
		b.updateNames(msg)
	case "366":
		// end of names list
	case "433":
		b.sendRaw("NICK %s_", b.config.Bot.Nick)
	case "PRIVMSG":
		b.handlePrivMsg(msg)
	case "JOIN":
		b.handleJoin(msg)
	case "PART", "QUIT":
		b.handlePart(msg)
	case "NICK":
		b.handleNickChange(msg)
	}
}

func (b *Bot) joinChannels() {
	for _, channel := range b.config.Network.Channels {
		b.sendRaw("JOIN %s", channel)
	}
}

func (b *Bot) updateNames(msg IRCMessage) {
	if len(msg.Params) < 4 {
		return
	}
	channel := msg.Params[2]
	names := strings.Fields(msg.Trailing)
	b.mu.Lock()
	modes, ok := b.channelModes[channel]
	if !ok {
		modes = make(map[string]rune)
		b.channelModes[channel] = modes
	}
	for _, name := range names {
		mode := rune(' ')
		if strings.HasPrefix(name, "@") {
			mode = '@'
			name = strings.TrimPrefix(name, "@")
		} else if strings.HasPrefix(name, "%") {
			mode = '%'
			name = strings.TrimPrefix(name, "%")
		} else if strings.HasPrefix(name, "+") {
			mode = '+'
			name = strings.TrimPrefix(name, "+")
		}
		modes[name] = mode
	}
	b.mu.Unlock()
}

func (b *Bot) handleJoin(msg IRCMessage) {
	channel := msg.Trailing
	if channel == "" && len(msg.Params) > 0 {
		channel = msg.Params[0]
	}
	if channel == "" || msg.Nick == "" {
		return
	}
	b.mu.Lock()
	if _, ok := b.channelModes[channel]; !ok {
		b.channelModes[channel] = make(map[string]rune)
	}
	b.channelModes[channel][msg.Nick] = ' '
	b.mu.Unlock()
}

func (b *Bot) handlePart(msg IRCMessage) {
	channel := ""
	if len(msg.Params) > 0 {
		channel = msg.Params[0]
	}
	b.mu.Lock()
	if channel != "" {
		if modes, ok := b.channelModes[channel]; ok {
			delete(modes, msg.Nick)
		}
	}
	b.mu.Unlock()
}

func (b *Bot) handleNickChange(msg IRCMessage) {
	oldNick := msg.Nick
	newNick := msg.Trailing
	b.mu.Lock()
	for _, modes := range b.channelModes {
		if mode, ok := modes[oldNick]; ok {
			delete(modes, oldNick)
			modes[newNick] = mode
		}
	}
	b.mu.Unlock()
}

func (b *Bot) handlePrivMsg(msg IRCMessage) {
	if len(msg.Params) < 1 {
		return
	}
	target := msg.Params[0]
	text := msg.Trailing
	sourceNick := msg.Nick
	if sourceNick == "" {
		return
	}

	if target == b.config.Bot.Nick {
		b.handlePrivateCommand(msg)
		return
	}

	if strings.HasPrefix(text, b.config.Bot.CommandPrefix) {
		line := strings.TrimPrefix(text, b.config.Bot.CommandPrefix)
		cmd, args := splitCommand(line)
		if cmd == "" {
			return
		}
		if b.dispatchPublicCommand(target, sourceNick, cmd, args) {
			return
		}
	}

	for _, barrel := range b.barrels {
		if barrel.Enabled() {
			barrel.HandleMessage(b, target, sourceNick, text)
		}
	}
}

func splitCommand(line string) (string, []string) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return "", nil
	}
	return strings.ToLower(parts[0]), parts[1:]
}

func (b *Bot) dispatchPublicCommand(target, nick, command string, args []string) bool {
	for _, barrel := range b.barrels {
		if barrel.Enabled() && barrel.HandleCommand(b, target, nick, command, args) {
			return true
		}
	}

	switch command {
	case "whoami":
		b.sendMessage(target, fmt.Sprintf("%s is %s", nick, b.permissionLevel(target, nick)))
		return true
	case "help":
		b.sendMessage(target, fmt.Sprintf("Commands: whoami, help, plus barrel commands and fish/url barrels if enabled."))
		return true
	}
	return false
}

func (b *Bot) handlePrivateCommand(msg IRCMessage) {
	if !b.isAdmin(msg.Prefix) {
		b.sendMessage(msg.Nick, "admin commands require an authorized mask")
		return
	}
	line := strings.TrimSpace(msg.Trailing)
	cmd, args := splitCommand(line)
	switch cmd {
	case "reload":
		b.reloadConfig()
		b.sendMessage(msg.Nick, "configuration reloaded")
	case "barrel":
		b.handleAdminBarrel(msg.Nick, args)
	case "get":
		b.handleAdminGet(msg.Nick, args)
	case "set":
		b.handleAdminSet(msg.Nick, args)
	case "write":
		b.handleAdminWrite(msg.Nick)
	case "reconnect":
		b.sendMessage(msg.Nick, "reconnecting...")
		b.requestReconnect()
	case "stop", "shutdown":
		b.sendMessage(msg.Nick, "shutting down")
		b.requestShutdown()
	default:
		b.sendMessage(msg.Nick, "unknown admin command")
	}
}

func (b *Bot) handleAdminBarrel(replyTo string, args []string) {
	if len(args) == 0 {
		b.sendMessage(replyTo, "usage: barrel list|enable|disable <name>")
		return
	}
	sub := strings.ToLower(args[0])
	switch sub {
	case "list":
		lines := []string{"barrels:"}
		for name, barrel := range b.barrels {
			lines = append(lines, fmt.Sprintf("- %s: %t", name, barrel.Enabled()))
		}
		b.sendMessage(replyTo, strings.Join(lines, " | "))
	case "enable", "disable":
		if len(args) < 2 {
			b.sendMessage(replyTo, "usage: barrel enable|disable <name>")
			return
		}
		name := strings.ToLower(args[1])
		barrel, ok := b.barrels[name]
		if !ok {
			b.sendMessage(replyTo, "unknown barrel: "+name)
			return
		}
		enabled := sub == "enable"
		barrel.SetEnabled(enabled)
		if b.config.Barrels == nil {
			b.config.Barrels = make(map[string]*BarrelConfig)
		}
		cfg := b.config.Barrels[name]
		if cfg == nil {
			cfg = &BarrelConfig{}
			b.config.Barrels[name] = cfg
		}
		cfg.Enabled = enabled
		b.sendMessage(replyTo, fmt.Sprintf("barrel %s %s", name, map[bool]string{true: "enabled", false: "disabled"}[enabled]))
	default:
		b.sendMessage(replyTo, "usage: barrel list|enable|disable <name>")
	}
}

func (b *Bot) handleAdminGet(replyTo string, args []string) {
	if len(args) == 0 {
		b.sendMessage(replyTo, "usage: get <config.key>")
		return
	}
	key := strings.Join(args, " ")
	switch strings.ToLower(key) {
	case "bot.nick":
		b.sendMessage(replyTo, b.config.Bot.Nick)
	case "bot.prefix", "bot.command_prefix":
		b.sendMessage(replyTo, b.config.Bot.CommandPrefix)
	case "bot.log_level":
		b.sendMessage(replyTo, b.config.Bot.LogLevel)
	case "network.server":
		b.sendMessage(replyTo, b.config.Network.Server)
	case "network.port":
		b.sendMessage(replyTo, fmt.Sprintf("%d", b.config.Network.Port))
	default:
		b.sendMessage(replyTo, "unknown config key")
	}
}

func (b *Bot) handleAdminSet(replyTo string, args []string) {
	if len(args) < 2 {
		b.sendMessage(replyTo, "usage: set <config.key> <value>")
		return
	}
	key := strings.ToLower(args[0])
	value := strings.Join(args[1:], " ")
	switch key {
	case "bot.nick":
		b.config.Bot.Nick = value
		b.sendRaw("NICK %s", value)
		b.sendMessage(replyTo, "nick updated")
	case "bot.prefix", "bot.command_prefix":
		b.config.Bot.CommandPrefix = value
		b.sendMessage(replyTo, "command prefix updated")
	case "bot.log_level":
		b.config.Bot.LogLevel = value
		b.logLevel = parseLogLevel(value)
		b.sendMessage(replyTo, "log level updated")
	case "bot.log_file":
		b.config.Bot.LogFile = value
		b.sendMessage(replyTo, "log file updated (restart to change existing log writer)")
	default:
		b.sendMessage(replyTo, "unknown config key")
	}
}

func (b *Bot) handleAdminWrite(replyTo string) {
	if err := b.config.Save(b.configPath); err != nil {
		b.sendMessage(replyTo, "failed to write config: "+err.Error())
		return
	}
	b.sendMessage(replyTo, "configuration written to disk")
}

func (b *Bot) reloadConfig() {
	cfg, err := LoadConfig(b.configPath)
	if err != nil {
		b.logWarn("failed to reload config: %v", err)
		return
	}
	b.config = cfg
	b.config.Bot.LogLevel = cfg.Bot.LogLevel
	b.applyBarrelConfig()
	b.logInfo("configuration reloaded")
}

func (b *Bot) applyBarrelConfig() {
	for name, barrel := range b.barrels {
		enabled := false
		if cfg, ok := b.config.Barrels[name]; ok {
			enabled = cfg.Enabled
		}
		barrel.SetEnabled(enabled)
	}
}

func (b *Bot) requestReconnect() {
	select {
	case b.restart <- struct{}{}:
		// signal reconnect
	default:
	}
	if b.conn != nil {
		b.conn.Close()
	}
}

func (b *Bot) requestShutdown() {
	close(b.quit)
	if b.conn != nil {
		b.conn.Close()
	}
}

func (b *Bot) permissionLevel(channel, nick string) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	if modes, ok := b.channelModes[channel]; ok {
		switch modes[nick] {
		case '@':
			return "op"
		case '%':
			return "halfop"
		case '+':
			return "voice"
		}
	}
	return "normal"
}

func (b *Bot) isAdmin(mask string) bool {
	_, ok := b.adminMasks[mask]
	return ok
}

func (b *Bot) sendRaw(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	b.logDebug("=> %s", message)
	if b.writer == nil {
		return
	}
	b.writer.WriteString(message + "\r\n")
	b.writer.Flush()
}

func (b *Bot) sendMessage(target, text string) {
	b.sendRaw("PRIVMSG %s :%s", target, text)
}

func (b *Bot) registerBarrel(barrel Barrel) {
	b.barrels[strings.ToLower(barrel.Name())] = barrel
	if cfg, ok := b.config.Barrels[strings.ToLower(barrel.Name())]; ok {
		barrel.SetEnabled(cfg.Enabled)
	}
}

func (b *Bot) logDebug(format string, args ...interface{}) {
	if b.logLevel <= LevelDebug {
		b.logger.Printf("DEBUG "+format, args...)
	}
}

func (b *Bot) logInfo(format string, args ...interface{}) {
	if b.logLevel <= LevelInfo {
		b.logger.Printf("INFO "+format, args...)
	}
}

func (b *Bot) logWarn(format string, args ...interface{}) {
	if b.logLevel <= LevelWarn {
		b.logger.Printf("WARN "+format, args...)
	}
}

func (b *Bot) logError(format string, args ...interface{}) {
	if b.logLevel <= LevelError {
		b.logger.Printf("ERROR "+format, args...)
	}
}
