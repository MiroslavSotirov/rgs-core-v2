package engine

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/features/feature"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type GenerateRound interface {
	ForceRound(EngineDef, GameParams) Gamestate
	FeatureRound(EngineDef, GameParams) Gamestate
	TriggerFeatures(EngineDef, [][]int, []int, GameParams, *feature.FeatureState) feature.FeatureState
}

type GenerateFeatureRound struct {
}

func (gen GenerateFeatureRound) ForceRound(engine EngineDef, parameters GameParams) Gamestate {
	return genForcedRound(gen, engine, parameters)
}

func (gen GenerateFeatureRound) FeatureRound(engine EngineDef, parameters GameParams) Gamestate {
	return genFeatureRound(gen, engine, parameters)
}

func (gen GenerateFeatureRound) TriggerFeatures(engine EngineDef, symbolGrid [][]int,
	stopList []int, parameters GameParams, state *feature.FeatureState) feature.FeatureState {
	return triggerConfiguredFeatures(engine, symbolGrid, stopList, parameters, state)
}

type GenerateFeatureCascade struct {
}

func (gen GenerateFeatureCascade) ForceRound(engine EngineDef, parameters GameParams) Gamestate {
	return genForcedRound(gen, engine, parameters)
}

func (gen GenerateFeatureCascade) FeatureRound(engine EngineDef, parameters GameParams) Gamestate {
	return genFeatureCascade(gen, engine, parameters)
}

func (gen GenerateFeatureCascade) TriggerFeatures(engine EngineDef, symbolGrid [][]int,
	stopList []int, parameters GameParams, state *feature.FeatureState) feature.FeatureState {
	return triggerConfiguredFeatures(engine, symbolGrid, stopList, parameters, state)
}

type GenerateFeatureCascadeMultiply struct {
}

func (gen GenerateFeatureCascadeMultiply) ForceRound(engine EngineDef, parameters GameParams) Gamestate {
	return genForcedRound(gen, engine, parameters)
}

func (gen GenerateFeatureCascadeMultiply) FeatureRound(engine EngineDef, parameters GameParams) Gamestate {
	return genFeatureCascadeMultiply(gen, engine, parameters)
}

func (gen GenerateFeatureCascadeMultiply) TriggerFeatures(engine EngineDef, symbolGrid [][]int,
	stopList []int, parameters GameParams, state *feature.FeatureState) feature.FeatureState {
	return triggerConfiguredFeatures(engine, symbolGrid, stopList, parameters, state)
}

type GenerateStatefulRound struct {
}

func (gen GenerateStatefulRound) ForceRound(engine EngineDef, parameters GameParams) Gamestate {
	return genForcedRound(gen, engine, parameters)
}

func (gen GenerateStatefulRound) FeatureRound(engine EngineDef, parameters GameParams) Gamestate {
	return genFeatureRound(gen, engine, parameters)
}

func (gen GenerateStatefulRound) TriggerFeatures(engine EngineDef, symbolGrid [][]int,
	stopList []int, parameters GameParams, state *feature.FeatureState) feature.FeatureState {
	return triggerStatefulFeatures(engine, symbolGrid, stopList, parameters, state)
}

type GenerateStatefulCascade struct {
}

func (gen GenerateStatefulCascade) ForceRound(engine EngineDef, parameters GameParams) Gamestate {
	return genForcedRound(gen, engine, parameters)
}

func (gen GenerateStatefulCascade) FeatureRound(engine EngineDef, parameters GameParams) Gamestate {
	return genFeatureCascade(gen, engine, parameters)
}

func (gen GenerateStatefulCascade) TriggerFeatures(engine EngineDef, symbolGrid [][]int,
	stopList []int, parameters GameParams, state *feature.FeatureState) feature.FeatureState {
	return triggerStatefulFeatures(engine, symbolGrid, stopList, parameters, state)
}

type GenerateStatefulCascadeMultiply struct {
}

func (gen GenerateStatefulCascadeMultiply) ForceRound(engine EngineDef, parameters GameParams) Gamestate {
	return genForcedRound(gen, engine, parameters)
}

func (gen GenerateStatefulCascadeMultiply) FeatureRound(engine EngineDef, parameters GameParams) Gamestate {
	return genFeatureCascadeMultiply(gen, engine, parameters)
}

func (gen GenerateStatefulCascadeMultiply) TriggerFeatures(engine EngineDef, symbolGrid [][]int,
	stopList []int, parameters GameParams, state *feature.FeatureState) feature.FeatureState {
	return triggerStatefulFeatures(engine, symbolGrid, stopList, parameters, state)
}

