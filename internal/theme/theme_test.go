package theme_test

import (
	"testing"
	"zenpomo/internal/theme"
)

func TestGetTheme(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"gruvbox", "gruvbox"},
		{"catppuccin", "catppuccin"},
		{"tokyonight", "tokyonight"},
		{"nord", "nord"},
		{"dracula", "dracula"},
		{"rosepine", "rosepine"},
		{"monochrome", "monochrome"},
		{"unknown", "gruvbox"},
		{"", "gruvbox"},
	}

	for _, tt := range tests {
		p := theme.GetTheme(tt.input)
		if p.Name != tt.expected {
			t.Errorf("GetTheme(%q) = %q; want %q", tt.input, p.Name, tt.expected)
		}
	}
}

func TestNextTheme(t *testing.T) {
	next := theme.NextTheme("gruvbox")
	if next != "catppuccin" {
		t.Errorf("NextTheme(gruvbox) = %q; want catppuccin", next)
	}

	last := theme.AvailableThemes[len(theme.AvailableThemes)-1]
	cycled := theme.NextTheme(last)
	if cycled != theme.AvailableThemes[0] {
		t.Errorf("NextTheme(%q) = %q; want %q", last, cycled, theme.AvailableThemes[0])
	}
}
