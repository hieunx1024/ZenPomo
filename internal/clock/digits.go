package clock

import "fmt"

// DigitHeight is the number of vertical lines for each big ASCII number.
const DigitHeight = 5

// digitMaps defines 5-line tall ASCII block numbers for 0-9 and separator ':'.
var digitMaps = map[rune][]string{
	'0': {
		"██████",
		"██  ██",
		"██  ██",
		"██  ██",
		"██████",
	},
	'1': {
		"  ████",
		"    ██",
		"    ██",
		"    ██",
		"  ████",
	},
	'2': {
		"██████",
		"    ██",
		"██████",
		"██    ",
		"██████",
	},
	'3': {
		"██████",
		"    ██",
		"██████",
		"    ██",
		"██████",
	},
	'4': {
		"██  ██",
		"██  ██",
		"██████",
		"    ██",
		"    ██",
	},
	'5': {
		"██████",
		"██    ",
		"██████",
		"    ██",
		"██████",
	},
	'6': {
		"██████",
		"██    ",
		"██████",
		"██  ██",
		"██████",
	},
	'7': {
		"██████",
		"    ██",
		"    ██",
		"    ██",
		"    ██",
	},
	'8': {
		"██████",
		"██  ██",
		"██████",
		"██  ██",
		"██████",
	},
	'9': {
		"██████",
		"██  ██",
		"██████",
		"    ██",
		"██████",
	},
	':': {
		"  ",
		"██",
		"  ",
		"██",
		"  ",
	},
	' ': {
		"  ",
		"  ",
		"  ",
		"  ",
		"  ",
	},
}

// RenderTime returns a 5-line slice of big ASCII text representing "MM:SS".
func RenderTime(totalSeconds int) []string {
	if totalSeconds < 0 {
		totalSeconds = 0
	}
	mins := totalSeconds / 60
	secs := totalSeconds % 60
	timeStr := fmt.Sprintf("%02d:%02d", mins, secs)

	lines := make([]string, DigitHeight)
	for _, char := range timeStr {
		glyph, ok := digitMaps[char]
		if !ok {
			glyph = digitMaps[' ']
		}
		for i := 0; i < DigitHeight; i++ {
			if len(lines[i]) > 0 {
				lines[i] += " "
			}
			lines[i] += glyph[i]
		}
	}
	return lines
}