func triggerFeatures(engine EngineDef, fs *feature.FeatureState, parameters GameParams) error {
	fs.CalculateWins = func(symbolGrid [][]int, payouts []feature.FeaturePayout) []feature.FeatureWin {
		var wins []Prize
		if len(payouts) == 0 {
			wins, _ = engine.DetermineWins(symbolGrid)
		} else {
			modifiedEngine := engine
			modifiedEngine.Payouts = make([]Payout, len(payouts))
			for i, p := range payouts {
				modifiedEngine.Payouts[i] = Payout{
					Symbol:     p.Symbol,
					Count:      p.Count,
					Multiplier: p.Multiplier,
				}
			}
			wins, _ = modifiedEngine.DetermineWins(symbolGrid)
		}
		featureWins := make([]feature.FeatureWin, len(wins))
		for i, p := range wins {
			featureWins[i] = feature.FeatureWin{
				Index:           p.Index,
				Multiplier:      p.Multiplier,
				Symbols:         []int{p.Payout.Symbol},
				SymbolPositions: p.SymbolPositions,
			}
		}
		return featureWins
	}
	fs.TotalStake = float64(parameters.previousGamestate.BetPerLine.Amount.Mul(NewFixedFromInt(engine.StakeDivisor)).ValueAsFloat())
	fs.ReelsetId = engine.ReelsetId
	fs.Reels = engine.Reels
	fs.Action = parameters.Action

	featureparams := feature.FeatureParams{
		"Engine": engine.ID,
	}
	if config.GlobalConfig.DevMode == true && parameters.Force != "" {
		logger.Debugf("trigger configured features using force %s", parameters.Force)
		featureparams["force"] = parameters.Force
	}
	featuredef := feature.FeatureDef{Features: engine.Features}
	feature.ActivateFeatures(featuredef, fs, featureparams)
	return nil
}

func triggerConfiguredFeatures(engine EngineDef, symbolGrid [][]int, stopList []int, parameters GameParams, state *feature.FeatureState) feature.FeatureState {
	logger.Debugf("Trigger configured features")
	var fs, prevfs feature.FeatureState
	if state != nil {
		fs = *state
	}
	prevfs.SymbolGrid = parameters.previousGamestate.FeatureView
	prevfs.Features = parameters.previousGamestate.Features
	if parameters.Action != "base" {
		fs.Stateless = &prevfs
	}
	fs.SetGrid(symbolGrid)
	fs.StopList = stopList
	if err := triggerFeatures(engine, &fs, parameters); err != nil {
		logger.Errorf("%v", err)
		return feature.FeatureState{}
	}
	return fs
}

func triggerStatefulFeatures(engine EngineDef, symbolGrid [][]int, stopList []int, parameters GameParams, state *feature.FeatureState) feature.FeatureState {
	logger.Debugf("Trigger stateful from previous features %#v", parameters.previousGamestate.Features)
	var fs, prevfs feature.FeatureState
	if state != nil {
		fs = *state
	}
	prevfs.SymbolGrid = parameters.previousGamestate.FeatureView
	prevfs.Features = parameters.previousGamestate.Features
	fs.Stateful = &prevfs
	fs.SetGrid(symbolGrid)
	fs.StopList = stopList
	if err := triggerFeatures(engine, &fs, parameters); err != nil {
		logger.Errorf("%v", err)
		return feature.FeatureState{}
	}
	return fs
}

type filterfunc func(state Gamestate) (bool, error)

