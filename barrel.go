package main

type Barrel interface {
	Name() string
	Enabled() bool
	SetEnabled(bool)
	LoadConfig(*BarrelConfig)
	HandleMessage(*Bot, string, string, string)
	HandleCommand(*Bot, string, string, string, []string) bool
}
