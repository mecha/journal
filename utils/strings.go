package utils

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/bbrks/wrap"
)

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

func ParseDayMonthYear(s string) (time.Time, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 3 {
		return time.Time{}, errors.New("invalid date: must be <day>/<month>/<year>")
	}

	nums := [3]int{0, 0, 0}
	for i, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			return time.Time{}, errors.New("invalid date: \"" + part + "\" is not a number")
		}
		nums[i] = num
	}

	return time.Date(nums[2], time.Month(nums[1]), nums[0], 0, 0, 0, 0, time.Local), nil
}
