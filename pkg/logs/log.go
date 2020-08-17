package logs

import (
	"encoding/json"
	"go.uber.org/zap"
	"io/ioutil"
)
var Log *zap.Logger

func init() {
	rawJSON, _ := ioutil.ReadFile("./config.json")
	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	Log, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	defer Log.Sync()
}