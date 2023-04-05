package parameterSelector

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"gopkg.in/yaml.v3"
)

// Module for selecting stakeValue and defaultBet parameters given currency, operator-specific settings, and player history and classification

type betConfig struct {
	StakeValues []int `yaml:"stakeValues"`
	DefaultBet  int   `yaml:"defaultBet"`
	//	CcyMultipliers map[string]float32        `yaml:"ccyMultipliers"`
	CcyMultipliers map[string]map[string]float32      `yaml:"ccyMultipliers"`
	CcyMinorUnits  map[string]int                     `yaml:"ccyMinorUnits"`
	Profiles       map[string]map[string]int          `yaml:"profiles"`
	HostProfiles   map[string]string                  `yaml:"hostProfiles"`
	Override       map[string]map[string]stakeConfigs `yaml:"override`
}

type stakeConfigs map[string]stakeConfig

type stakeConfig struct {
	StakeValues []float32 `yaml:"stakeValues"`
	DefaultBet  float32   `yaml:"defaultBet"`
	MinBet      float32   `yaml:"minBet"`
	MaxBet      float32   `yaml:"maxBet"`
}

var cachedConfig atomic.Value = atomic.Value{}
var cachedTime time.Time = time.Now()
var semaphore int32 = 0

const cacheRefresh time.Duration = time.Duration(10000000000) // 10 seconds

func parseBetConfig() (betConfig, rgse.RGSErr) {
	now := time.Now()

	if cachedConfig.Load() == nil {
		if atomic.CompareAndSwapInt32(&semaphore, 0, 1) {
			cfg, err := readBetConfig()
			if err != nil {
				return betConfig{}, err
			}
			logger.Infof("Loaded and cached betConfig")
			cachedTime = now
			cachedConfig.Store(&cfg)
			atomic.StoreInt32(&semaphore, 0)
		} else {
			for cachedConfig.Load() == nil {
				time.Sleep(1000000) // 1ms
			}
		}
	}

	if now.Sub(cachedTime) > cacheRefresh {
		if atomic.CompareAndSwapInt32(&semaphore, 0, 1) {
			cfg, err := readBetConfig()
			if err != nil {
				return betConfig{}, err
			}
			logger.Infof("Reloaded and cached betConfig")
			cachedTime = now
			cachedConfig.Store(&cfg)
			atomic.StoreInt32(&semaphore, 0)
		}
	}
	return *cachedConfig.Load().(*betConfig), nil
}

func readBetConfig() (betConfig, rgse.RGSErr) {
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
		logger.Fatalf("Error unmarshaling %v", err)
		return betConfig{}, rgse.Create(rgse.YamlError)
	}
	err = validateBetConfig(conf)
	if err != nil {
		logger.Fatalf("Error validating bet config")
		return betConfig{}, rgse.Create(rgse.BadConfigError)
	}
	return conf, nil
}

func validateBetConfig(betConf betConfig) rgse.RGSErr {
	paramService := createLocalParameterService(betConf)
	valid := true
	for _, gc := range config.GlobalGameConfig {
		for ccy, _ := range betConf.CcyMultipliers["default"] {
			sv, _, _, _, err := getGameplayParameters(engine.Money{0, ccy}, "", gc.Games[0].Name, "", betConf, paramService)
			if err != nil {
				logger.Infof("Error validating %s: %v", gc.EngineID, err)
				return err
			}
			if len(sv) == 0 {
				valid = false
				//logger.Infof("Error validating, %s has no stakes", gc.EngineID)
				//				return rgse.Create(rgse.BadConfigError)
			}
		}
	}
	if !valid {
		return rgse.Create(rgse.BadConfigError)
	}

	logger.Infof("betConfig validation is OK")
	return nil
}

