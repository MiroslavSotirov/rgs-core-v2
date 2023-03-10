package engine

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"reflect"
	sync "sync"
	"sync/atomic"
	"time"

	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"gopkg.in/yaml.v3"
)

// game slug is attached to engine config in memcached
// player accesses play url for a specific game : /play/the-year-of-zhu
// engine is retrieved from memcached linking game slug to engine : the-year-of-zhu ==> mvgEngineI
// EngineConfig is built from configs/mvgEngineI.yml
// engingConfig consists of:
// - action-definition pairs (i.e. an engineDef and a function with which it should be run)
//    - one of these pairs is marked as the base play
//

// EngineConfig ...
type EngineConfig struct {
	RTP        float32     `yaml:"rtp"`
	Volatility float64     `yaml:"volatility"`
	Version    string      `yaml:"version"`
	EngineDefs []EngineDef `yaml:"EngineDefs,flow"`
}

type cachedEngineConfig struct {
	engineConfig atomic.Value
	loaded       bool
	cacheTime    time.Time
	semaphore    int32
}

var configCache map[string]*cachedEngineConfig = nil

const cacheRefresh time.Duration = time.Duration(1000000000000) // 10 seconds
var cacheLock sync.Mutex

// BuildEngineDefs wrapper to reuse cached versions of the configs
func BuildEngineDefs(engineID string) EngineConfig {
	//	return ReadEngineDefs(engineID)

	var config *cachedEngineConfig = nil
	var ok bool = false

	if configCache != nil {
		config, ok = configCache[engineID]
	}

	now := time.Now()
	if !ok {
		cacheLock.Lock()
		if configCache == nil {
			configCache = make(map[string]*cachedEngineConfig)
		}
		config, ok = configCache[engineID]
		if !ok {
			config = &cachedEngineConfig{
				engineConfig: atomic.Value{},
				loaded:       false,
				cacheTime:    now,
				semaphore:    0,
			}
			cfg := EngineConfig{}
			config.engineConfig.Store(&cfg)
			configCache[engineID] = config
		}
		cacheLock.Unlock()
	}

	if !config.loaded || now.Sub(config.cacheTime) > cacheRefresh {
		if atomic.CompareAndSwapInt32(&config.semaphore, 0, 1) {
			cfg := ReadEngineDefs(engineID)
			config.engineConfig.Store(&cfg)
			config.cacheTime = now
			config.loaded = true
			atomic.StoreInt32(&config.semaphore, 0)
			logger.Infof("read and cached config for engine %s", engineID)
		} else {
			for atomic.LoadInt32(&config.semaphore) != 0 {
				time.Sleep(1000000) // 1ms
			}
		}
	}

	return *config.engineConfig.Load().(*EngineConfig)
}

