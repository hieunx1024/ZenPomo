package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Palette defines the semantic colors for the ZenPomo TUI.
type Palette struct {
	Name         string
	Work         lipgloss.Color
	Break        lipgloss.Color
	LongBreak    lipgloss.Color
	Border       lipgloss.Color
	BorderActive lipgloss.Color
	Text         lipgloss.Color
	TextDim      lipgloss.Color
	Accent       lipgloss.Color
	Highlight    lipgloss.Color
	TabActiveBg  lipgloss.Color
	TabActiveFg  lipgloss.Color
}

// Built-in terminal developer themes (Zero AI Neon, pure clean unix aesthetics).
var (
	// Gruvbox Dark (Default)
	Gruvbox = Palette{
		Name:         "gruvbox",
		Work:         lipgloss.Color("#FB4934"), // Terracotta red
		Break:        lipgloss.Color("#8EC07C"), // Aqua green
		LongBreak:    lipgloss.Color("#83A598"), // Slate blue
		Border:       lipgloss.Color("#504945"), // Charcoal
		BorderActive: lipgloss.Color("#A89984"), // Warm gray
		Text:         lipgloss.Color("#EBDBB2"), // Warm cream
		TextDim:      lipgloss.Color("#928374"), // Muted stone
		Accent:       lipgloss.Color("#FABD2F"), // Warm amber
		Highlight:    lipgloss.Color("#D79921"), // Deep amber
		TabActiveBg:  lipgloss.Color("#3C3836"), // Dark medium
		TabActiveFg:  lipgloss.Color("#EBDBB2"),
	}

	// Catppuccin Mocha
	Catppuccin = Palette{
		Name:         "catppuccin",
		Work:         lipgloss.Color("#F38BA8"), // Red
		Break:        lipgloss.Color("#A6E3A1"), // Green
		LongBreak:    lipgloss.Color("#89B4FA"), // Blue
		Border:       lipgloss.Color("#45475A"), // Surface1
		BorderActive: lipgloss.Color("#BAC2DE"), // Subtext1
		Text:         lipgloss.Color("#CDD6F4"), // Text
		TextDim:      lipgloss.Color("#6C7086"), // Overlay0
		Accent:       lipgloss.Color("#F9E2AF"), // Yellow
		Highlight:    lipgloss.Color("#FAB387"), // Peach
		TabActiveBg:  lipgloss.Color("#313244"), // Surface0
		TabActiveFg:  lipgloss.Color("#CDD6F4"),
	}

	// Tokyo Night
	TokyoNight = Palette{
		Name:         "tokyonight",
		Work:         lipgloss.Color("#F7768E"), // Red
		Break:        lipgloss.Color("#9ECE6A"), // Green
		LongBreak:    lipgloss.Color("#7AA2F7"), // Blue
		Border:       lipgloss.Color("#292E42"), // Dark border
		BorderActive: lipgloss.Color("#7AA2F7"), // Blue highlight
		Text:         lipgloss.Color("#C0CAF5"), // White/blue text
		TextDim:      lipgloss.Color("#565F89"), // Comment dark
		Accent:       lipgloss.Color("#E0AF68"), // Yellow
		Highlight:    lipgloss.Color("#BB9AF7"), // Purple/Lavender
		TabActiveBg:  lipgloss.Color("#1F2335"), // Dark bg
		TabActiveFg:  lipgloss.Color("#C0CAF5"),
	}

	// Nord
	Nord = Palette{
		Name:         "nord",
		Work:         lipgloss.Color("#BF616A"), // Nord11 Red
		Break:        lipgloss.Color("#A3BE8C"), // Nord14 Green
		LongBreak:    lipgloss.Color("#88C0D0"), // Nord8 Frost cyan
		Border:       lipgloss.Color("#3B4252"), // Nord1
		BorderActive: lipgloss.Color("#81A1C1"), // Nord9 Frost blue
		Text:         lipgloss.Color("#ECEFF4"), // Nord6 Snow storm
		TextDim:      lipgloss.Color("#4C566A"), // Nord3 Polar night
		Accent:       lipgloss.Color("#EBCB8B"), // Nord13 Yellow
		Highlight:    lipgloss.Color("#D08770"), // Nord12 Orange
		TabActiveBg:  lipgloss.Color("#3B4252"),
		TabActiveFg:  lipgloss.Color("#ECEFF4"),
	}

	// Dracula
	Dracula = Palette{
		Name:         "dracula",
		Work:         lipgloss.Color("#FF5555"), // Red
		Break:        lipgloss.Color("#50FA7B"), // Green
		LongBreak:    lipgloss.Color("#8BE9FD"), // Cyan
		Border:       lipgloss.Color("#44475A"), // Current line
		BorderActive: lipgloss.Color("#BD93F9"), // Purple
		Text:         lipgloss.Color("#F8F8F2"), // Foreground
		TextDim:      lipgloss.Color("#6272A4"), // Comment
		Accent:       lipgloss.Color("#F1FA8C"), // Yellow
		Highlight:    lipgloss.Color("#FFB86C"), // Orange
		TabActiveBg:  lipgloss.Color("#44475A"),
		TabActiveFg:  lipgloss.Color("#F8F8F2"),
	}

	// Rose Pine
	RosePine = Palette{
		Name:         "rosepine",
		Work:         lipgloss.Color("#EB6F92"), // Love
		Break:        lipgloss.Color("#9CCFD8"), // Foam
		LongBreak:    lipgloss.Color("#31748F"), // Pine
		Border:       lipgloss.Color("#26233A"), // Highlight Med
		BorderActive: lipgloss.Color("#EBBCBA"), // Rose
		Text:         lipgloss.Color("#E0DEF4"), // Text
		TextDim:      lipgloss.Color("#6E6A86"), // Muted
		Accent:       lipgloss.Color("#F6C177"), // Gold
		Highlight:    lipgloss.Color("#C4A7E7"), // Iris
		TabActiveBg:  lipgloss.Color("#2A283E"),
		TabActiveFg:  lipgloss.Color("#E0DEF4"),
	}

	// Monochrome (High Contrast / E-Ink)
	Monochrome = Palette{
		Name:         "monochrome",
		Work:         lipgloss.Color("#FFFFFF"),
		Break:        lipgloss.Color("#CCCCCC"),
		LongBreak:    lipgloss.Color("#AAAAAA"),
		Border:       lipgloss.Color("#666666"),
		BorderActive: lipgloss.Color("#FFFFFF"),
		Text:         lipgloss.Color("#FFFFFF"),
		TextDim:      lipgloss.Color("#777777"),
		Accent:       lipgloss.Color("#FFFFFF"),
		Highlight:    lipgloss.Color("#DDDDDD"),
		TabActiveBg:  lipgloss.Color("#333333"),
		TabActiveFg:  lipgloss.Color("#FFFFFF"),
	}
)

// AvailableThemes lists all supported theme keys in display order.
var AvailableThemes = []string{
	"gruvbox",
	"catppuccin",
	"tokyonight",
	"nord",
	"dracula",
	"rosepine",
	"monochrome",
}

// GetTheme returns the palette for the specified theme name, defaulting to Gruvbox.
func GetTheme(name string) Palette {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "catppuccin", "mocha":
		return Catppuccin
	case "tokyonight", "tokyo-night":
		return TokyoNight
	case "nord":
		return Nord
	case "dracula":
		return Dracula
	case "rosepine", "rose-pine":
		return RosePine
	case "monochrome", "e-ink", "bw":
		return Monochrome
	case "gruvbox":
		fallthrough
	default:
		return Gruvbox
	}
}

// NextTheme cycles to the next theme in the available list.
func NextTheme(current string) string {
	cur := strings.ToLower(strings.TrimSpace(current))
	for i, t := range AvailableThemes {
		if t == cur {
			return AvailableThemes[(i+1)%len(AvailableThemes)]
		}
	}
	return AvailableThemes[0]
}
