package controls

// AxisDirection applies hysteresis so an analog stick emits one navigation
// step when crossing the threshold and must return near center before repeating.
func AxisDirection(value float64, previous int) int {
	if previous != 0 && value > -0.35 && value < 0.35 {
		return 0
	}
	if previous == 0 {
		if value <= -0.65 {
			return -1
		}
		if value >= 0.65 {
			return 1
		}
	}
	return previous
}
