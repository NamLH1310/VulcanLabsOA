package mathutil

func AbsDiff[T int | uint | int8 | uint8 | int16 | uint16 | int32 | uint32 | int64 | uint64](x, y T) T {
	if x < y {
		return y - x
	}
	return x - y
}
