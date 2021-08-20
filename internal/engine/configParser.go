package engine

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"reflect"

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

// BuildEngineDefs reads engine definition from yml
func BuildEngineDefs(engineID string) EngineConfig {
	// takes an engineId string and parses the corresponding yaml file into an EngineConfig
	currentDir, err := os.Getwd()
	if err != nil {
		logger.Fatalf("Failed opening current directory")
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
	for i := 1; i < len(c.EngineDefs); i++ {
		completeDef := c.EngineDefs[0]
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
		if c.EngineDefs[i].WinConfig.Flags != "" {
			completeDef.WinConfig = c.EngineDefs[i].WinConfig
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

func (config EngineConfig) getEngineAndMethod(action string) (reflect.Value, rgse.RGSErr) {
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
		e := rgse.Create(rgse.EngineNotFoundError)
		e.AppendErrorText(fmt.Sprintf("No engine matched action %v", action))
		return reflect.Value{}, e
	}
	if len(matchedEngines) == 1 {
		selectedEngine = matchedEngines[0]
	} else {
		engineDefThreshold := rng.RandFromRange(sumEngineProbabilities)
		logger.Debugf("Using Probabilities to Select Engine -- Probability Sum: %v; Threshold: %v", sumEngineProbabilities, engineDefThreshold)
		engineDefCurrent := -1
		for _, engine := range matchedEngines {
			engineDefCurrent += engine.Probability
			if engineDefCurrent >= engineDefThreshold {
				selectedEngine = engine
				break
			}
		}
	}
	return reflect.ValueOf(selectedEngine).MethodByName(selectedEngine.Function), nil
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
