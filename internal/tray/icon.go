package tray

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"zenpomo/assets"
	"zenpomo/internal/core"

	xdraw "golang.org/x/image/draw"
)

var (
	iconWorkScaled  []byte
	iconBreakScaled []byte
	iconPauseScaled []byte
)

func init() {
	iconWorkScaled = renderVectorTomato(24, core.SessionWork, core.StateRunning)
	iconBreakScaled = renderVectorTomato(24, core.SessionShortBreak, core.StateRunning)
	iconPauseScaled = renderVectorTomato(24, core.SessionWork, core.StatePaused)
}

// GetIconForState returns a clean, pixel-perfect 24x24 icon tailored for Linux/Windows tray panels.
func GetIconForState(session core.SessionType, state core.TimerState) []byte {
	if state == core.StatePaused && len(iconPauseScaled) > 0 {
		return iconPauseScaled
	}
	if (session == core.SessionShortBreak || session == core.SessionLongBreak) && len(iconBreakScaled) > 0 {
		return iconBreakScaled
	}
	if len(iconWorkScaled) > 0 {
		return iconWorkScaled
	}
	return assets.IconTomato
}

// renderVectorTomato procedurally draws a crisp, clean, 100% transparent vector tomato.
func renderVectorTomato(size int, session core.SessionType, state core.TimerState) []byte {
	// Draw at 4x resolution and downscale for perfect anti-aliasing
	hiRes := size * 4
	img := image.NewRGBA(image.Rect(0, 0, hiRes, hiRes))

	var bodyColor, bodyShadow, highlightColor, leafColor, stemColor color.RGBA

	if session == core.SessionShortBreak || session == core.SessionLongBreak {
		bodyColor = color.RGBA{0, 180, 140, 255}       // Emerald Green
		bodyShadow = color.RGBA{0, 135, 105, 255}
		highlightColor = color.RGBA{120, 235, 200, 180}
		leafColor = color.RGBA{20, 120, 80, 255}
		stemColor = color.RGBA{15, 90, 60, 255}
	} else if state == core.StatePaused {
		bodyColor = color.RGBA{245, 140, 30, 255}      // Amber Warm
		bodyShadow = color.RGBA{205, 110, 20, 255}
		highlightColor = color.RGBA{255, 195, 90, 180}
		leafColor = color.RGBA{100, 140, 60, 255}
		stemColor = color.RGBA{70, 100, 40, 255}
	} else {
		bodyColor = color.RGBA{235, 50, 45, 255}       // Vibrant Tomato Red
		bodyShadow = color.RGBA{195, 30, 25, 255}
		highlightColor = color.RGBA{255, 130, 120, 180}
		leafColor = color.RGBA{46, 175, 80, 255}       // Fresh Leaf Green
		stemColor = color.RGBA{30, 130, 60, 255}
	}

	cx := float64(hiRes) / 2.0
	cy := float64(hiRes) * 0.54
	rx := float64(hiRes) * 0.38
	ry := float64(hiRes) * 0.34

	// Draw smooth tomato body
	for y := 0; y < hiRes; y++ {
		for x := 0; x < hiRes; x++ {
			dx := (float64(x) - cx) / rx
			dy := (float64(y) - cy) / ry
			dist := dx*dx + dy*dy
			if dist <= 1.0 {
				edgeDist := 1.0 - math.Sqrt(dist)
				alpha := 1.0
				if edgeDist < 0.08 {
					alpha = edgeDist / 0.08
				}

				c := bodyColor
				if dist > 0.75 {
					c = bodyShadow
				}

				// Specular highlight on top-left
				hx := (float64(x) - (cx - rx*0.4)) / (rx * 0.35)
				hy := (float64(y) - (cy - ry*0.4)) / (ry * 0.25)
				if hx*hx+hy*hy <= 1.0 {
					c = blendColor(c, highlightColor)
				}

				img.SetRGBA(x, y, color.RGBA{
					R: c.R,
					G: c.G,
					B: c.B,
					A: uint8(float64(c.A) * alpha),
				})
			}
		}
	}

	// Draw leaves
	leafY := cy - ry*0.75
	leafLen := float64(hiRes) * 0.22
	angles := []float64{-75, -35, 0, 35, 75}
	for _, deg := range angles {
		rad := deg * math.Pi / 180.0
		lx := cx + math.Sin(rad)*leafLen
		ly := leafY - math.Cos(rad)*leafLen*0.6
		drawThickLine(img, cx, leafY, lx, ly, float64(hiRes)*0.045, leafColor)
	}

	// Draw center stem
	drawThickLine(img, cx, leafY, cx-float64(hiRes)*0.02, leafY-float64(hiRes)*0.18, float64(hiRes)*0.065, stemColor)

	// Downsample with Catmull-Rom for sharp, crisp anti-aliasing
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return assets.IconTomato
	}
	return buf.Bytes()
}

func blendColor(base, overlay color.RGBA) color.RGBA {
	a := float64(overlay.A) / 255.0
	return color.RGBA{
		R: uint8(float64(base.R)*(1-a) + float64(overlay.R)*a),
		G: uint8(float64(base.G)*(1-a) + float64(overlay.G)*a),
		B: uint8(float64(base.B)*(1-a) + float64(overlay.B)*a),
		A: base.A,
	}
}

func drawThickLine(img *image.RGBA, x0, y0, x1, y1, width float64, c color.RGBA) {
	dx := x1 - x0
	dy := y1 - y0
	length := math.Hypot(dx, dy)
	if length == 0 {
		return
	}
	steps := int(length * 2)
	radius := width / 2.0

	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		cx := x0 + dx*t
		cy := y0 + dy*t

		minX := int(cx - radius)
		maxX := int(cx + radius)
		minY := int(cy - radius)
		maxY := int(cy + radius)

		for y := minY; y <= maxY; y++ {
			for x := minX; x <= maxX; x++ {
				if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
					dist := math.Hypot(float64(x)-cx, float64(y)-cy)
					if dist <= radius {
						alpha := 1.0
						if radius-dist < 1.0 {
							alpha = radius - dist
						}
						img.SetRGBA(x, y, color.RGBA{
							R: c.R,
							G: c.G,
							B: c.B,
							A: uint8(float64(c.A) * alpha),
						})
					}
				}
			}
		}
	}
}
