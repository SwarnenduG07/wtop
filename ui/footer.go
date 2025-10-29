package ui

import "fmt"

func (d *Dashboard) updateFooter(snap *snapshot, rates netRates) {
	lineOne := "[::b]F1[-] Help  [::b]/[-] Filter  [::b]s[-] Sort  [::b]t[-] Theme  [::b]↑↓[-] Scroll  [::b]q[-] Quit"

	themeLabel := "Dark"
	if !d.themeIsDark {
		themeLabel = "Light"
	}

	parts := []string{
		fmt.Sprintf("Refresh %.0fs", d.refreshInterval.Seconds()),
		fmt.Sprintf("Theme %s", themeLabel),
		fmt.Sprintf("Sort %s", d.sortMode.String()),
	}

	if snap != nil {
		parts = append(parts, fmt.Sprintf("Tasks %d", snap.ProcessSummary.Total))
		if snap.Memory != nil {
			parts = append(parts, fmt.Sprintf("Mem %.1f%%", snap.Memory.UsedPercent))
		}
	}

	if !rates.Valid && d.lastRates.Valid {
		rates = d.lastRates
	}

	if rates.Valid {
		parts = append(parts, fmt.Sprintf("Net %s ↑ %s ↓", formatBytesPerSec(rates.Up), formatBytesPerSec(rates.Down)))
	}

	lineTwo := joinWithSpacing(parts)
	d.footer.SetText(lineOne + "\n" + lineTwo)
}
