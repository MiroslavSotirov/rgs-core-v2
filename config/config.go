package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/crgimenes/goconfig"
	_ "github.com/crgimenes/goconfig/yaml"
	"github.com/golang/glog"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"gopkg.in/yaml.v3"
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
	Name string `yaml:"name" cfg:"name" cfgDefault:""`
}

func (s Server) IsV3() bool {
	return strings.Contains(s.Name, "elysium")
}

type StoreConfig struct {
	StoreRemoteUrl  string `yaml:"storeurl" cfg:"storeurl" cfgDefault:"https://gnrc-api.dashur.io/v1/gnrc/maverick"`
	StoreAppId      string `yaml:"storeappid" cfg:"storeappid" cfgDefault:"maverick_user"`
	StoreAppPass    string `yaml:"storeapppass" cfg:"storeapppass" cfgDefault:"Passw0rd!"`
	StoreMaxRetries int    `yaml:"storemaxretries" cfg:"storemaxretries" cfgDefault:1`
	StoreTimeoutMs  int64  `yaml:"storetimeoutms" cfg:"storetimeoutms" cfgDefault:3000`
}

// Config structure
type Config struct {
	DevMode         bool   `yaml:"devmode" cfg:"devmode" cfgDefault:"false"`
	MCRouter        string `yaml:"mcrouter" cfg:"mcrouter" cfgDefault:"10.42.0.86:5000"`
	Server          `yaml:"server"`
	Local           bool        `yaml:"local" cfg:"local" cfgDefault:"false"`
	Logging         string      `yaml:"logging" cfg:"logging" cfgDefault:"debug"`
	DashurConfig    StoreConfig `yaml:"dashurconf"`
	DefaultPlatform string      `yaml:"defaultplatform" cfg:"defaultplatform" cfgDefault:"html5"`
	DefaultLanguage string      `yaml:"defaultlanguage" cfg:"defaultlanguage" cfgDefault:"en"`
	DemoTokenPrefix string      `yaml:"demotokenprefix" cfg:"demotokenprefix" cfgDefault:"demo-token"`
	DemoCurrency    string      `yaml:"democurrency" cfg:"democurrency" cfgDefault:"USD"`
	LogAccount      string      `yaml:"logaccount" cfg:"logaccount" cfgDefault:"145472021_144443389"`
	SentryDsn       string      `yaml:"sentryDsn" cfg:"sentryDsn" cfgDefault:""`
	Environment     string      `yaml:"environment" cfg:"environment" cfgDefault:"local"`
	DataLimit       int         `yaml:"datalimit" cfg:"datalimit" cfgDefault:"800"`
	LocalDataTtl    int64       `yaml:"localdatattl" cfg:"localdatattl" cfgDefault:"0"`
	ExtPlaycheck    string      `yaml:"extplaycheck" cfg:"extplaycheck" cfgDefault:"https://dev.elysiumstudios.se/game-history"`
}

// Game config structure
type GameConfig struct {
	EngineID string    `yaml:"engineID"`
	Category string    `yaml:"category"`
	Games    []GameDef `yaml:"games"`
}
type GameDef struct {
	Name  string `yaml:"name"`
	Item  string `yaml:"item"`
	Title string `yaml:"title"`
	Flags string `taml:"flags"`
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
	err = InitGamification()
	if err != nil {
		BadConfigError(err)
	}
	err = InitHashes()
	if err != nil {
		BadConfigError(err)
	}
	////prints configuration
	logger.Infof("Game Config: %v", GlobalGameConfig)
}

type GamificationType struct {
	Levels   int32  `yaml:"levels"`
	Stages   int32  `yaml:"stages"`
	Function string `yaml:"function"`
	SpinsMin int    `yaml:"spinsMin"`
	SpinsMax int    `yaml:"spinsMax"`
}

var GameGamification map[string]GamificationType

func InitGamification() error {

	currentDir, err := os.Getwd()
	if err != nil {
		logger.Errorf("Failed opening current directory")
		return err
	}
	configFile := filepath.Join(currentDir, "config/gamification.yml")
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		logger.Fatalf("Error reading gamification config file: %v", err)
		return err
	}

	err = yaml.Unmarshal(yamlFile, &GameGamification)
	if err != nil {
		logger.Fatalf("Error unmarshaling parameter yaml %v", err)
		return err

	}
	return nil
}

func InitGameConfig() error {
	yamlFile, err := ioutil.ReadFile(gameConfigFile)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(yamlFile, &GlobalGameConfig)
}

func GetEngineFromGame(gameName string) (engineID string, err rgse.RGSErr) {
	for i := 0; i < len(GlobalGameConfig); i++ {
		for j := 0; j < len(GlobalGameConfig[i].Games); j++ {
			if GlobalGameConfig[i].Games[j].Name == gameName {
				engineID = GlobalGameConfig[i].EngineID
				return
			}
		}
	}
	err = rgse.Create(rgse.EngineNotFoundError)
	err.AppendErrorText(fmt.Sprintf(" for %s", gameName))
	return
}
