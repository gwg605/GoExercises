package base

import (
	"io"
	"os"
	"strings"

	"github.com/valyala/fasttemplate"
)

var (
	contextVars map[string]string = make(map[string]string)
)

func SetGlobalValue(key string, val string) {
	contextVars[key] = val
}

func GetGlobalValue(key string) string {
	val, ok := contextVars[key]
	if ok {
		return val
	}
	return os.Getenv(key)
}

func GetGlobalValues() map[string]string {
	return contextVars
}

func ExpandString(val string, options Opts) string {
	if !strings.Contains(val, "%") {
		return val
	}

	t, err := fasttemplate.NewTemplate(val, "%", "%")
	if err != nil {
		return ""
	}
	result := t.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		var v string
		v, ok := options[tag]
		if !ok {
			v = GetGlobalValue(tag)
		}
		ve := ExpandString(v, options)
		return w.Write([]byte(ve))
	})

	return result
}
