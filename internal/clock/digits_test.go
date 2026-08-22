package clock

import (
	"testing"
)

func TestRenderTime(t *testing.T) {
	lines := RenderTime(25 * 60)
	if len(lines) != DigitHeight {
		t.Fatalf("expected %d lines, got %d", DigitHeight, len(lines))
	}
	for i, l := range lines {
		if len(l) == 0 {
			t.Errorf("line %d is empty", i)
		}
	}
}