func genForcedRound(gen GenerateRound, engine EngineDef, parameters GameParams) Gamestate {
	if parameters.Force != "" {
		logger.Debugf("FeatureRound devmode: %v force: %s IsV3: %v", config.GlobalConfig.DevMode, parameters.Force, config.GlobalConfig.Server.IsV3())
		if config.GlobalConfig.DevMode == true {
			fp := feature.FeatureParams{"force": parameters.Force}
			if config.GlobalConfig.Server.IsV3() {
				filter := fp.GetForce("filter")
				if filter != "" {
					state, err := genFilteredRound(gen, engine, parameters, 100000000,
						func(s Gamestate) (bool, error) {
							js, err := json.Marshal(s)
							if err != nil {
								return false, fmt.Errorf("could not marshal gamestate")
							}
							return strings.Contains(string(js), filter), nil
						})
					if err != nil {
						rgse.Create(rgse.Forcing)
					}
					return state
				}
				minwin := fp.GetForce("minwin")
				if minwin != "" {
					minfp64, err := strconv.ParseFloat(minwin, 64)
					minfixed := NewFixedFromFloat64(minfp64)
					if err != nil {
						rgse.Create(rgse.Forcing)
					}
					state, err := genFilteredRound(gen, engine, parameters, 100000000*100,
						func(s Gamestate) (bool, error) {
							relativePayout := NewFixedFromInt(s.RelativePayout * s.Multiplier)
							win := relativePayout.Mul(parameters.Stake) // s.BetPerLine.Amount)
							//							logger.Infof("win: %d minfixed: %d", win, minfixed)
							return win >= minfixed, nil
						})
					if err != nil {
						rgse.Create(rgse.Forcing)
					}
					return state
				}
			}
			stops := fp.GetForce("stops")
			if stops != "" {
				logger.Infof("force play using stops: \"%s\"", stops)
				stopStrs := strings.Split(stops, ",")
				if len(stopStrs) == len(engine.Reels) {
					stopList := make([]int, len(engine.Reels))
					for i, s := range stopStrs {
						rl := len(engine.Reels[i])
						p := rng.RandFromRange(rl)
						if s != "" {
							p64, err := strconv.ParseInt(s, 10, 64)
							if err != nil {
								logger.Infof("skipping force due to parse error")
								return gen.FeatureRound(engine, parameters) // engine.FeatureRoundGen(parameters)
							}
							p = int(p64) % rl
							if p < 0 {
								p += rl
							}
						}
						stopList[i] = p
					}
					forcedEngine := engine
					forcedEngine.force = stopList
					return gen.FeatureRound(forcedEngine, parameters) // forcedEngine.FeatureRoundGen(parameters)
				} else {
					logger.Infof("skipping force due to wrong number of stop values")
				}
			}
		} else {
			logger.Infof("attempted force on a non devmode server configuration")
			rgse.Create(rgse.Forcing)
		}
	}
	return gen.FeatureRound(engine, parameters) // engine.FeatureRoundGen(parameters)
}

func genFilteredRound(gen GenerateRound, engine EngineDef, parameters GameParams, timeout int64, filter filterfunc) (Gamestate, error) {
	startTime := time.Now()
	for true {
		state := gen.FeatureRound(engine, parameters)
		satisfied, err := filter(state)
		if err != nil {
			logger.Errorf("filter halted due to error %v", err)
			return state, err
		}
		if satisfied {
			logger.Infof("filter force was satisfied")
			return state, nil
		}
		elapsed := time.Now().Sub(startTime)
		if elapsed > time.Duration(timeout) {
			logger.Warnf("filter halted due to timeout")
			return state, nil
		}
	}
	return Gamestate{}, nil
}

func genFeatureRound(gen GenerateRound, engine EngineDef, parameters GameParams) Gamestate {
	// the base gameplay round
	// uses round multiplier if included
	// no dynamic reel calculation

	var wl []int
	wl, engine = engine.ProcessWinLines(parameters.SelectedWinLines)

	// spin
	symbolGrid, stopList := engine.Spin()

	// replace any symbols with sticky wilds
	symbolGrid = engine.addStickyWilds(parameters.previousGamestate, symbolGrid)

	featurestate := gen.TriggerFeatures(engine, symbolGrid, stopList, parameters, nil)
	logger.Debugf("symbolGrid= %v featureGrid= %v", symbolGrid, featurestate.SymbolGrid)
	engine.Reels = featurestate.Reels

	var nextActions []string
	wins, relativePayout := engine.DetermineWins(featurestate.SymbolGrid)
	featureWins, featureRelPayout, featureNextActions := engine.convertFeaturePrizes(featurestate.Wins)
	relativePayout += featureRelPayout
	wins = append(wins, featureWins...)
	nextActions = append(featureNextActions, nextActions...)
	// calculate specialWin
	specialWin := DetermineSpecialWins(featurestate.SymbolGrid, engine.SpecialPayouts)
	if specialWin.Index != "" {
		var specialPayout int
		specialPayout, nextActions = engine.CalculatePayoutSpecialWin(&specialWin)
		relativePayout += specialPayout
		wins = append(wins, specialWin)
	}
	logger.Debugf("got %v wins: %v", len(wins), wins)
	// get Multiplier
	multiplier := 1
	if len(engine.Multiplier.Multipliers) > 0 {
		multiplier = SelectFromWeightedOptions(engine.Multiplier.Multipliers, engine.Multiplier.Probabilities)
	}
	// if no features were generated then no need to store a featureview
	if len(featurestate.Features) == 0 {
		featurestate.SymbolGrid = nil
	}

	// Build gamestate
	gamestate := Gamestate{
		DefID:            engine.Index,
		Prizes:           wins,
		SymbolGrid:       symbolGrid,
		RelativePayout:   relativePayout,
		Multiplier:       multiplier,
		StopList:         stopList,
		NextActions:      nextActions,
		SelectedWinLines: wl,
		Features:         featurestate.Features,
		FeatureView:      featurestate.SymbolGrid,
		ReelsetID:        featurestate.ReelsetId,
	}

	return gamestate
}

