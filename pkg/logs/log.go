package logs

import (
	"encoding/json"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
)

func GetLogger() *zap.Logger {
	pwd, _ := os.Getwd()
	rawJSON, _ := ioutil.ReadFile(pwd + "/pkg/logs/config.json")
	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	log, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return log
}