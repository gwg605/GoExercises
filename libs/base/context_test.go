package base_test

import (
	"testing"
	"valerygordeev/go/exercises/libs/base"
)

func TestGlobalContext(t *testing.T) {
	base.SetGlobalValue("key1", "key#1_value")

	val0 := base.GetGlobalValue("key1")
	if val0 != "key#1_value" {
		t.Errorf("val0 has wrong value (%s)", val0)
	}

	val1 := base.ExpandString("Before:%key1%:after", base.Opts{})
	if val1 != "Before:key#1_value:after" {
		t.Errorf("val1 has wrong value (%s)", val1)
	}

	val2 := base.ExpandString("Before:%key2%:after", base.Opts{})
	if val2 != "Before::after" {
		t.Errorf("val2 has wrong value (%s)", val2)
	}

	val3 := base.ExpandString("Before:%%:after", base.Opts{})
	if val3 != "Before::after" {
		t.Errorf("val3 has wrong value (%s)", val3)
	}

	val4 := base.ExpandString("ZZZZ%YYYY", base.Opts{})
	if val4 != "" {
		t.Errorf("val4 has wrong value (%s)", val4)
	}

	vars := base.GetGlobalValues()
	if len(vars) == 0 {
		t.Errorf("Wrong vars count %d", len(vars))
	}
}
