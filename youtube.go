// Package main includes the YouTube barrel for enhanced YouTube link handling.
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

type YouTubeBarrel struct {
	enabled  bool
	cooldown time.Duration
	lastSeen map[string]time.Time
}

func NewYouTubeBarrel() *YouTubeBarrel {
	return &YouTubeBarrel{
		cooldown: 60 * time.Second,
		lastSeen: make(map[string]time.Time),
	}
}

func (b *YouTubeBarrel) Name() string {
	return "youtube"
}

func (b *YouTubeBarrel) Enabled() bool {
	return b.enabled
}

func (b *YouTubeBarrel) SetEnabled(enabled bool) {
	b.enabled = enabled
}

func (b *YouTubeBarrel) LoadConfig(cfg *BarrelConfig) {
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

var youtubeRegex = regexp.MustCompile(`https?://(?:www\.)?(?:youtube\.com/watch\?v=|youtu\.be/)([a-zA-Z0-9_-]+)`)

func (b *YouTubeBarrel) HandleMessage(bot *Bot, channel, nick, text string) {
	if nick == bot.config.Bot.Nick {
		return
	}
	urls := youtubeRegex.FindAllString(text, -1)
	for _, rawURL := range urls {
		key := channel + "|" + rawURL
		if last, ok := b.lastSeen[key]; ok && time.Since(last) < b.cooldown {
			continue
		}
		title, likes, uploadDate, err := fetchYouTubeMetadata(rawURL)
		if err != nil {
			bot.logDebug("YouTube fetch failed: %v", err)
			continue
		}
		b.lastSeen[key] = time.Now()
		bot.sendMessage(channel, fmt.Sprintf("[YouTube] %s | Likes: %s | Uploaded: %s", title, likes, uploadDate))
		return
	}
}

func (b *YouTubeBarrel) HandleCommand(bot *Bot, channel, nick, command string, args []string) bool {
	// No commands for now
	return false
}

func fetchYouTubeMetadata(rawURL string) (string, string, string, error) {
	// For simplicity, fetch title as before, and placeholder for likes and date
	// In a real implementation, use YouTube API or parse page JSON
	httpClient := http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Get(rawURL)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 100*1024))
	if err != nil {
		return "", "", "", err
	}
	content := string(body)
	match := titleRegex.FindStringSubmatch(content)
	title := strings.TrimSpace("[no title]")
	if len(match) > 1 {
		title = strings.TrimSpace(match[1])
	}
	// Placeholder values
	likes := "N/A"
	uploadDate := "N/A"
	return title, likes, uploadDate, nil
}
