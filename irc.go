package main

import (
	"strings"
)

type IRCMessage struct {
	Raw      string
	Prefix   string
	Nick     string
	User     string
	Host     string
	Command  string
	Params   []string
	Trailing string
}

// ParseIRCLine parses a raw IRC protocol line into structured fields.
func ParseIRCLine(line string) IRCMessage {
	msg := IRCMessage{Raw: line}
	cursor := 0
	if strings.HasPrefix(line, ":") {
		next := strings.IndexByte(line, ' ')
		if next == -1 {
			return msg
		}
		msg.Prefix = line[1:next]
		cursor = next + 1
	}

	if msg.Prefix != "" {
		if bang := strings.IndexByte(msg.Prefix, '!'); bang != -1 {
			msg.Nick = msg.Prefix[:bang]
			at := strings.IndexByte(msg.Prefix[bang+1:], '@')
			if at != -1 {
				msg.User = msg.Prefix[bang+1 : bang+1+at]
				msg.Host = msg.Prefix[bang+1+at+1:]
			}
		} else {
			msg.Nick = msg.Prefix
		}
	}

	if cursor >= len(line) {
		return msg
	}

	after := line[cursor:]
	trailingIndex := strings.Index(after, " :")
	if trailingIndex != -1 {
		msg.Trailing = after[trailingIndex+2:]
		after = strings.TrimSpace(after[:trailingIndex])
	}

	parts := strings.Fields(after)
	if len(parts) == 0 {
		return msg
	}
	msg.Command = parts[0]
	if len(parts) > 1 {
		msg.Params = parts[1:]
	}
	return msg
}

// TruePrefix strips the leading colon from an IRC prefix, if present.
func TruePrefix(prefix string) string {
	if strings.HasPrefix(prefix, ":") {
		return prefix[1:]
	}
	return prefix
}
