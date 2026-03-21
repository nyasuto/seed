package render

import (
	"fmt"
	"strings"

	"github.com/nyasuto/seed/core/types"
)

// ANSI color escape sequences.
const (
	Reset   = "\033[0m"
	Bold    = "\033[1m"
	Dim     = "\033[2m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"
	// Brown is approximated via dark yellow (33 without bold).
	Brown = "\033[33m"
)

// ElementColor returns the ANSI color code for the given element.
//
//	Fire=Red, Water=Blue, Wood=Green, Metal=Yellow, Earth=Brown
func ElementColor(e types.Element) string {
	switch e {
	case types.Fire:
		return Red
	case types.Water:
		return Blue
	case types.Wood:
		return Green
	case types.Metal:
		return Yellow
	case types.Earth:
		return Brown
	default:
		return White
	}
}

// Colorize wraps text with the given ANSI color code and a reset suffix.
func Colorize(text, color string) string {
	return color + text + Reset
}

// HPBar returns a colored bar representing current/max HP.
// The bar is exactly width characters wide (excluding brackets).
// Format: [████░░░░░░] where filled portion is colored by ratio.
func HPBar(current, max, width int) string {
	if max <= 0 {
		return "[" + strings.Repeat("░", width) + "]"
	}
	ratio := float64(current) / float64(max)
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(width))

	color := Green
	switch {
	case ratio <= 0.25:
		color = Red
	case ratio <= 0.5:
		color = Yellow
	}

	bar := Colorize(strings.Repeat("█", filled), color) + strings.Repeat("░", width-filled)
	return "[" + bar + "]"
}

// ProgressBar returns a plain bar representing a 0.0–1.0 ratio.
// The bar is exactly width characters wide (excluding brackets).
func ProgressBar(ratio float64, width int) string {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio * float64(width))
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
}

// FormatHP returns a "current/max" HP string with color based on ratio.
func FormatHP(current, max int) string {
	if max <= 0 {
		return "0/0"
	}
	ratio := float64(current) / float64(max)
	color := Green
	switch {
	case ratio <= 0.25:
		color = Red
	case ratio <= 0.5:
		color = Yellow
	}
	return Colorize(fmt.Sprintf("%d/%d", current, max), color)
}

// StripANSI removes ANSI escape sequences from text, returning plain text.
// This is useful for measuring visible width.
func StripANSI(s string) string {
	var out strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\033' {
			// Skip until 'm' (end of ANSI escape).
			for i < len(s) && s[i] != 'm' {
				i++
			}
			if i < len(s) {
				i++ // skip the 'm'
			}
			continue
		}
		out.WriteByte(s[i])
		i++
	}
	return out.String()
}

// VisibleWidth returns the visible character count of a string,
// ignoring ANSI escape sequences.
func VisibleWidth(s string) int {
	return len(StripANSI(s))
}
