package forceTool

import (
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/engine"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/store"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
)

func SetForce(gameID string, forceID string, playerID string) error {
	// sets force in memcached
	return store.MC.Set(&memcache.Item{Key: playerID + "::" + gameID, Value: []byte(forceID), Expiration: 3600})
}

func ClearForce(gameID string, playerID string) error {
	logger.Warnf("Deleting force %v on player %v", gameID, playerID)
	return store.MC.Delete(playerID + "::" + gameID)
}

func GetForceValues(betPerLine engine.Fixed, previousGamestate engine.Gamestate, gameID string, playerID string) (engine.Gamestate, error) {
	forceID, err := store.MC.Get(playerID + "::" + gameID)
	if err != nil {
		logger.Warnf("No force value set")
		return engine.Gamestate{}, err
	}
	// automatically clear the force once it has been used
	_ = ClearForce(gameID, playerID)

	forcedGamestate := smartForceFromID(betPerLine, previousGamestate, gameID, string(forceID.Value))

	logger.Warnf("Created forced gamestate: %v", forcedGamestate)
	return forcedGamestate, nil
}

type ForceGameplay struct {
	// stores values for forcing a gameplay
	ID          string         `yaml:"id"`
	Action      string         `yaml:"action"`
	ReelsetId   int            `yaml:"reelsetId"`
	Prizes      []engine.Prize `yaml:"prizes"`
	StopList    []int          `yaml:"stopList"`
	NextActions []string       `yaml:"nextActions"`
	Multiplier  int            `yaml:"multiplier"`
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
		if c[i].Multiplier == 0 {
			c[i].Multiplier = 1
		}
		if len(c[i].StopList) == 0 {
			//todo: smart generate stop list. for now don't allow this to be empty
			logger.Fatalf("Must set stopList")
		}
		if len(c[i].NextActions) == 0 {
			c[i].NextActions = []string{"finish"}
		}
	}
	return c
}

func generateSymbolGrid(stopList []int, engineID string, reelsetID int) [][]int {
	// get engineDef
	engineConf := engine.BuildEngineDefs(engineID)
	engineDef := engineConf.EngineDefs[reelsetID]

	// from stop positions, generate view
	return engine.GetSymbolGridFromStopList(engineDef.Reels, engineDef.ViewSize, stopList)
}

func smartForceFromID(betPerLine engine.Fixed, previousGamestate engine.Gamestate, gameID string, forceID string) engine.Gamestate {
	// build force gamestates
	engineID, err := config.GetEngineFromGame(gameID)
	if err != nil {
		return engine.Gamestate{}
	}
	engineConf := engine.BuildEngineDefs(engineID)
	forces := BuildForce(engineID)
	actions := previousGamestate.NextActions
	if len(actions) == 1 && actions[0] == "finish" {
		actions = []string{"base", "finish"}
	}
	var gamestate engine.Gamestate
	for _, force := range forces {
		if force.ID == forceID {
			symbolGrid := generateSymbolGrid(force.StopList, engineID, force.ReelsetId)
			engineDef := engineConf.EngineDefs[force.ReelsetId]
			totalBet := engine.Money{betPerLine.Mul(engine.NewFixedFromInt(engineDef.StakeDivisor)), previousGamestate.BetPerLine.Currency}

			var transactions []engine.WalletTransaction
			transactions = append(transactions, engine.WalletTransaction{Id: previousGamestate.NextGamestate, Type: "WAGER", Amount: totalBet})

			// use engine win type to determine wins
			wins, relativePayout := engineDef.DetermineWins(symbolGrid)
			var nextActions []string
			specialWin := engine.DetermineSpecialWins(symbolGrid, engineDef.SpecialPayouts)
			if specialWin.Index != "" {
				var specialPayout int
				specialPayout, nextActions = engineDef.CalculatePayoutSpecialWin(specialWin)
				relativePayout += specialPayout
				wins = append(wins, specialWin)
			}
			// get Multiplier
			multiplier := 1
			if len(engineDef.Multiplier.Multipliers) > 0 {
				multiplier = engine.SelectFromWeightedOptions(engineDef.Multiplier.Multipliers, engineDef.Multiplier.Probabilities)
			}
			// Build gamestate
			gamestate = engine.Gamestate{Action: force.Action, GameID: fmt.Sprintf("%v:%v", gameID, force.ReelsetId), SymbolGrid: symbolGrid, Prizes: wins, StopList: force.StopList, NextActions: nextActions, Multiplier: multiplier, RelativePayout: relativePayout, Transactions: transactions}
			gamestate.Action = actions[0]
			gamestate.BetPerLine = engine.Money{betPerLine, previousGamestate.BetPerLine.Currency}
			gamestate.SelectedWinLines = previousGamestate.SelectedWinLines
			gamestate.Gamification = previousGamestate.Gamification
			gamestate.UpdateGamification(previousGamestate, gameID)
			gamestate.PrepareActions(actions)
			gamestate.Id = previousGamestate.NextGamestate
			nextID := rng.RandStringRunes(8)
			gamestate.NextGamestate = nextID
			gamestate.PrepareTransactions(previousGamestate, false)

		}
	}

	return gamestate
}
