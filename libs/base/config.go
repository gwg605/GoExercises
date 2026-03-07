package base

import (
	"io"

	"github.com/valyala/fasttemplate"
)

type WebConfig struct {
	Endpoint       string `json:"endpoint"`
	BindingAddress string `json:"binding_address"`
}

func TranslateConfig(templateValue string, overrides Opts) (string, error) {
	tmpl, err := fasttemplate.NewTemplate(templateValue, "${", "}")
	if err != nil {
		return "", err
	}
	result := tmpl.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		var v string
		v, ok := overrides[tag]
		if !ok {
			v = GetGlobalValue(tag)
		}
		return w.Write([]byte(v))
	})
	return result, nil
}
