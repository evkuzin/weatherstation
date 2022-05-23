package impl

func tokHPa(v int64) float64 {
	precision := v % 1000000000
	base := int(v / 1000000000)
	if precision > 500000000 {
		base++
	}
	frac := base % 1000
	base = base / 1000
	return float64(base) + float64(frac)/1000
}
