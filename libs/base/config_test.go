package base_test

import (
	"testing"
	"valerygordeev/go/exercises/libs/base"
)

func TestTranslateConfig(t *testing.T) {
	base.SetGlobalValue("Y", "YYY")
	translated, err := base.TranslateConfig("{{A}}${Z}-${Y}{{B}}", base.Opts{"Z": "ZZZ"})
	if err != nil {
		t.Fatalf("base.TranslateConfig()=%v", err)
	}

	translatedExpected := "{{A}}ZZZ-YYY{{B}}"
	if translated != translatedExpected {
		t.Fatalf("Wrong value '%s'. Expected '%s'", translated, translatedExpected)
	}

	_, err = base.TranslateConfig("${XXXXX", base.Opts{"Z": "YYY"})
	if err == nil {
		t.Fatalf("base.TranslateConfig(${XXXXX) must return error")
	}
}