func genFeatureCascade(gen GenerateRound, engine EngineDef, parameters GameParams) Gamestate {
	var reelsetId string
	var cascadePositions []int
	var featureState feature.FeatureState
	symbolGrid, stopList := engine.Spin()
	cascade := strings.Contains(parameters.Action, "cascade")
	if cascade {
		previousGamestate := parameters.previousGamestate

		if len(engine.FeatureStages) == 0 || ContainsString(engine.FeatureStages, "reelupdate") {
			featureState = gen.TriggerFeatures(engine, previousGamestate.SymbolGrid, previousGamestate.StopList, parameters, nil)
			logger.Debugf("update to reelset %s", featureState.ReelsetId)
			engine.Reels = featureState.Reels
			reelsetId = featureState.ReelsetId
		}

		// if previous gamestate contains a win, we need to cascade new tiles into the old space
		previousGrid := previousGamestate.FeatureView // previousGamestate.SymbolGrid
		remainingGrid := [][]int{}
		//determine map of symbols to disappear
		for i := 0; i < len(previousGamestate.Prizes); i++ {
			for j := 0; j < len(previousGamestate.Prizes[i].SymbolPositions); j++ {
				// to avoid having to assume symbol grid is regular:
				col := 0
				row := 0

				for ii := 0; ii < previousGamestate.Prizes[i].SymbolPositions[j]; ii++ {
					row++
					if row >= len(previousGamestate.SymbolGrid[col]) {
						col++
						row = 0
					}
				}
				previousGrid[col][row] = -1
			}
		}

		for i := 0; i < len(previousGrid); i++ {
			remainingReel := []int{}
			// remove winning symbols from the view
			for j := 0; j < len(previousGrid[i]); j++ {
				if previousGrid[i][j] >= 0 {
					remainingReel = append(remainingReel, previousGrid[i][j])
				}
			}
			remainingGrid = append(remainingGrid, remainingReel)
		}

		// adjust reels in case need to cascade symbols from the end of the strip
		adjustedReels := make([][]int, len(engine.Reels))

		for i := 0; i < len(engine.Reels); i++ {
			adjustedReels[i] = append(engine.Reels[i], engine.Reels[i][:engine.ViewSize[i]]...)
		}

		_, stopList = engine.Spin()
		// return grid to full size by filling in empty spaces
		for i := 0; i < len(engine.ViewSize); i++ {
			numToAdd := engine.ViewSize[i] - len(remainingGrid[i])
			cascadePositions = append(cascadePositions, numToAdd)
			stop := stopList[i] - numToAdd
			// get adjusted index if the previous win was at the top of the reel
			if stop < 0 {
				stop = len(engine.Reels[i]) + stop
			}
			symbolGrid[i] = append(adjustedReels[i][stop:stop+numToAdd], remainingGrid[i]...)
			stopList[i] = stop
		}
	}

	multiplier := parameters.previousGamestate.Multiplier
	logger.Debugf("previous multiplier: %d", multiplier)

	// get first Multiplier
	if multiplier == 0 && len(engine.Multiplier.Multipliers) > 0 {
		//multiplier = SelectFromWeightedOptions(engine.Multiplier.Multipliers, engine.Multiplier.Probabilities)
		multiplier = engine.Multiplier.Multipliers[0]
		logger.Debugf("initiating multiplier: %d", multiplier)
	}

	featureState = gen.TriggerFeatures(engine, symbolGrid, stopList, parameters,
		&feature.FeatureState{CascadePositions: cascadePositions, Multiplier: multiplier})
	multiplier = featureState.Multiplier
	logger.Debugf("updated multiplier: %d", multiplier)
	if cascade {
		logger.Debugf("cascade positions: %v", cascadePositions)
		featureState.Reels = engine.Reels
		featureState.ReelsetId = reelsetId
	}

	/*
		logger.Debugf("symbolGrid= %v featureGrid= %v", symbolGrid, featureState.SymbolGrid)
		logger.Debugf("update reels after feature activation.")
		logger.Debugf("from %v", engine.Reels)
		logger.Debugf("to %v", featureState.Reels)
		engine.Reels = featureState.Reels
	*/
	// calculate wins
	wins, relativePayout := engine.DetermineWins(featureState.SymbolGrid)
	// calculate specialWin
	var nextActions []string
	cascade = false
	// if any win is present, next action should be cascade
	if len(wins) > 0 {
		logger.Debugf("cascade is true")
		cascade = true
	} else {
		// only check for special win after cascading has completed
		logger.Debugf("determining special wins: %#v", engine.SpecialPayouts)
		specialWin := DetermineSpecialWins(featureState.SymbolGrid, engine.SpecialPayouts)
		if specialWin.Index != "" {
			var specialPayout int
			specialPayout, nextActions = engine.CalculatePayoutSpecialWin(&specialWin)
			relativePayout += specialPayout
			wins = append(wins, specialWin)
		}
	}
	if cascade {
		respinAction := "cascade"
		if engine.RespinAction != "" {
			respinAction = engine.RespinAction
		}
		nextActions = append([]string{respinAction}, nextActions...)
	}
	featureWins, featureRelPayout, featureNextActions := engine.convertFeaturePrizes(featureState.Wins)
	relativePayout += featureRelPayout
	wins = append(wins, featureWins...)
	if func() bool {
		for _, w := range featureWins {
			if strings.Contains(w.Index, "finish") {
				return true
			}
		}
		return false
	}() {
		logger.Debugf("feature win found of type finish. end round")
		nextActions = []string{}
	} else {
		nextActions = append(nextActions, featureNextActions...)
	}

	// for now, do a bit of a hack to get the cascade positions. as soon as we need to implement cascade with variable
	// win lines, this will need to be adjusted to be added into a new field in the gamestate message

	// Build gamestate
	gamestate := Gamestate{
		DefID:            engine.Index,
		Prizes:           wins,
		SymbolGrid:       symbolGrid,
		RelativePayout:   relativePayout,
		Multiplier:       multiplier,
		StopList:         stopList,
		NextActions:      nextActions,
		SelectedWinLines: cascadePositions, // resuse hack from regular cascade to encode positions in winlines
		Features:         featureState.Features,
		FeatureView:      featureState.SymbolGrid,
		ReelsetID:        featureState.ReelsetId,
	}
	return gamestate
}

