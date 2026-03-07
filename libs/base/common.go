package base

func AbsInt8(x int8) uint8 {
	if x < 0 {
		return uint8(-x)
	}
	return uint8(x)
}

func AbsInt16(x int16) uint16 {
	if x < 0 {
		return uint16(-x)
	}
	return uint16(x)
}

func AbsInt32(x int32) uint32 {
	if x < 0 {
		return uint32(-x)
	}
	return uint32(x)
}

func AbsInt64(x int64) uint64 {
	if x < 0 {
		return uint64(-x)
	}
	return uint64(x)
}

func AbsInt(x int) uint {
	if x < 0 {
		return uint(-x)
	}
	return uint(x)
}
