package logs

import (
	"encoding/json"
	"go.uber.org/zap"
)

var log *zap.Logger

func init() {
	var cfg zap.Config
	rawJSON := []byte(`{
		"level": "info",
		"outputPaths": ["stdout"],
		"errorOutputPaths": ["stderr"],
		"encoding": "console",
		"encoderConfig": {
			"messageKey": "message",
			"levelKey": "level",
			"levelEncoder": "lowercase"
		}
	}`)
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	log, _ = cfg.Build()
	defer log.Sync()
}

func GetLogger() *zap.Logger {
	return log
}
