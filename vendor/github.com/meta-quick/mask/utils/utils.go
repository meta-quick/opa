package utils

import (
	"math/rand"
	"time"
	"unicode"
)

func LeftOverlayString(in string, overlay string, end int) (overlayed string) {
	return OverlayString(in, overlay, 0, end)
}

func RightOverlayString(in string, overlay string, start int) (overlayed string) {
	return OverlayString(in, overlay, start, len(in))
}

func OverlayString(in string, overlay string, start int, end int) (overlayed string) {
	r := []rune(in)
	l := len([]rune(r))

	if l == 0 {
		return ""
	}

	if start < 0 {
		start = 0
	}
	if start > l {
		start = l
	}
	if end < 0 {
		end = 0
	}
	if end > l {
		end = l
	}

	if start > end {
		start, end = end, start
	}

	overlayed = ""
	overlayed += string(r[:start])
	overlayed += overlay
	overlayed += string(r[end:])
	return overlayed
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func CharMapping(c rune, replace rune, fixed bool) rune {
	if unicode.IsDigit(c) {
		return rune('0' + rand.Intn(10))
	} else if unicode.IsLetter(c) {
		if fixed {
			return replace
		}
		if c <= 90 {
			return rune('A' + rand.Intn(26))
		} else {
			return rune('a' + rand.Intn(26))
		}
	}
	return replace
}