func genFeatureCascadeMultiply(gen GenerateRound, engine EngineDef, parameters GameParams) Gamestate {
	// multiplier increments for each cascade, up to the highest multiplier in the engine
	gamestate := genFeatureCascade(gen, engine, parameters)
	// get next multiplier
	prevIndex := getIndex(parameters.previousGamestate.Multiplier, engine.Multiplier.Multipliers)
	logger.Debugf("multiplier %v index %v in %v action %s",
		parameters.previousGamestate.Multiplier, prevIndex, engine.Multiplier.Multipliers, parameters.Action)

	nextMultiplierActions := engine.NextMultiplierActions
	holdMultiplierActions := engine.HoldMultiplierActions
	if len(nextMultiplierActions) == 0 && len(holdMultiplierActions) == 0 {
		nextMultiplierActions = []string{"cascade"}
		holdMultiplierActions = []string{"freespin"}
	}

	if ContainsString(nextMultiplierActions, parameters.Action) {
		logger.Debugf("next multiplier action")
		if prevIndex < 0 {
			// fallback to first multiplier
			logger.Debugf("fallback to first multiplier")
			gamestate.Multiplier = engine.Multiplier.Multipliers[0]
		} else {
			gamestate.Multiplier = engine.Multiplier.Multipliers[minInt(prevIndex+1, len(engine.Multiplier.Multipliers)-1)]
		}
	} else if ContainsString(holdMultiplierActions, parameters.Action) {
		logger.Debugf("hold multiplier action")
		if prevIndex < 0 {
			// fallback to first multiplier
			gamestate.Multiplier = engine.Multiplier.Multipliers[0]
		} else {
			gamestate.Multiplier = engine.Multiplier.Multipliers[prevIndex]
		}
		logger.Debugf("freespin with same multiplier %v and index %v as last round", gamestate.Multiplier, prevIndex)
	} else {
		logger.Debugf("use first multiplier")
		gamestate.Multiplier = engine.Multiplier.Multipliers[0]
	}

	return gamestate
}

