package main

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type NetworkConfig struct {
	Server   string   `toml:"server"`
	Port     int      `toml:"port"`
	TLS      bool     `toml:"tls"`
	Password string   `toml:"password,omitempty"`
	Channels []string `toml:"channels"`
}

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

type BarrelConfig struct {
	Enabled  bool           `toml:"enabled"`
	Settings map[string]any `toml:"settings,omitempty"`
}

type Config struct {
	Network NetworkConfig            `toml:"network"`
	Bot     BotConfig                `toml:"bot"`
	Barrels map[string]*BarrelConfig `toml:"barrels"`
}

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
		Barrels: map[string]*BarrelConfig{
			"url": {
				Enabled: true,
			},
			"fish": {
				Enabled: true,
			},
		},
	}
}

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

func (cfg *Config) Save(path string) error {
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
