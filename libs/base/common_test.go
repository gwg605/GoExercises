package base_test

import (
	"math"
	"testing"
	"valerygordeev/go/exercises/libs/base"
)

func TestCommonBasic(t *testing.T) {
	u8 := base.AbsInt8(int8(0))
	if u8 != 0 {
		t.Fatalf("Wrong value %d. Expected 0", u8)
	}
	u8 = base.AbsInt8(int8(math.MinInt8))
	if u8 != uint8(-math.MinInt8) {
		t.Fatalf("Wrong value %d. Expected %d", u8, uint8(-math.MinInt8))
	}

	u16 := base.AbsInt16(int16(0))
	if u16 != 0 {
		t.Fatalf("Wrong value %d. Expected 0", u16)
	}
	u16 = base.AbsInt16(int16(math.MinInt16))
	if u16 != uint16(-math.MinInt16) {
		t.Fatalf("Wrong value %d. Expected %d", u16, uint16(-math.MinInt16))
	}

	u32 := base.AbsInt32(int32(0))
	if u32 != 0 {
		t.Fatalf("Wrong value %d. Expected 0", u32)
	}
	u32 = base.AbsInt32(int32(math.MinInt32))
	if u32 != uint32(-math.MinInt32) {
		t.Fatalf("Wrong value %d. Expected %d", u32, uint32(-math.MinInt32))
	}

	u64 := base.AbsInt64(int64(0))
	if u64 != 0 {
		t.Fatalf("Wrong value %d. Expected 0", u64)
	}
	u64 = base.AbsInt64(int64(math.MinInt64))
	if u64 != uint64(-math.MinInt64) {
		t.Fatalf("Wrong value %d. Expected %d", u64, uint64(-math.MinInt64))
	}

	u := base.AbsInt(int(0))
	if u != 0 {
		t.Fatalf("Wrong value %d. Expected 0", u)
	}
	u = base.AbsInt(int(math.MinInt))
	if u != uint(-math.MinInt) {
		t.Fatalf("Wrong value %d. Expected %d", u, uint(-math.MinInt))
	}
}
