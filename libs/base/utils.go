package base

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"time"
)

const (
	PackagesFolder = "valerygordeev"
)

func TempFileName(prefix, suffix string) string {
	randBytes := make([]byte, 16)
	_, _ = rand.Read(randBytes)
	return filepath.Join(os.TempDir(), prefix+hex.EncodeToString(randBytes)+suffix)
}

func GetRevision() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	}
	return ""
}

func MakeFullAppSpecificPath(path string) string {
	appPackage := GetGlobalValue(VAR_APP_PACKAGE_NAME)
	if appPackage != "" {
		return EnsureTrailingPathSeparator(filepath.Join(path, PackagesFolder, appPackage))
	}
	appName := GetGlobalValue(VAR_APP_NAME)
	if appName != "" {
		return EnsureTrailingPathSeparator(filepath.Join(path, PackagesFolder, appName))
	}
	execName := GetGlobalValue(VAR_EXECUTABLE_NAME)
	return EnsureTrailingPathSeparator(filepath.Join(path, PackagesFolder, filepath.Base(execName)))
}

func InitVars(appPackage string, appName string, useMachineAsRootPath bool) error {
	ex, err := os.Executable()
	if err != nil {
		return err
	}

	SetGlobalValue(VAR_APP_PACKAGE_NAME, appPackage)
	SetGlobalValue(VAR_APP_NAME, appName)
	SetGlobalValue(VAR_EXECUTABLE_NAME, ex)
	SetGlobalValue(VAR_EXECUTABLE_FOLDER_NAME, EnsureTrailingPathSeparator(filepath.Dir(ex)))

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	SetGlobalValue(VAR_CURRENT_FOLDER_NAME, EnsureTrailingPathSeparator(cwd))

	osname := runtime.GOOS
	if osname == "windows" {
		SetGlobalValue(VAR_MACHINE_FOLDER_NAME, MakeFullAppSpecificPath(os.Getenv("ALLUSERSPROFILE")))
		SetGlobalValue(VAR_USER_FOLDER_NAME, MakeFullAppSpecificPath(os.Getenv("USERPROFILE")))
	} else {
		SetGlobalValue(VAR_MACHINE_FOLDER_NAME, MakeFullAppSpecificPath("/var/lib/"))
		SetGlobalValue(VAR_USER_FOLDER_NAME, MakeFullAppSpecificPath(os.Getenv("HOME")))
	}

	var folderName string
	if useMachineAsRootPath {
		folderName = VAR_MACHINE_FOLDER_NAME
	} else {
		folderName = VAR_USER_FOLDER_NAME
	}
	SetGlobalValue(VAR_APP_ROOT_FOLDER_NAME, "%"+folderName+"%")
	SetGlobalValue(VAR_APP_CONFIG_FOLDER_NAME, "%"+folderName+"%configs/")
	SetGlobalValue(VAR_APP_STORE_FOLDER_NAME, "%"+folderName+"%store/")

	return nil
}

func EnsureTrailingPathSeparator(path string) string {
	if path[len(path)-1] == os.PathSeparator {
		return path
	}
	return path + string(os.PathSeparator)
}

func ResolvePath(base string, specified string) string {
	if filepath.IsAbs(specified) {
		return specified
	}
	joined := filepath.Join(base, specified)
	result, err := filepath.Abs(joined)
	if err != nil {
		return joined
	}
	return result
}

type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}

func GenerateTextErrorResponse(w http.ResponseWriter, status_code int, text string, err error) {
	w.Header().Add("Content-Type", "text/plain")
	w.WriteHeader(status_code)
	_, _ = w.Write([]byte(text))
	_, _ = fmt.Fprintf(w, " Error=%s", err)
}