// BuildEngineDefs reads engine definition from yml
func ReadEngineDefs(engineID string) EngineConfig {
	// takes an engineId string and parses the corresponding yaml file into an EngineConfig
	logger.Debugf("reading engine config %s", engineID)
	currentDir, err := os.Getwd()
	if err != nil {
		panic("Failed opening current directory")
	}
	engineDef := filepath.Join(currentDir, "internal/engine/engineConfigs", engineID+".yml")
	yamlFile, err := ioutil.ReadFile(engineDef)
	if err != nil {
		logger.Errorf("No config found for engine %v  %v ", engineID, err)
	}
	c := EngineConfig{}
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		logger.Errorf("Unmarshal: %v", err)
	}
	// take values from default wherever available
	filledEngineDefs := []EngineDef{c.EngineDefs[0]}
	completeDef := c.EngineDefs[0]
	for i := 1; i < len(c.EngineDefs); i++ {
		if c.EngineDefs[i].Inheritance != "" {
			switch c.EngineDefs[i].Inheritance {
			case inheritance_none:
				completeDef = EngineDef{}
			case inheritance_first:
				completeDef = c.EngineDefs[0]
			case inheritance_prev:
			default:
				panic(fmt.Sprintf("Unrecognized inheritance mode: [%s]", c.EngineDefs[i].Inheritance))
			}
		} else {
			completeDef = c.EngineDefs[0]
		}
		completeDef.ID = c.EngineDefs[i].ID
		completeDef.Index = i

		// NB: function is used to send information about win type to client
		if c.EngineDefs[i].Function != "" {
			completeDef.Function = c.EngineDefs[i].Function
		}
		if c.EngineDefs[i].StakeDivisor != 0 {
			completeDef.StakeDivisor = c.EngineDefs[i].StakeDivisor
		}
		if c.EngineDefs[i].Probability != 0 {
			completeDef.Probability = c.EngineDefs[i].Probability
		}
		if c.EngineDefs[i].RTP != 0 {
			completeDef.RTP = c.EngineDefs[i].RTP
		}
		if len(c.EngineDefs[i].Reels) != 0 {
			completeDef.Reels = c.EngineDefs[i].Reels
		}
		if len(c.EngineDefs[i].ViewSize) != 0 {
			completeDef.ViewSize = c.EngineDefs[i].ViewSize
		}
		if len(c.EngineDefs[i].Multiplier.Multipliers) != 0 {
			completeDef.Multiplier = c.EngineDefs[i].Multiplier
		}
		if len(c.EngineDefs[i].Payouts) != 0 {
			completeDef.Payouts = c.EngineDefs[i].Payouts
		}
		if c.EngineDefs[i].WinType != "" {
			completeDef.WinType = c.EngineDefs[i].WinType
		}
		if len(c.EngineDefs[i].SpecialPayouts) != 0 {
			completeDef.SpecialPayouts = c.EngineDefs[i].SpecialPayouts
		} // must set a sham non-zero payout if override is desired in a  non-base engine
		if len(c.EngineDefs[i].WinLines) != 0 {
			completeDef.WinLines = c.EngineDefs[i].WinLines
		} // must set a sham non-zero payout if override is desired in a  non-base engine
		if len(c.EngineDefs[i].Wilds) != 0 {
			completeDef.Wilds = c.EngineDefs[i].Wilds
		} // must set a sham non-zero payout if override is desired in a  non-base engine
		if len(c.EngineDefs[i].Multiplier.Multipliers) != 0 {
			completeDef.Multiplier = c.EngineDefs[i].Multiplier
		} // must set a sham non-zero payout if override is desired in a  non-base engine
		if len(c.EngineDefs[i].Features) != 0 {
			completeDef.Features = c.EngineDefs[i].Features
		}
		if len(c.EngineDefs[i].RoulettePayouts) != 0 {
			completeDef.RoulettePayouts = c.EngineDefs[i].RoulettePayouts
		}
		if c.EngineDefs[i].WinConfig.Flags != "" {
			completeDef.WinConfig = c.EngineDefs[i].WinConfig
		}
		if c.EngineDefs[i].ReelsetId != "" {
			completeDef.ReelsetId = c.EngineDefs[i].ReelsetId
		}
		if c.EngineDefs[i].RespinAction != "" {
			completeDef.RespinAction = c.EngineDefs[i].RespinAction
		}
		if len(c.EngineDefs[i].NextMultiplierActions) != 0 {
			completeDef.NextMultiplierActions = c.EngineDefs[i].NextMultiplierActions
		}
		if len(c.EngineDefs[i].HoldMultiplierActions) != 0 {
			completeDef.HoldMultiplierActions = c.EngineDefs[i].HoldMultiplierActions
		}
		if len(c.EngineDefs[i].FeatureStages) != 0 {
			completeDef.FeatureStages = c.EngineDefs[i].FeatureStages
		}

		// respin must be explicitly set to true if it is intended to be true, no inheritance from base
		completeDef.RespinAllowed = c.EngineDefs[i].RespinAllowed
		// same for variable winlines because it is a boolean and no way to tell if false or omitted
		completeDef.VariableWL = c.EngineDefs[i].VariableWL
		completeDef.Compounding = c.EngineDefs[i].Compounding
		completeDef.ExpectedPayout = c.EngineDefs[i].ExpectedPayout
		filledEngineDefs = append(filledEngineDefs, completeDef)
	}
	c.EngineDefs = filledEngineDefs
	return c

}

func (config EngineConfig) DefIdByName(action string) int {
	// gets first enginedef matching action name, returns -1 if no match
	for i, engine := range config.EngineDefs {
		if engine.ID == action {
			return i
		}
	}
	return -1
}

func (config EngineConfig) getEngineAndMethodInternal(action string, exception bool) (reflect.Value, int, rgse.RGSErr) {
	//log.Printf("Retrieving method: %v", action)
	var matchedEngines []EngineDef
	sumEngineProbabilities := 0
	var selectedEngine EngineDef
	// find all engineDefs whose ID match the action
	for _, engine := range config.EngineDefs {
		if engine.ID == action {

			matchedEngines = append(matchedEngines, engine)
			sumEngineProbabilities += engine.Probability
		}
	}
	//logger.Debugf("matched %v engines", len(matchedEngines))
	if len(matchedEngines) == 0 {
		var e rgse.RGSErr
		if exception {
			e = rgse.Create(rgse.EngineNotFoundError)
		} else {
			e = rgse.CreateWithoutException(rgse.EngineNotFoundError)
		}
		e.AppendErrorText(fmt.Sprintf("No engine matched action %v", action))
		return reflect.Value{}, 0, e
	}
	if len(matchedEngines) == 1 {
		selectedEngine = matchedEngines[0]
	} else {
		engineDefThreshold := rng.RandFromRange(sumEngineProbabilities)
		logger.Debugf("Using Probabilities to Select Engine -- Probability Sum: %v; Threshold: %v", sumEngineProbabilities, engineDefThreshold)
		engineDefCurrent := -1
		for idx, engine := range matchedEngines {
			engineDefCurrent += engine.Probability
			if engineDefCurrent >= engineDefThreshold {
				selectedEngine = engine
				logger.Debugf("selecting engine %d", idx)
				break
			}
		}
	}
	logger.Debugf("method selected: %s engine index: %d", selectedEngine.Function, selectedEngine.Index)
	return reflect.ValueOf(selectedEngine).MethodByName(selectedEngine.Function), selectedEngine.Index, nil
}

func (config EngineConfig) getEngineAndMethod(action string) (reflect.Value, int, rgse.RGSErr) {
	return config.getEngineAndMethodInternal(action, true)
}

func (engine EngineConfig) NumSpinsStat() int {
	// returns the number of spins required to achieve RTP within less than 1% with 95% confidence
	stdev := math.Sqrt(engine.Volatility)
	numSpins := float64(1000)
	tolerance := float64(100)
	for ; tolerance > 0.05; numSpins *= 10 {
		tolerance = stdev * 2 / math.Sqrt(numSpins)
	}
	return int(numSpins)
}
