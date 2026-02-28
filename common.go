package main

type SignedInt interface {
	int | int8 | int16 | int32 | int64
}

func AbsInt[T SignedInt](x T) T {
	if x < 0 {
		return -x
	}
	return x
}
