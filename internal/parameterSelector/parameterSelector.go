package parameterSelector

import (
	"errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgserror "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Module for selecting stakeValue and defaultBet parameters given currency, operator-specific settings, and player history and classification

type betConfig struct {
	StakeValues    []int                     `yaml:"stakeValues"`
	DefaultBet     int                       `yaml:"defaultBet"`
	CcyMultipliers map[string]float32        `yaml:"ccyMultipliers"`
	Profiles       map[string]map[string]int `yaml:"profiles"`
	HostProfiles   map[string]string         `yaml:"hostProfiles"`
}

func parseBetConfig() (betConfig, error) {
	var conf betConfig

	currentDir, err := os.Getwd()
	if err != nil {
		logger.Errorf("Failed opening current directory")
		return betConfig{}, errors.New("Parameter Selection unmarshaling error")
	}
	configFile := filepath.Join(currentDir, "internal/parameterSelector/parameterConfig.yml")
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		logger.Fatalf("Error reading bet config file: %v", err)
		return betConfig{}, errors.New("parameter selection unmarshaling error")
	}

	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		logger.Fatalf("Error unmarshaling parameter yaml")
		return betConfig{}, errors.New("parameter selection unmarshaling error")

	}
	return conf, nil
}

func GetGameplayParameters(lastBet engine.Fixed, player store.PlayerStore, gameID string) ([]engine.Fixed, engine.Fixed, rgserror.IRGSError) {
	// returns stakeValues and defaultBet based on host and player configuration
	logger.Debugf("getting %v stake params for player: %#v", gameID, player)
	betConf, err := parseBetConfig()
	//logger.Debugf("Bet Configuration: %#v", betConf)
	if err != nil {
		betconfigerr := rgserror.ErrBetConfig
		betconfigerr.AppendErrorText(err.Error())
		return []engine.Fixed{}, engine.Fixed(0), betconfigerr
	}
	// get stakevalues based on host config
	baseStakeValues := betConf.StakeValues

	ccyMult, ok := betConf.CcyMultipliers[player.Balance.Currency]
	if !ok {
		betconfigerr := rgserror.ErrBetConfig
		betconfigerr.AppendErrorText("Unknown Bet Multiplier")
		return []engine.Fixed{}, engine.Fixed(0), rgserror.ErrBetConfig
	}

	profile, ok := betConf.HostProfiles[player.BetLimitSettingCode]
	if !ok {
		profile = "base"
	}

	// get default value
	defaultIndex, ok := betConf.Profiles[profile]["default"]
	if !ok {
		defaultIndex = betConf.DefaultBet
	}
	defaultStake := engine.NewFixedFromInt(betConf.StakeValues[defaultIndex]).Mul(engine.NewFixedFromFloat(ccyMult))

	// slice from max
	max, ok := betConf.Profiles[profile]["max"]
	if ok {
		baseStakeValues = baseStakeValues[:max]
	}
	// slice from min
	min, ok := betConf.Profiles[profile]["min"]
	if ok {
		baseStakeValues = baseStakeValues[min:]
	}

	// convert for ccy
	fixedStakeValues := make([]engine.Fixed, len(baseStakeValues))
	for i, stake := range baseStakeValues {
		fixedStakeValues[i] = engine.NewFixedFromInt(stake).Mul(engine.NewFixedFromFloat(ccyMult))
	}
	// process for game
	engineID, err := config.GetEngineFromGame(gameID)
	if err != nil {
		logger.Errorf("No such game found: %v", gameID)
		return []engine.Fixed{}, engine.Fixed(0), rgserror.ErrEngineNotFound
	}
	switch engineID {
	case "mvgEngineX":
		// select minimum parameter for this game
		baseVal := fixedStakeValues[0]
		fixedStakeValues = []engine.Fixed{baseVal, baseVal.Mul(engine.NewFixedFromInt(2)), baseVal.Mul(engine.NewFixedFromInt(3))}
		// default stake is max val
		defaultStake = fixedStakeValues[2]
	}

	// if lastBet is not in stakeValues then use defaultBet
	if lastBet >= fixedStakeValues[0] && lastBet <= fixedStakeValues[len(fixedStakeValues)-1] {
		defaultStake = lastBet
	}

	if defaultStake < fixedStakeValues[0] {
		logger.Warnf("defaultStake too low, setting to min stakeValue")
		defaultStake = fixedStakeValues[0]
	} else if defaultStake > fixedStakeValues[len(fixedStakeValues)-1] {
		logger.Warnf("defaultStake too high, setting to max stakeValue")
		defaultStake = fixedStakeValues[len(fixedStakeValues)-1]
	}
	logger.Debugf("stake values: %v; default stake: %v", fixedStakeValues, defaultStake)
	return fixedStakeValues, defaultStake, nil

}
