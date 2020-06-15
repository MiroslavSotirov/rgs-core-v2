package forceTool

import (
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func SetForce(gameID string, forceID string, playerID string) error {
	// sets force in memcached
	return store.MC.Set(&memcache.Item{Key: playerID + "::" + gameID, Value: []byte(forceID), Expiration: 3600})
}

func ClearForce(gameID string, playerID string) error {
	logger.Warnf("Deleting force %v on player %v", gameID, playerID)
	return store.MC.Delete(playerID + "::" + gameID)
}

func GetForceValues(params engine.GameParams, previousGamestate engine.Gamestate, playerID string) (forcedGamestate engine.Gamestate, err rgse.RGSErr) {
	forceID, mcerr := store.MC.Get(playerID + "::" + previousGamestate.Game)
	if mcerr != nil {
		err = rgse.Create(rgse.NoForceError)
		return
	}
	// automatically clear the force once it has been used
	_ = ClearForce(previousGamestate.Game, playerID)

	if params.Stake == 0 && previousGamestate.Action != "base" {
		params.Stake = previousGamestate.BetPerLine.Amount
	}
	forcedGamestate, err = smartForceFromID(params, previousGamestate, string(forceID.Value))
	if err != nil {
		// special handling for forces not allowed right now
		if err.(*rgse.RGSError).ErrCode == rgse.ForceProhibited {
			// return force to mc
			mcerr = SetForce(previousGamestate.Game, string(forceID.Value), playerID)
			if mcerr != nil {
				err = rgse.Create(rgse.NoForceError)
			}
		}
	}
	logger.Warnf("Created forced gamestate: %v", forcedGamestate)
	return
}

type ForceGameplay struct {
	// stores values for forcing a gameplay
	ID          string         `yaml:"id"`
	Action      string         `yaml:"action"`
	ReelsetId   int            `yaml:"reelsetId"`
	StopList    []int          `yaml:"stopList"`
}

func BuildForce(engineID string) []ForceGameplay {
	// takes an engineId string and parses the corresponding yaml file into a slice of ForceGameplay structs
	currentDir, err := os.Getwd()
	if err != nil {
		logger.Warnf("Failed opening current directory")
		return []ForceGameplay{}
	}
	forceDef := filepath.Join(currentDir, "internal/forceTool/forcedGameplays", engineID+".yml")
	yamlFile, err := ioutil.ReadFile(forceDef)
	if err != nil {
		logger.Warnf("No force config found for engine #%v  #%v ", engineID, err)
		return []ForceGameplay{}

	}
	var c []ForceGameplay
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		logger.Warnf("Unmarshal: %v", err)
		return []ForceGameplay{}

	}

	for i := 0; i < len(c); i++ {
		if c[i].Action == "" {
			c[i].Action = "base"
		}
		if len(c[i].StopList) == 0 {
			//todo: smart generate stop list. for now don't allow this to be empty
			logger.Fatalf("Must set stopList")
		}
	}
	return c
}

func smartForceFromID(params engine.GameParams, previousGamestate engine.Gamestate, forceID string) (engine.Gamestate, rgse.RGSErr) {
	// build force gamestates

	engineID, err := config.GetEngineFromGame(previousGamestate.Game)
	if err != nil {
		return engine.Gamestate{}, err
	}
	engineConf := engine.BuildEngineDefs(engineID)
	forces := BuildForce(engineID)
	actions := previousGamestate.NextActions
	if len(actions) == 1 && actions[0] == "finish" {
		actions = []string{"base", "finish"}
	}
	// for engine VII make retrigger multiplier increments automatically
	if engineID == "mvgEngineVII" && (strings.HasPrefix(forceID, "retrigger") || strings.HasSuffix(forceID, "scatter")) && strings.HasPrefix(actions[0], "freespin"){
		isScatter := strings.HasSuffix(forceID, "scatter")
		if previousGamestate.DefID >= 5 &&  previousGamestate.DefID < 8 { // reelset 5 and above are freespins
			if isScatter {
				fsAction := previousGamestate.DefID - 3
				forceID = fmt.Sprintf("FS%d-%s", fsAction, forceID)
			} else {
				retriggerIndex := previousGamestate.DefID - 4
				forceID = fmt.Sprintf("retrigger%d", retriggerIndex)
			}
		}
		if previousGamestate.DefID >= 8 {
			if isScatter {
				fsAction := previousGamestate.DefID - 6
				forceID = fmt.Sprintf("FS%d-%s", fsAction, forceID)
			} else {
				// reset
				retriggerIndex := 4
				forceID = fmt.Sprintf("retrigger%d", retriggerIndex)
			}

		}
	}

	logger.Debugf("Engine: %s ForceID: %s ReelsetID: %d", engineID, forceID, previousGamestate.DefID)
	var gamestate engine.Gamestate
	for _, force := range forces {
		if force.ID == forceID {
			// check if force is invalid
			if force.Action != params.Action {
				return gamestate, rgse.Create(rgse.ForceProhibited)
			}
			engineDef := engineConf.EngineDefs[force.ReelsetId]
			err = engineDef.SetForce(force.StopList)
			if err != nil {
				return gamestate, err
			}
			// get engine and action
			method := reflect.ValueOf(engineDef).MethodByName(engineDef.Function)
			gamestateAndNextActions := method.Call([]reflect.Value{reflect.ValueOf(params)})

			gamestate, ok := gamestateAndNextActions[0].Interface().(engine.Gamestate)
			if !ok {
				panic("value not a gamestate")
			}
			chargeWager := true
			betPerLine :=  params.Stake
			if previousGamestate.NextActions[0] != "finish" {
				chargeWager = false
				betPerLine = previousGamestate.BetPerLine.Amount
			}
			var totalBet engine.Money
			currency := previousGamestate.BetPerLine.Currency
			if force.Action == "respin" {
				betPerLine = previousGamestate.BetPerLine.Amount
				totalBet = engine.Money{previousGamestate.RespinPriceReel(params.RespinReel), currency}
			}
			gamestate.PostProcess(previousGamestate, chargeWager, totalBet, engineConf, betPerLine, actions, currency)

			return gamestate, nil
		}
	}

	return gamestate, rgse.Create(rgse.NoForceError)
}
