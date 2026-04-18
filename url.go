package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// URLBarrel watches chat messages for URLs and posts page titles back to the channel.
// It also caches the last seen URL per channel and respects a cooldown interval.
type URLBarrel struct {
	enabled  bool
	cache    map[string]*URLMetadata
	cooldown time.Duration
	lastSeen map[string]time.Time
}

// NewURLBarrel constructs the URL barrel with default cooldown settings.
func NewURLBarrel() *URLBarrel {
	return &URLBarrel{
		cache:    make(map[string]*URLMetadata),
		cooldown: 60 * time.Second,
		lastSeen: make(map[string]time.Time),
	}
}

func (b *URLBarrel) Name() string {
	return "url"
}

func (b *URLBarrel) Enabled() bool {
	return b.enabled
}

func (b *URLBarrel) SetEnabled(enabled bool) {
	b.enabled = enabled
}

// LoadConfig reads barrel-specific settings from the common barrel config structure.
// `cooldown` is interpreted in seconds and controls URL repeat suppression.
func (b *URLBarrel) LoadConfig(cfg *BarrelConfig) {
	b.cooldown = 60 * time.Second
	if cfg == nil || cfg.Settings == nil {
		return
	}
	if raw, ok := cfg.Settings["cooldown"]; ok {
		switch value := raw.(type) {
		case int64:
			b.cooldown = time.Duration(value) * time.Second
		case int:
			b.cooldown = time.Duration(value) * time.Second
		case float64:
			b.cooldown = time.Duration(int(value)) * time.Second
		case string:
			if secs, err := strconv.Atoi(value); err == nil {
				b.cooldown = time.Duration(secs) * time.Second
			}
		}
	}
}

var urlRegex = regexp.MustCompile(`https?://[^\s]+`)
var titleRegex = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)

// HandleMessage processes incoming channel messages and resolves the first URL
// that is not suppressed by the barrel cooldown.
func (b *URLBarrel) HandleMessage(bot *Bot, channel, nick, text string) {
	if nick == bot.config.Bot.Nick {
		return
	}
	urls := urlRegex.FindAllString(text, -1)
	for _, rawURL := range urls {
		key := channel + "|" + rawURL
		if last, ok := b.lastSeen[key]; ok && time.Since(last) < b.cooldown {
			continue
		}
		title, detail, err := fetchURLMetadata(rawURL)
		if err != nil {
			bot.logDebug("url fetch failed: %v", err)
			continue
		}
		b.lastSeen[key] = time.Now()
		b.cache[channel] = &URLMetadata{URL: rawURL, Title: title, Detail: detail}
		bot.sendMessage(channel, fmt.Sprintf("[%s] %s", rawURL, title))
		return
	}
}

func (b *URLBarrel) HandleCommand(bot *Bot, channel, nick, command string, args []string) bool {
	if command != "more" {
		return false
	}
	meta, ok := b.cache[channel]
	if !ok {
		bot.sendMessage(channel, "no recent URL information available")
		return true
	}
	bot.sendMessage(channel, fmt.Sprintf("more: %s - %s", meta.Title, meta.Detail))
	return true
}

func fetchURLMetadata(rawURL string) (string, string, error) {
	httpClient := http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Get(rawURL)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 100*1024))
	if err != nil {
		return "", "", err
	}
	content := string(body)
	match := titleRegex.FindStringSubmatch(content)
	title := strings.TrimSpace("[no title]")
	if len(match) > 1 {
		title = strings.TrimSpace(match[1])
	}
	detail := fmt.Sprintf("%s %s", resp.Header.Get("Content-Type"), resp.Status)
	return title, detail, nil
}
