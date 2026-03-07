package base

import (
	"strconv"
	"time"
)

type Opts map[string]string

func (o *Opts) GetString(name string, def_value string) string {
	val, ok := (*o)[name]
	if ok {
		return val
	}
	return def_value
}

func (o *Opts) GetInt(name string, def_value int) int {
	val, ok := (*o)[name]
	if ok {
		res, _ := strconv.Atoi(val)
		return res
	}
	return def_value
}

func (o *Opts) GetInt64(name string, def_value int64) int64 {
	val, ok := (*o)[name]
	if ok {
		res, _ := strconv.ParseInt(val, 10, 64)
		return res
	}
	return def_value
}

func (o *Opts) GetBool(name string, def_value bool) bool {
	val, ok := (*o)[name]
	if ok {
		return val == "true" || val == "1"
	}
	return def_value
}

func (o *Opts) GetDuration(name string, def_value time.Duration) time.Duration {
	val, ok := (*o)[name]
	if ok {
		res, _ := time.ParseDuration(val)
		return res
	}
	return def_value
}

func (o *Opts) SetDefaultString(name string, def_value string) {
	_, ok := (*o)[name]
	if !ok {
		(*o)[name] = def_value
	}
}

func MergeOpts(base Opts, extra Opts) Opts {
	result := Opts{}
	for key, val := range base {
		result[key] = val
	}
	for key, val := range extra {
		result[key] = val
	}
	return result
}

func (o *Opts) Expand(Opts Opts) {
	for k, v := range Opts {
		(*o)[k] = v
	}
}
