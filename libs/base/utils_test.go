package base_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
	"valerygordeev/go/exercises/libs/base"
)

func TestInitVars(t *testing.T) {
	err := base.InitVars("", "test", true)
	if err != nil {
		t.Fatalf("InitVars()=%v", err)
	}
	osname := runtime.GOOS
	if osname != "windows" {
		rootFolder := base.ExpandString(base.GetGlobalValue(base.VAR_APP_ROOT_FOLDER_NAME), base.Opts{})
		rootFolderExpected := "/var/lib/valerygordeev/test/"
		if rootFolder != rootFolderExpected {
			t.Errorf("Wrong rootFolder '%s'. Expected '%s'", rootFolder, rootFolderExpected)
		}

		machineFolder := base.ExpandString(base.GetGlobalValue(base.VAR_MACHINE_FOLDER_NAME), base.Opts{})
		machineFolderExpected := "/var/lib/valerygordeev/test/"
		if machineFolder != machineFolderExpected {
			t.Errorf("Wrong machineFolder '%s'. Expected '%s'", machineFolder, machineFolderExpected)
		}

		userFolder := base.ExpandString(base.GetGlobalValue(base.VAR_USER_FOLDER_NAME), base.Opts{})
		userFolderExpected := base.EnsureTrailingPathSeparator(filepath.Join(os.Getenv("HOME"), "valerygordeev/test/"))
		if userFolder != userFolderExpected {
			t.Errorf("Wrong userFolder '%s'. Expected '%s'", userFolder, userFolderExpected)
		}
	}
}

func TestEnsureTrailingPathSeparator(t *testing.T) {
	folderExpected := "/folder/"
	folder := base.EnsureTrailingPathSeparator("/folder/")
	if folder != folderExpected {
		t.Errorf("Wrong value '%s'. Expected '%s'", folder, folderExpected)
	}
	folder = base.EnsureTrailingPathSeparator("/folder")
	if folder != folderExpected {
		t.Errorf("Wrong value '%s'. Expected '%s'", folder, folderExpected)
	}
}

func TestDurationMarshalUnmarshal(t *testing.T) {
	var dr base.Duration

	d := base.Duration{time.Second}
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("json.Marshal(d)=%v", err)
	}

	err = json.Unmarshal(data, &dr)
	if err != nil {
		t.Fatalf("json.Unmarshal(data, &dr)=%v", err)
	}
	if d != dr {
		t.Fatalf("Wrong value %v. Expected %v", dr, d)
	}

	err = json.Unmarshal([]byte("\"5s\""), &dr)
	if err != nil {
		t.Fatalf("json.Unmarshal(data, &dr)=%v", err)
	}
	if dr.Duration != 5*time.Second {
		t.Fatalf("Wrong value %v. Expected %v", dr, d)
	}

	err = json.Unmarshal([]byte("1000000.0"), &dr)
	if err != nil {
		t.Fatalf("json.Unmarshal(data, &dr)=%v", err)
	}
	if dr.Duration != time.Millisecond {
		t.Fatalf("Wrong value %v. Expected %v", dr, d)
	}

	err = json.Unmarshal([]byte("1000000"), &dr)
	if err != nil {
		t.Fatalf("json.Unmarshal(data, &dr)=%v", err)
	}
	if dr.Duration != time.Millisecond {
		t.Fatalf("Wrong value %v. Expected %v", dr, d)
	}

	err = json.Unmarshal([]byte("\"\""), &dr)
	if err == nil {
		t.Fatalf("json.Unmarshal(nil, &dr) must return error")
	}

	err = json.Unmarshal([]byte("{}"), &dr)
	if err == nil {
		t.Fatalf("json.Unmarshal(nil, &dr) must return error")
	}
}
