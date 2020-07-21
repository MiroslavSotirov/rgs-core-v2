package forceTool

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type ForceGP struct {
	ID       string `yaml:"id"`
	Action   string `yaml:"action"`
	StopList []int  `yaml:"stopList"`
}

type ForceGPList struct {
	Engine string
	Forces []ForceGP
}

func ReadForcedGameplays(gameName string) []ForceGPList {
	var fgList []ForceGPList
	currentDir, err := os.Getwd()
	logger.Debugf("Current Dir: %s", currentDir)
	if err != nil {
		logger.Warnf("Failed opening current directory")
		return []ForceGPList{}
	}
	forceDefDir := strings.Join([]string{currentDir, "internal/forceTool/forcedGameplays"}, "/")
	//logger.Debugf("forceDefDir: %s", forceDefDir)

	files, err := ioutil.ReadDir(forceDefDir)
	//logger.Debugf("Files in Dir: %+v", files)

	if err != nil {
		logger.Warnf("Failed listing files from current directory")
		return []ForceGPList{}
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		var c []ForceGP

		forceDef := filepath.Join(forceDefDir, f.Name())
		yamlFile, err := ioutil.ReadFile(forceDef)
		if err != nil {
			logger.Warnf("No force config found for engine #%v  #%v ", f.Name(), err)
		} else {
			err = yaml.Unmarshal(yamlFile, &c)
			if err != nil {
				logger.Warnf("Unmarshal: %v", err)

			} else {
				fg := ForceGPList{Engine: f.Name(), Forces: c}
				fgList = append(fgList, fg)
			}
		}
	}

	if gameName != "" {
		gameEngine, err := config.GetEngineFromGame(gameName)
		logger.Debugf("Found engine %s for game %s", gameEngine, gameName)
		if err != nil {
			logger.Debugf("Unable to extract game engine from game name %s", err.Error())
		} else {
			for _, fg := range fgList {
				if fg.Engine == strings.TrimSuffix(fg.Engine, ".yml") {
					return []ForceGPList{fg}
				}
			}
		}
	}
	return fgList
}

type GamesEngines struct {
	GameName string
	Engine   string
}

func ReadGamesEngines() []GamesEngines {
	var ge []GamesEngines

	for i := 0; i < len(config.GlobalGameConfig); i++ {
		for j := 0; j < len(config.GlobalGameConfig[i].Games); j++ {
			ge = append(ge,
				GamesEngines{
					GameName: config.GlobalGameConfig[i].Games[j],
					Engine:   config.GlobalGameConfig[i].EngineID,
				})
		}
	}
	return ge
}

type ForceToolParams struct {
	PlayerID string `json:"playerID"`
	GameSlug string `json:"gameSlug"`
	ForceID  string `json:"forceID"`
}
