// Package main includes the YouTube barrel for enhanced YouTube link handling.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type YouTubeBarrel struct {
	enabled  bool
	cooldown time.Duration
	lastSeen map[string]time.Time
	apikey   string
}

func NewYouTubeBarrel() *YouTubeBarrel {
	return &YouTubeBarrel{
		cooldown: 60 * time.Second,
		lastSeen: make(map[string]time.Time),
		apikey:   "",
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
	if cfg == nil {
		return
	}
	if cfg.Settings != nil {
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
	if cfg.Apikey != "" {
		if cfg.Apikey == "your_youtube_api_key_here" {
			b.apikey = ""
		} else {
			b.apikey = cfg.Apikey
		}
	} else if cfg.Settings != nil {
		if raw, ok := cfg.Settings["apikey"]; ok {
			if key, ok := raw.(string); ok {
				if key == "your_youtube_api_key_here" {
					b.apikey = ""
				} else {
					b.apikey = key
				}
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
		title, likes, dislikes, uploadDate, channelTitle, err := b.fetchYouTubeMetadata(bot.logDebug, rawURL)
		if err != nil {
			bot.logDebug("YouTube fetch failed: %v", err)
			continue
		}
		b.lastSeen[key] = time.Now()
		message := fmt.Sprintf("[YouTube] %s | Likes: %s | Uploaded: %s | Channel: %s", title, likes, uploadDate, channelTitle)
		if dislikes != "" {
			message = fmt.Sprintf("[YouTube] %s | Likes: %s | Dislikes: %s | Uploaded: %s | Channel: %s", title, likes, dislikes, uploadDate, channelTitle)
		}
		bot.sendMessage(channel, message)
		return
	}
}

func (b *YouTubeBarrel) HandleCommand(bot *Bot, channel, nick, command string, args []string) bool {
	// No commands for now
	return false
}

func (b *YouTubeBarrel) fetchYouTubeMetadata(logger func(format string, args ...interface{}), rawURL string) (string, string, string, string, string, error) {
	// Extract video ID
	match := youtubeRegex.FindStringSubmatch(rawURL)
	if len(match) < 2 {
		logger("YouTube: Invalid URL format: %s", rawURL)
		return "", "", "", "", "", fmt.Errorf("invalid YouTube URL")
	}
	videoID := match[1]
	logger("YouTube: Extracted video ID: %s", videoID)

	if b.apikey == "" {
		logger("YouTube: No API key configured, falling back to oEmbed")
		// Fallback to oEmbed for title
		oembedURL := fmt.Sprintf("https://www.youtube.com/oembed?url=https://www.youtube.com/watch?v=%s&format=json", videoID)
		logger("YouTube: oEmbed URL: %s", oembedURL)
		resp, err := http.Get(oembedURL)
		if err != nil {
			logger("YouTube: oEmbed HTTP GET error: %v", err)
			return "", "", "", "", "", err
		}
		defer resp.Body.Close()
		logger("YouTube: oEmbed response status: %s", resp.Status)
		if resp.StatusCode != 200 {
			logger("YouTube: oEmbed request failed: %s", resp.Status)
			return "", "", "", "", "", fmt.Errorf("oEmbed request failed: %s", resp.Status)
		}
		var oembed map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&oembed); err != nil {
			logger("YouTube: oEmbed JSON decode error: %v", err)
			return "", "", "", "", "", err
		}
		title, ok := oembed["title"].(string)
		if !ok {
			logger("YouTube: oEmbed missing title")
			title = "[no title]"
		}
		logger("YouTube: oEmbed success, title: %s", title)
		// Placeholders
		likes := "N/A"
		dislikes := ""
		uploadDate := "N/A"
		channelTitle := "N/A"
		return title, likes, dislikes, uploadDate, channelTitle, nil
	}

	// Use YouTube Data API
	logger("YouTube: Using API key length: %d", len(b.apikey))
	apiURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?id=%s&key=%s&part=snippet,statistics", videoID, b.apikey)
	logger("YouTube: API URL exact: %s", apiURL)
	resp, err := http.Get(apiURL)
	if err != nil {
		logger("YouTube: API HTTP GET error: %v", err)
		return "", "", "", "", "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	logger("YouTube: API response status: %s", resp.Status)
	if resp.StatusCode != 200 {
		logger("YouTube: API response body: %s", string(body))
		return "", "", "", "", "", fmt.Errorf("API request failed: %s", resp.Status)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		logger("YouTube: API JSON decode error: %v", err)
		logger("YouTube: API JSON body: %s", string(body))
		return "", "", "", "", "", err
	}
	items, ok := data["items"].([]interface{})
	if !ok || len(items) == 0 {
		logger("YouTube: API no video data found")
		return "", "", "", "", "", fmt.Errorf("no video data found")
	}
	item := items[0].(map[string]interface{})
	snippet, ok := item["snippet"].(map[string]interface{})
	if !ok {
		logger("YouTube: API missing snippet")
		return "", "", "", "", "", fmt.Errorf("missing snippet")
	}
	statistics, ok := item["statistics"].(map[string]interface{})
	if !ok {
		logger("YouTube: API missing statistics")
		return "", "", "", "", "", fmt.Errorf("missing statistics")
	}
	title := "N/A"
	if t, ok := snippet["title"].(string); ok {
		title = t
	}
	uploadDate := "N/A"
	if d, ok := snippet["publishedAt"].(string); ok {
		uploadDate = d
	}
	channelTitle := "N/A"
	if c, ok := snippet["channelTitle"].(string); ok {
		channelTitle = c
	}
	likes := "N/A"
	if l, ok := statistics["likeCount"].(string); ok {
		likes = l
	}
	dislikes := ""
	if d, ok := statistics["dislikeCount"].(string); ok {
		dislikes = d
	}
	logger("YouTube: API success - Title: %s, Likes: %s, Dislikes: %s, Upload: %s, Channel: %s", title, likes, dislikes, uploadDate, channelTitle)
	return title, likes, dislikes, uploadDate, channelTitle, nil
}
