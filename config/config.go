package config

import (
	"fmt"
	"github.com/crgimenes/goconfig"
	_ "github.com/crgimenes/goconfig/yaml"
	"github.com/golang/glog"
	rgserror "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

const configFile = "config/config.yml"
const gameConfigFile = "config/gameConfig.yml"

// GlobalConfig variable
var GlobalConfig Config
var GlobalGameConfig []GameConfig

// Server config structure
type Server struct {
	Host string `yaml:"host" cfg:"host" cfgDefault:"0.0.0.0"`
	Port string `yaml:"port" cfg:"port" cfgDefault:"3000"`
}

type StoreConfig struct {
	StoreRemoteUrl  string `yaml:"storeurl" cfg:"storeurl" cfgDefault:"https://gnrc-api.dashur.io/v1/gnrc/maverick"`
	StoreAppId      string `yaml:"storeappid" cfg:"storeappid" cfgDefault:"maverick_user"`
	StoreAppPass    string `yaml:"storeapppass" cfg:"storeapppass" cfgDefault:"Passw0rd!"`
}

// Config structure
type Config struct {
	DevMode         bool   `yaml:"devmode" cfg:"devmode" cfgDefault:"false"`
	MCRouter        string `yaml:"mcrouter" cfg:"mcrouter" cfgDefault:"10.42.0.86:5000"`
	Server          `yaml:"server"`
	Local           bool   `yaml:"local" cfg:"local" cfgDefault:"false"`
	Logging         string `yaml:"logging" cfg:"logging" cfgDefault:"debug"`
	DashurConfig 	StoreConfig `yaml:"dashurconf"`
	DefaultPlatform string `yaml:"defaultplatform" cfg:"defaultplatform" cfgDefault:"html5"`
	DefaultLanguage string `yaml:"defaultlanguage" cfg:"defaultlanguage" cfgDefault:"en"`
	DemoTokenPrefix string `yaml:"demotokenprefix" cfg:"demotokenprefix" cfgDefault:"demo-token"`
	DemoCurrency    string `yaml:"democurrency" cfg:"democurrency" cfgDefault:"USD"`
}

// Game config structure
type GameConfig struct {
	EngineID string   `yaml:"engineID"`
	Games    []string `yaml:"games"`
}

// ConfigError - forces server to exit for configuration error
func BadConfigError(err error) {
	fmt.Printf("Bad Configuration %s", err)
	os.Exit(2)
}

// read and parse config file to config structure
func InitConfig() {
	goconfig.File = configFile
	err := goconfig.Parse(&GlobalConfig)
	if err != nil {
		BadConfigError(err)
	}

	////prints configuration
	glog.Infof("Config: %v", GlobalConfig)

	logConfig := logger.Configuration{false, GlobalConfig.Logging}

	err = logger.NewLogger(logConfig)
	if err != nil {
		glog.Errorf("Logging configuration error: %s", err)
	}
	err = InitGameConfig()
	if err != nil {
		BadConfigError(err)
	}
	////prints configuration
	logger.Infof("Game Config: %v", GlobalGameConfig)
}

func InitGameConfig() error {
	yamlFile, err := ioutil.ReadFile(gameConfigFile)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(yamlFile, &GlobalGameConfig)
}

func GetEngineFromGame(gameName string) (engineID string, err rgserror.IRGSError) {
	for i := 0; i < len(GlobalGameConfig); i++ {
		for j := 0; j < len(GlobalGameConfig[i].Games); j++ {
			if GlobalGameConfig[i].Games[j] == gameName {
				engineID = GlobalGameConfig[i].EngineID
				return
			}
		}
	}
	err = rgserror.ErrEngineConfig
	err.AppendErrorText(fmt.Sprintf(" for %s", gameName))
	return
}
