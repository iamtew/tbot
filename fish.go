package main

import (
	"fmt"
	"math/rand"
	"strings"
)

// FishBarrel is a simple example barrel that responds to `fish` commands.
type FishBarrel struct {
	enabled bool
}

// NewFishBarrel constructs the fish barrel plugin.
func NewFishBarrel() *FishBarrel {
	return &FishBarrel{}
}

func (b *FishBarrel) Name() string {
	return "fish"
}

func (b *FishBarrel) Enabled() bool {
	return b.enabled
}

func (b *FishBarrel) SetEnabled(enabled bool) {
	b.enabled = enabled
}

func (b *FishBarrel) LoadConfig(_ *BarrelConfig) {
}

// HandleMessage is a no-op for the fish barrel; it only responds to commands.
func (b *FishBarrel) HandleMessage(_ *Bot, _ string, _ string, _ string) {
}

func (b *FishBarrel) HandleCommand(bot *Bot, channel, nick, command string, args []string) bool {
	if command != "fish" {
		return false
	}
	joke := fishJokes[rand.Intn(len(fishJokes))]
	if len(args) > 0 {
		targets := strings.Join(args, " ")
		bot.sendMessage(channel, fmt.Sprintf("%s: %s", targets, joke))
		return true
	}
	bot.sendMessage(channel, fmt.Sprintf("%s: %s", nick, joke))
	return true
}

var fishJokes = []string{
	"Why don’t fish play basketball? They’re afraid of the net.",
	"What do you call a fish with no eyes? Fsh.",
	"Why are fish so smart? Because they live in schools.",
	"What do you call a fish wearing a crown? King Neptune.",
	"How do fish get to school? By octo-bus.",
}
