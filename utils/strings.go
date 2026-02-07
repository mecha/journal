package utils

import (
	"errors"
	"strconv"
	"strings"

	"github.com/bbrks/wrap"
)

func ParseDayMonthYear(s string) (day, month, year int, err error) {
	day, month, year, err = ParseNumTriplet(s, "/")
	if err != nil {
		return 0, 0, 0, errors.New("invalid date format, must be day/month/year")
	}
	return day, month, year, nil
}

func ParseNumTriplet(s string, sep string) (int, int, int, error) {
	parts := strings.Split(s, sep)
	if len(parts) != 3 {
		return 0, 0, 0, errors.New("invalid format")
	}
	parts = parts[len(parts)-3:]

	one, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, err
	}

	two, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, err
	}

	three, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, 0, 0, err
	}

	return one, two, three, nil
}

func FixedString(s string, maxLength int, pad string) string {
	switch {
	case len(s) > maxLength:
		return s[:max(0, maxLength-1)] + "…"

	case pad == "":
		return s

	default:
		return s + strings.Repeat(pad, maxLength-len(s))
	}
}

func ScrollString(s string, scroll, maxLength int, pad string) string {
	scroll = max(0, scroll)
	if scroll > 0 {
		if scroll < len(s) {
			s = s[scroll:]
		} else {
			s = ""
		}
	}
	return FixedString(s, maxLength, pad)
}

func MaxLength(strings []string) int {
	maxLen := 0
	for _, item := range strings {
		maxLen = max(maxLen, len(item))
	}
	return maxLen
}

func WrapString(s string, n int) []string {
	wrapped := wrap.Wrap(s, n)
	lines := strings.Split(wrapped, "\n")
	return lines
}