func (engine EngineDef) FeatureRound(parameters GameParams) Gamestate {
	return GenerateFeatureRound{}.ForceRound(engine, parameters)
}

func (engine EngineDef) FeatureCascade(parameters GameParams) Gamestate {
	return GenerateFeatureCascade{}.ForceRound(engine, parameters)
}

func (engine EngineDef) FeatureCascadeMultiply(parameters GameParams) Gamestate {
	return GenerateFeatureCascadeMultiply{}.ForceRound(engine, parameters)
}

func (engine EngineDef) StatefulRound(parameters GameParams) Gamestate {
	return GenerateStatefulRound{}.ForceRound(engine, parameters)
}

func (engine EngineDef) StatefulCascade(parameters GameParams) Gamestate {
	return GenerateStatefulCascade{}.ForceRound(engine, parameters)
}

func (engine EngineDef) StatefulCascadeMultiply(parameters GameParams) Gamestate {
	return GenerateStatefulCascadeMultiply{}.ForceRound(engine, parameters)
}

func (engine EngineDef) InitRound(parameters GameParams) (state Gamestate) {
	stopList := make([]int, len(engine.ViewSize))
	for i := range stopList {
		stopList[i] = rng.RandFromRange(len(engine.Reels[i]))
	}
	state.SymbolGrid = GetSymbolGridFromStopList(engine.Reels, engine.ViewSize, stopList)
	engine.InitRoundFeatures(parameters, stopList, &state)

	logger.Debugf("Init state: %#v", state)
	return
}

func (engine EngineDef) InitRoundNoSpin(parameters GameParams) (state Gamestate) {
	logger.Debugf("InitRoundNoSpin with ViewSize %v", engine.ViewSize)
	stopList := make([]int, len(engine.ViewSize))
	for i := range stopList {
		stopList[i] = 0
	}
	state.SymbolGrid = GetSymbolGridFromStopList(engine.Reels, engine.ViewSize, stopList)
	engine.InitRoundFeatures(parameters, stopList, &state)
	return
}

func (engine EngineDef) InitRoundFeatures(parameters GameParams, stopList []int, state *Gamestate) {
	featuredef := feature.FeatureDef{Features: engine.Features}
	var fs feature.FeatureState
	fs.SetGrid(state.SymbolGrid)
	fs.StopList = stopList
	fs.ReelsetId = engine.ReelsetId
	fs.Reels = engine.Reels
	fs.Action = parameters.Action
	feature.InitFeatures(featuredef, &fs)
	state.Features = fs.Features
	state.ReelsetID = fs.ReelsetId
	return
}

func (engine EngineDef) convertFeaturePrizes(featureWins []feature.FeatureWin) (wins []Prize, relativePayout int, nextActions []string) {
	wins = []Prize{}
	relativePayout = 0
	nextActions = []string{}
	for _, w := range featureWins {
		if w.Index == "" {
			symbol := 0
			index := "0:0"
			if len(w.Symbols) > 0 {
				index = fmt.Sprintf("%d:%d", w.Symbols[0], len(w.Symbols))
			}
			prize := Prize{
				Payout: Payout{
					Symbol:     symbol,
					Count:      len(w.Symbols),
					Multiplier: engine.StakeDivisor,
				},
				Index:           index,
				Multiplier:      w.Multiplier,
				SymbolPositions: w.SymbolPositions,
				Winline:         -1, // until features have prizes associated with lines
			}
			wins = append(wins, prize)
		} else {
			prize := Prize{
				Payout: Payout{
					Symbol:     0,
					Count:      len(w.Symbols),
					Multiplier: engine.StakeDivisor, // 1, // w.Multiplier, // engine.StakeDivisor,
				},
				Index:           w.Index,
				Multiplier:      w.Multiplier,
				SymbolPositions: w.SymbolPositions,
				Winline:         -1, // until features have prizes associated with lines
			}
			sp, na := engine.CalculatePayoutSpecialWin(&prize)
			if sp > 0 {
				//				panic(fmt.Sprintf("feature prize is counted double %#v", prize))
				//				relativePayout += sp
				logger.Debugf("feature prize special win is zeroed to not pay double %#v", prize)
			}
			nextActions = append(na, nextActions...)
			logger.Debugf("Adding special payout: %v with actions: %v to final action list: %v", sp, na, nextActions)
			wins = append(wins, prize)
		}
	}
	relativePayout += calculatePayoutWins(wins)
	return
}
