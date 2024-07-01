package jstat

func icon(icons []string, max, val float64) string {
	var index, iconsLen int

	iconsLen = len(icons)
	if iconsLen == 0 {
		return ""
	}

	index = int(float64(iconsLen) / max * val)
	if index >= iconsLen {
		return icons[iconsLen-1]
	}

	return icons[index]
}
