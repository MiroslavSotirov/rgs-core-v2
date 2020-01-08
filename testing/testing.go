package testing

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"log"
	"os"
	"path"
	"runtime"
)

func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
	//config.InitConfig()
	logConfig := logger.Configuration{false, logger.Debug}

	err = logger.NewLogger(logConfig)
	if err != nil {
		log.Printf("Logging configuration error: %s", err)
	}
	err = config.InitGameConfig()
	if err != nil {
		log.Printf("Game configuration error: %s", err)
	}
}
