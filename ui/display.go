package ui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

func colorTag(color tcell.Color) string {
	r, g, b := color.RGB()
	return fmt.Sprintf("[#%02x%02x%02x]", r, g, b)
}

func resetTag() string {
	return "[-:-:-]"
}

func usageColor(percent float64) tcell.Color {
	switch {
	case percent >= 85:
		return tcell.ColorIndianRed
	case percent >= 65:
		return tcell.ColorYellow
	default:
		return tcell.ColorGreen
	}
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func renderUsageBar(percent float64, width int) string {
	width = clampInt(width, 6, 60)
	filled := int(math.Round(percent / 100 * float64(width)))
	if filled > width {
		filled = width
	}

	var b strings.Builder
	fillColor := usageColor(percent)
	b.WriteString(colorTag(tcell.ColorDarkSlateGray))
	b.WriteRune('[')
	b.WriteString(resetTag())
	for i := 0; i < width; i++ {
		if i < filled {
			b.WriteString(colorTag(fillColor))
			b.WriteRune('█')
		} else {
			b.WriteString(colorTag(tcell.ColorGray))
			b.WriteRune(' ')
		}
	}
	b.WriteString(resetTag())
	b.WriteString(colorTag(tcell.ColorDarkSlateGray))
	b.WriteRune(']')
	b.WriteString(resetTag())
	b.WriteRune(' ')
	b.WriteString(colorTag(fillColor))
	b.WriteString(fmt.Sprintf("%5.1f%%", percent))
	b.WriteString(resetTag())
	return b.String()
}

type sparkHistory struct {
	values []float64
	maxLen int
}

func newSparkHistory(size int) *sparkHistory {
	if size <= 0 {
		size = 1
	}
	return &sparkHistory{values: make([]float64, 0, size), maxLen: size}
}

func (s *sparkHistory) Push(v float64) {
	if s == nil || s.maxLen <= 0 {
		return
	}
	if len(s.values) == s.maxLen {
		copy(s.values, s.values[1:])
		s.values[len(s.values)-1] = v
		return
	}
	s.values = append(s.values, v)
}

func (s *sparkHistory) Series() []float64 {
	if s == nil {
		return nil
	}
	return s.values
}

func renderSparkline(series []float64, width int) string {
	width = clampInt(width, 4, 80)
	if width == 0 {
		return ""
	}
	if len(series) == 0 {
		return colorTag(tcell.ColorGray) + strings.Repeat("·", width) + resetTag()
	}

	maxVal := 0.0
	for _, v := range series {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal <= 0 {
		return colorTag(tcell.ColorGray) + strings.Repeat("·", width) + resetTag()
	}

	blocks := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	var samples []float64
	if len(series) > width {
		step := float64(len(series)) / float64(width)
		samples = make([]float64, width)
		for i := 0; i < width; i++ {
			start := int(math.Floor(float64(i) * step))
			if start < 0 {
				start = 0
			}
			end := int(math.Floor(float64(i+1) * step))
			if end <= start {
				end = start + 1
			}
			if end > len(series) {
				end = len(series)
			}
			maxSegment := 0.0
			for _, v := range series[start:end] {
				if v > maxSegment {
					maxSegment = v
				}
			}
			samples[i] = maxSegment
		}
	} else {
		samples = make([]float64, width)
		pad := width - len(series)
		for i := 0; i < pad; i++ {
			samples[i] = 0
		}
		copy(samples[pad:], series)
	}

	var b strings.Builder
	for _, v := range samples {
		if v <= 0 {
			b.WriteString(colorTag(tcell.ColorGray))
			b.WriteRune('·')
			continue
		}
		percent := 100 * (v / maxVal)
		idx := int(percent / (100 / float64(len(blocks))))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}
		b.WriteString(colorTag(usageColor(percent)))
		b.WriteRune(blocks[idx])
	}
	b.WriteString(resetTag())
	return b.String()
}

func formatBytes(value float64) string {
	units := []string{"B", "K", "M", "G", "T", "P"}
	idx := 0
	for value >= 1024 && idx < len(units)-1 {
		value /= 1024
		idx++
	}
	if value >= 100 || idx == 0 {
		return fmt.Sprintf("%.0f%s", value, units[idx])
	}
	return fmt.Sprintf("%.1f%s", value, units[idx])
}

func formatBytesPerSec(value float64) string {
	if value <= 0 {
		return "0B/s"
	}
	return fmt.Sprintf("%s/s", formatBytes(value))
}

func formatUptime(d time.Duration) string {
	if d < time.Minute {
		return d.Truncate(time.Second).String()
	}
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

func formatProcessRuntime(seconds int64) string {
	if seconds < 60 {
		return fmt.Sprintf("0:%02d", seconds)
	}
	minutes := seconds / 60
	secs := seconds % 60
	if minutes < 60 {
		return fmt.Sprintf("%d:%02d", minutes, secs)
	}
	hours := minutes / 60
	mins := minutes % 60
	return fmt.Sprintf("%d:%02d:%02d", hours, mins, secs)
}

func joinWithSpacing(parts []string) string {
	clean := parts[:0]
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}
		clean = append(clean, part)
	}
	return strings.Join(clean, "  ")
}
