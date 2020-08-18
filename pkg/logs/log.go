package logs

import (
	"encoding/json"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
)

var (
	Log *zap.Logger
	logerr error
)

func init(){
	pwd, _ := os.Getwd()
	rawJSON, _ := ioutil.ReadFile(pwd + "/pkg/logs/config.json")
	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	Log, logerr = cfg.Build()
	if logerr != nil {
		panic(logerr)
	}

}