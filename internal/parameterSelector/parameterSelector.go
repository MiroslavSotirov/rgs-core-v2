package parameterSelector

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"gopkg.in/yaml.v3"
)

// Module for selecting stakeValue and defaultBet parameters given currency, operator-specific settings, and player history and classification

type betConfig struct {
	StakeValues    []int                     `yaml:"stakeValues"`
	DefaultBet     int                       `yaml:"defaultBet"`
	CcyMultipliers map[string]float32        `yaml:"ccyMultipliers"`
	Profiles       map[string]map[string]int `yaml:"profiles"`
	HostProfiles   map[string]string         `yaml:"hostProfiles"`
	Override       map[string]stakeConfigs   `yaml:"override`
}

type stakeConfigs map[string]stakeConfig

type stakeConfig struct {
	StakeValues []float32 `yaml:"stakeValues"`
	DefaultBet  float32   `yaml:"defaultBet"`
}

func parseBetConfig() (betConfig, rgse.RGSErr) {
	var conf betConfig

	currentDir, err := os.Getwd()
	if err != nil {
		logger.Errorf("Failed opening current directory")
		return betConfig{}, rgse.Create(rgse.BadConfigError)
	}
	configFile := filepath.Join(currentDir, "internal/parameterSelector/parameterConfig.yml")
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		logger.Fatalf("Error reading bet config file: %v", err)
		return betConfig{}, rgse.Create(rgse.YamlError)
	}

	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		logger.Fatalf("Error unmarshaling parameter yaml")
		return betConfig{}, rgse.Create(rgse.YamlError)

	}
	return conf, nil
}

func GetDemoWalletDefaults(currency string, gameID string, betSettingsCode string, playerID string) (walletInitBal engine.Money, ctFS int, waFS engine.Fixed, err rgse.RGSErr) {
	logger.Debugf("getting demo wallet defaults for player %v, ccy %v", playerID, currency)

	// default wallet amt is 100x the max bet amount for the game
	stakeValues, _, paramErr := GetGameplayParameters(engine.Money{0, currency}, betSettingsCode, gameID)
	if paramErr != nil {
		err = paramErr
		return
	}

	EC, confErr := engine.GetEngineDefFromGame(gameID)
	if confErr != nil {
		err = confErr
		return
	}
	walletInitBal = engine.Money{stakeValues[len(stakeValues)-1].Mul(engine.NewFixedFromInt(EC.EngineDefs[0].StakeDivisor)).Mul(engine.NewFixedFromInt(50)), currency}
	// solution for testing low balance
	if playerID == "lowbalance" {
		walletInitBal = engine.Money{0, currency}
	} else if playerID == "" {
		playerID = rng.RandStringRunes(8)
	} else if strings.Contains(playerID, "campaign") {
		ctFS = 10
		waFS = stakeValues[0].Mul(engine.NewFixedFromInt(EC.EngineDefs[0].StakeDivisor))
		if len(playerID) > 8 {
			i, strerr := strconv.Atoi(playerID[8:])
			if strerr == nil && i < len(stakeValues) {
				waFS = stakeValues[i].Mul(engine.NewFixedFromInt(EC.EngineDefs[0].StakeDivisor))
			}
		}
	}
	logger.Debugf("set balance: %v ; freespins: %v; fs value: %v", walletInitBal, ctFS, waFS)
	return
}

func GetGameplayParameters(lastBet engine.Money, betSettingsCode string, gameID string) ([]engine.Fixed, engine.Fixed, rgse.RGSErr) {
	// returns stakeValues and defaultBet based on host and player configuration
	logger.Debugf("getting %v stake params for config %v (lastbet %#v)", gameID, betSettingsCode, lastBet)
	betConf, err := parseBetConfig()
	//logger.Debugf("Bet Configuration: %#v", betConf)
	if err != nil {
		return []engine.Fixed{}, engine.Fixed(0), err
	}
	// get stakevalues based on host config
	baseStakeValues := betConf.StakeValues

	ccyMult, ok := betConf.CcyMultipliers[lastBet.Currency]
	if !ok {
		return []engine.Fixed{}, engine.Fixed(0), rgse.Create(rgse.BetConfigError)
	}

	profile, ok := betConf.HostProfiles[betSettingsCode]
	if !ok {
		profile, ok = betConf.HostProfiles[lastBet.Currency]
		if !ok {
			profile = "base"
		}
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
		return []engine.Fixed{}, engine.Fixed(0), rgse.Create(rgse.EngineNotFoundError)
	}
	switch engineID {
	case "mvgEngineX":
		// select minimum parameter for this game
		baseVal := fixedStakeValues[0]
		logger.Infof("stakevals: %v", fixedStakeValues)
		if len(fixedStakeValues) > 6 {
			// if possible, take a higher value
			baseVal = fixedStakeValues[6]
		}

		fixedStakeValues = []engine.Fixed{baseVal, baseVal.Mul(engine.NewFixedFromInt(2)), baseVal.Mul(engine.NewFixedFromInt(3))}
		// default stake is max val
		defaultStake = fixedStakeValues[2]
	case "mvgEngineIX":
		// default stake is the 5th index by default
		if len(fixedStakeValues) > 4 {
			defaultStake = fixedStakeValues[4]
		}
	}

	// if lastBet is not in stakeValues then use defaultBet
	if lastBet.Amount >= fixedStakeValues[0] && lastBet.Amount <= fixedStakeValues[len(fixedStakeValues)-1] {
		defaultStake = lastBet.Amount
	}

	if defaultStake < fixedStakeValues[0] {
		logger.Warnf("defaultStake too low, setting to min stakeValue")
		defaultStake = fixedStakeValues[0]
	} else if defaultStake > fixedStakeValues[len(fixedStakeValues)-1] {
		logger.Warnf("defaultStake too high, setting to max stakeValue")
		defaultStake = fixedStakeValues[len(fixedStakeValues)-1]
	}

	override, ok := betConf.Override[gameID]
	if ok {
		stakeconf, ok := override[lastBet.Currency]
		if ok {
			fixedStakeValues = make([]engine.Fixed, len(stakeconf.StakeValues))
			for i, s := range stakeconf.StakeValues {
				fixedStakeValues[i] = engine.NewFixedFromFloat(s)
			}
			defaultStake = engine.NewFixedFromFloat(stakeconf.DefaultBet)
			logger.Infof("overriding stake values: stakes= %v, defaultbet= %v", fixedStakeValues, defaultStake)
		}
	}

	logger.Debugf("stake values: %v; default stake: %v", fixedStakeValues, defaultStake)
	return fixedStakeValues, defaultStake, nil

}
