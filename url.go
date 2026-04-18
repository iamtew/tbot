package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type URLBarrel struct {
	enabled bool
	cache   map[string]*URLMetadata
}

func NewURLBarrel() *URLBarrel {
	return &URLBarrel{cache: make(map[string]*URLMetadata)}
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

func (b *URLBarrel) LoadConfig(_ *BarrelConfig) {
}

var urlRegex = regexp.MustCompile(`https?://[^\s]+`)
var titleRegex = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)

func (b *URLBarrel) HandleMessage(bot *Bot, channel, nick, text string) {
	if nick == bot.config.Bot.Nick {
		return
	}
	urls := urlRegex.FindAllString(text, -1)
	for _, rawURL := range urls {
		title, detail, err := fetchURLMetadata(rawURL)
		if err != nil {
			bot.logDebug("url fetch failed: %v", err)
			continue
		}
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
