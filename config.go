// Package main contains the shared bot implementation and configuration model.
package main

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

// NetworkConfig holds IRC server connection details and channel list.
type NetworkConfig struct {
	Server   string   `toml:"server"`
	Port     int      `toml:"port"`
	TLS      bool     `toml:"tls"`
	Password string   `toml:"password,omitempty"`
	Channels []string `toml:"channels"`
}

// BotConfig contains bot identity, admin list, and runtime settings.
type BotConfig struct {
	Nick          string   `toml:"nick"`
	User          string   `toml:"user"`
	RealName      string   `toml:"realname"`
	CommandPrefix string   `toml:"command_prefix"`
	Admins        []string `toml:"admins"`
	LogLevel      string   `toml:"log_level"`
	LogFile       string   `toml:"log_file,omitempty"`
	PidFile       string   `toml:"pidfile,omitempty"`
}

// BarrelConfig describes optional behavior settings for a named barrel.
type BarrelConfig struct {
	Enabled  bool           `toml:"enabled"`
	Apikey   string         `toml:"apikey,omitempty"`
	Settings map[string]any `toml:"settings,omitempty"`
}

// Config is the top-level TOML configuration structure for tbot.
type Config struct {
	Network NetworkConfig            `toml:"network"`
	Bot     BotConfig                `toml:"bot"`
	Barrel  map[string]*BarrelConfig `toml:"barrel"`
}

// ExampleConfig returns a starter configuration with sane defaults.
func ExampleConfig() *Config {
	return &Config{
		Network: NetworkConfig{
			Server:   "irc.example.net",
			Port:     6667,
			TLS:      false,
			Channels: []string{"#example"},
		},
		Bot: BotConfig{
			Nick:          "tbot",
			User:          "tbot",
			RealName:      "tbot IRC bot",
			CommandPrefix: ".",
			Admins:        []string{"tew!~tew@example.net"},
			LogLevel:      "info",
			LogFile:       "",
			PidFile:       "",
		},
		Barrel: map[string]*BarrelConfig{
			"url": {
				Enabled: true,
			},
			"fish": {
				Enabled: true,
			},
			"youtube": {
				Enabled: true,
				Apikey:  "your_youtube_api_key_here",
			},
		},
	}
}

// LoadConfig reads a TOML file from disk and unmarshals it into Config.
// It applies default values for missing bot settings.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := ExampleConfig()
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.Bot.CommandPrefix == "" {
		cfg.Bot.CommandPrefix = "."
	}
	if cfg.Bot.LogLevel == "" {
		cfg.Bot.LogLevel = "info"
	}
	return cfg, nil
}

// Save writes the Config back to disk as TOML.
func (cfg *Config) Save(path string) error {
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
