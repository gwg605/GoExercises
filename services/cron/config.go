package main

import (
	"encoding/json"
	"fmt"
	"valerygordeev/go/exercises/libs/base"
	"valerygordeev/go/exercises/libs/cron"
)

type CronConfig struct {
	DBConfig      base.Opts           `json:"db"`
	WebConfig     base.WebConfig      `json:"web"`
	ServiceConfig *cron.ServiceConfig `json:"service"`
}

func NewCronConfig() *CronConfig {
	return &CronConfig{
		DBConfig:      base.Opts{},
		WebConfig:     base.WebConfig{},
		ServiceConfig: cron.NewServiceConfig(),
	}
}

func LoadCronConfig(location string) (*CronConfig, error) {
	expandedLocation := base.ExpandString(location, base.Opts{})

	content, code, err := base.LoadDataFromLocation(expandedLocation, base.Opts{})
	if err != nil {
		return nil, fmt.Errorf("unable to load config from '%s' location. Error=%d/%v", expandedLocation, code, err)
	}

	translatedBody, err := base.TranslateConfig(string(content), base.Opts{})
	if err != nil {
		return nil, fmt.Errorf("unable to translate config '%s'. Error=%v", expandedLocation, err)
	}

	result := NewCronConfig()
	err = json.Unmarshal([]byte(translatedBody), &result)
	if err != nil {
		return nil, fmt.Errorf("unable to load json from '%s' location. Error=%s", expandedLocation, err)
	}

	return result, nil
}