func GetDemoWalletDefaults(currency string, gameID string, betSettingsCode string, playerID string, betSettingId string) (walletInitBal engine.Money, ctFS int, waFS engine.Fixed, err rgse.RGSErr) {
	logger.Debugf("getting demo wallet defaults for player: %v, ccy: %v, betSettingId: %v", playerID, currency, betSettingId)

	// default wallet amt is 100x the max bet amount for the game (except in local mode to enable long automated playtesting)
	walletamtmult := 100
	if config.GlobalConfig.DevMode {
		walletamtmult = 100000
	}
	stakeValues, _, _, _, paramErr := GetGameplayParameters(engine.Money{0, currency}, betSettingsCode, gameID, betSettingId)
	if paramErr != nil {
		err = paramErr
		return
	}

	EC, confErr := engine.GetEngineDefFromGame(gameID)
	if len(EC.EngineDefs) == 0 {
		logger.Debugf("  EC.EngineDefs has zero length")
	}
	if confErr != nil {
		err = confErr
		return
	}
	walletInitBal = engine.Money{stakeValues[len(stakeValues)-1].MulFloat(engine.NewFixedFromInt(EC.EngineDefs[0].StakeDivisor)).MulFloat(engine.NewFixedFromInt(walletamtmult)), currency}
	logger.Debugf("wallet initial balance= %v", walletInitBal)
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

func GetGameplayParameters(lastBet engine.Money, betSettingsCode string, gameID string, betSettingId string) (
	stakeValues []engine.Fixed, defaultBet engine.Fixed, minBet engine.Fixed, maxBet engine.Fixed, rgserr rgse.RGSErr) {
	betConf, err := parseBetConfig()
	if err != nil {
		rgserr = err
		return
	}
	return getGameplayParameters(lastBet, betSettingsCode, gameID, betSettingId, betConf, GetParameterService())
}

func getGameplayParameters(lastBet engine.Money, betSettingsCode string, gameID string, betSettingId string, betConf betConfig, paramService ParameterService) (
	stakeValues []engine.Fixed, defaultBet engine.Fixed, minBet engine.Fixed, maxBet engine.Fixed, rgserr rgse.RGSErr) { // ([]engine.Fixed, engine.Fixed, rgse.RGSErr) {
	baseStakeValues := betConf.StakeValues

	ccyMult, ok := paramService.CurrencyMultiplier(lastBet.Currency, betSettingId)
	if !ok {
		//		return []engine.Fixed{}, engine.Fixed(0), rgse.Create(rgse.BetConfigError)
		rgserr = rgse.Create(rgse.BetConfigError)
		return
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
		rgserr = rgse.Create(rgse.EngineNotFoundError)
		return
	}

	overrideId := betSettingId
	override, ok := betConf.Override[overrideId][engineID]
	if !ok {
		override, ok = betConf.Override["default"][engineID]
		if !ok {
			override, ok = betConf.Override["default"][gameID]
		}
	}
	if ok {
		var mult engine.Fixed
		stakeconf, ok := override[lastBet.Currency]
		if ok {
			mult = engine.NewFixedFromInt(1)
		} else {
			stakeconf, ok = override["credits"]
			mult = engine.NewFixedFromFloat(ccyMult)
		}
		if ok {

			defaultStake = engine.NewFixedFromFloat(stakeconf.DefaultBet).Mul(mult)
			minBet = engine.NewFixedFromFloat(stakeconf.MinBet).Mul(mult)
			maxBet = engine.NewFixedFromFloat(stakeconf.MaxBet).Mul(mult)
			fixedStakeValues = make([]engine.Fixed, len(stakeconf.StakeValues))
			for i, s := range stakeconf.StakeValues {
				fixedStakeValues[i] = engine.NewFixedFromFloat(s).Mul(mult)
			}
		}
	}

	minorMults := []int64{
		1,
		10,
		100,
	}
	minorUnit := 2
	if mu, ok := betConf.CcyMinorUnits[lastBet.Currency]; ok {
		if mu > 2 {
			mu = 2
		}
		minorUnit = mu
	}
	if minorUnit >= len(minorMults) {
		logger.Errorf("currency %s has an unsuported minor unit", lastBet.Currency)
		rgserr = rgse.Create(rgse.BetConfigError)
		return
	}
	minorMult := minorMults[minorUnit]
	validStakeValues := []engine.Fixed{}
	for _, s := range fixedStakeValues {

		minor := s.MulInt(minorMult)
		if minor != minor.Trunc() {
			// bet amount has too many decimals for this currency
			// logger.Debugf("bet amount %v has too many decimals for currency %s", s, lastBet.Currency)
			continue
		}
		validStakeValues = append(validStakeValues, s)
	}
	if len(validStakeValues) == 0 {
		logger.Errorf("Bet limiter disallowed all engine %s for currency %s", engineID, lastBet.Currency)
		rgserr = rgse.Create(rgse.BetConfigError)
		return
	}

	abs := func(v int64) int64 {
		if v < 0 {
			return -v
		}
		return v
	}

	if len(validStakeValues) != len(fixedStakeValues) {
		idx := -1
		for i, s := range validStakeValues {
			if idx < 0 ||
				abs(int64(s)-int64(defaultStake)) < abs(int64(s)-int64(validStakeValues[idx])) {
				idx = i
			}
		}
		defaultStake = validStakeValues[idx]
	}

	fixedStakeValues = validStakeValues

	switch engineID {
	case "mvgEngineX":
		// select minimum parameter for this game
		baseVal := fixedStakeValues[0]
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
		//		logger.Warnf("defaultStake too low, setting to min stakeValue")
		defaultStake = fixedStakeValues[0]
	} else if defaultStake > fixedStakeValues[len(fixedStakeValues)-1] {
		//		logger.Warnf("defaultStake too high, setting to max stakeValue")
		defaultStake = fixedStakeValues[len(fixedStakeValues)-1]
	}

	stakeValues = fixedStakeValues
	defaultBet = defaultStake

	return
}

func GetCurrencyMinorUnit(ccy string) (minorUnit int, err rgse.RGSErr) {
	var betConf betConfig
	betConf, err = parseBetConfig()
	if err != nil {
		return
	}
	var ok bool
	minorUnit, ok = betConf.CcyMinorUnits[ccy]
	if !ok {
		minorUnit = 2
		return
	}
	return
}
