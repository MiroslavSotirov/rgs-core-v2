package engine

// Spin and play functions of engines

import (
	"errors"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"reflect"
	"strconv"
	"strings"
)

func (engine EngineDef) Spin() ([][]int, []int) {
	// Performs a random spin of the reels
	if len(engine.ViewSize) != len(engine.Reels) {
		logger.Fatalf("ViewSize (%v) and Reel size (%v) mismatch", len(engine.ViewSize), len(engine.Reels))
		return [][]int{}, []int{}
	}
	stopList := make([]int, len(engine.ViewSize))

	for index, reel := range engine.Reels {
		// choose a random index on the reel
		reelIndex := rng.RandFromRange(len(reel))
		stopList[index] = reelIndex
	}
	if config.GlobalConfig.DevMode == true && len(engine.force) == len(engine.ViewSize) {
		stopList = engine.force
		rgse.Create(rgse.Forcing)
		//logger.Warnf("forcing engine %v", engine.ID)
	}
	symbolGrid := GetSymbolGridFromStopList(engine.Reels, engine.ViewSize, stopList)
	return symbolGrid, stopList
}

func GetSymbolGridFromStopList(reels [][]int, viewSize []int, stopList []int) [][]int {
	symbolGrid := make([][]int, len(viewSize))
	for i, reel := range reels {
		// Add first symbols to the end of reel in case final symbols are chosen
		reelEquiv := append(reel, reel[0:viewSize[i]]...)
		symbolGrid[i] = reelEquiv[stopList[i] : stopList[i]+viewSize[i]]
	}
	return symbolGrid
}

func getNewWildMultiplier(previousMultiplier int, newMultiplier int, compounding bool) int {
	if compounding {
		return previousMultiplier * newMultiplier
	} else {
		if newMultiplier > previousMultiplier {
			return newMultiplier
		} else {
			return previousMultiplier
		}
	}
}

func DetermineLineWinsAnywhere(symbolGrid [][]int, WinLines [][]int, linePayouts []Payout, wilds []wild, compounding bool) (lineWins []Prize) {
	// this function determines line wins not necessarily starting at the first symbol
	// only one win per symbol per line is permitted

	wildMultipliers := make(map[int]int)

	for winLineIndex, winLine := range WinLines {
		// because we can have multiple wins per line, but we can only have one line per win per symbol, we need to keep track of which symbols have been matched already
		var matchedSymbols []int

		lineContent := make([]int, len(symbolGrid))
		symbolPositions := make([]int, len(symbolGrid))
		for reel, index := range winLine {
			lineContent[reel] = symbolGrid[reel][index]
			// to determine a unique symbol position per location in the view, this only works if view is regularly sized (i.e. all reels show same number of symbols. todo: make this dynamic for all reel configurations
			symbolPositions[reel] = len(symbolGrid[reel])*reel + index
		}

		// iterate through the line content
		for i:=0; i<len(lineContent)-1; i++ {
			// adjust payouts to exclude any symbols already matched

			var adjustedPayouts []Payout
			for j:=0; j<len(linePayouts); j++ {
				seen := false
				for k:=0; k<len(matchedSymbols); k++{
					if linePayouts[j].Symbol == matchedSymbols[k] {
						seen = true
						break
					}
				}
				if !seen {
					adjustedPayouts = append(adjustedPayouts, linePayouts[j])
				}
			}
			win := GetWinInLine(lineContent[i:], wilds, adjustedPayouts, compounding, wildMultipliers)
			if win.Index == "" {
				continue
			}
			// add any new wins to matchedSymbols
			matchedSymbols = append(matchedSymbols, win.Payout.Symbol)
			win.SymbolPositions = symbolPositions[i:i+win.Payout.Count]
			win.Winline = winLineIndex
			lineWins = append(lineWins, win)
		}
	}
	return lineWins
}

func GetWinInLine(lineContent []int, wilds []wild, linePayouts []Payout, compounding bool, wildMultipliers map[int]int) (prize Prize) {
	// wildMultipliers is the list of multipliers already defined for variable wilds. if each wild multiplier is meant
	// to be determined independently regardless of previous setting, wildMultipliers should be empty

	numMatch := 1
	lineSymbol := lineContent[0]
	multiplier := 1

	// Determine how many consecutive symbols are in the line
	for _, symbol := range lineContent[1:] {
		match := false
		if symbol == lineSymbol {
			numMatch++
			match = true
		} else {
			// process wilds
			// NB Multipliers override here. i.e. a line with 4 wilds of multiplier 2 will have total multiplier 2, not 16
			for _, engineWild := range wilds {
				if lineSymbol == engineWild.Symbol {
					// if the lineSymbol is a wild then all symbols so far have been wild, and this one is not, so update lineSymbol
					// if all 5 symbols are wild then no multiplier will be added. this should be handled in the payout
					lineSymbol = symbol
					// multipliers do not compound, there may be only one line multiplier
					engineWildMultiplier, ok := wildMultipliers[engineWild.Symbol]
					if !ok {
						engineWildMultiplier = SelectFromWeightedOptions(engineWild.Multiplier.Multipliers, engineWild.Multiplier.Probabilities)
						wildMultipliers[engineWild.Symbol] = engineWildMultiplier
					}

					multiplier = getNewWildMultiplier(multiplier, engineWildMultiplier, compounding)

					numMatch++
					// stop checking for wilds if one has been matched
					match = true
					break
				} else if symbol == engineWild.Symbol {
					// if the new symbol on the line is a wild count it as a match
					numMatch++
					engineWildMultiplier, ok := wildMultipliers[engineWild.Symbol]
					if !ok {
						engineWildMultiplier = SelectFromWeightedOptions(engineWild.Multiplier.Multipliers, engineWild.Multiplier.Probabilities)
						wildMultipliers[engineWild.Symbol] = engineWildMultiplier
					}

					multiplier = getNewWildMultiplier(multiplier, engineWildMultiplier, compounding)

					match = true
					break
				}
			}
		}
		if match == false {
			// stop checking the line once a non-matching symbol is found
			break
		}
	}

	// Compare best run to Payouts, not duplicating any symbols
	for _, payout := range linePayouts {
		if lineSymbol == payout.Symbol && numMatch == payout.Count {
			//copy payout object, NB can only do because no reference fields in Payout
			linePayout := payout
			prize = Prize{Payout: linePayout, Index: fmt.Sprintf("%v:%v", lineSymbol, numMatch), Multiplier: multiplier}
			// Only one win possible per line, earliest payout in payout dict takes precedence
			break
		}
	}
	return prize
}


func DetermineLineWins(symbolGrid [][]int, WinLines [][]int, linePayouts []Payout, wilds []wild, compounding bool) (lineWins []Prize) {
	// determines prizes from line wins including wilds with multipliers
	// highest wild multiplier takes precedence for multiple wilds on the same line (i.e. wild multipliers do not compound)

	// store wild multiplier selections if they are to be reused for future wilds
	wildMultipliers := make(map[int]int)

	for winLineIndex, winLine := range WinLines {
		lineContent := make([]int, len(symbolGrid))
		symbolPositions := make([]int, len(symbolGrid))
		for reel, index := range winLine {
			lineContent[reel] = symbolGrid[reel][index]
			// to determine a unique symbol position per location in the view, this only works if view is regularly sized (i.e. all reels show same number of symbols. todo: make this dynamic for all reel configurations
			symbolPositions[reel] = len(symbolGrid[reel])*reel + index
		}

		// wildMultipliers is passed in and modulated, we get to keep the results of the modulation in the next rounds of the for loop
		win := GetWinInLine(lineContent, wilds, linePayouts, compounding, wildMultipliers)
		if win.Index == "" {
			continue
		}

		win.SymbolPositions = symbolPositions[:win.Payout.Count]
		win.Winline = winLineIndex
		lineWins = append(lineWins, win)
	}

	return lineWins
}

func determineBarLineWins(symbolGrid [][]int, winLines [][]int, payouts []Payout, bars []bar, wilds []wild, compoundingMultipliers bool) []Prize {
	// assume no symbol is included in two bar types
	adjustedSymbolGrid := make([][]int, len(symbolGrid))
	for _, bar := range bars {
		for i, row := range symbolGrid {
			adjustedRow := make([]int, len(row))
			for j, symbol := range row {
				for _, barSymbol := range bar.Symbols {
					if symbol == barSymbol {
						adjustedRow[j] = bar.PayoutID
						break
					}
					adjustedRow[j] = symbol
				}
			}
			adjustedSymbolGrid[i] = adjustedRow
		}
	}
	lineWinsWithBar := DetermineLineWins(adjustedSymbolGrid, winLines, payouts, wilds, compoundingMultipliers)
	lineWinsWithoutBar := DetermineLineWins(symbolGrid, winLines, payouts, wilds, compoundingMultipliers)
	var highestWinsPerLine []Prize

	for _, barWin := range lineWinsWithBar {
		// wins with bars included should be larger set, should include all wins without bars
		matched := false
		for _, noBarWin := range lineWinsWithoutBar {
			// if line is included already
			if barWin.Winline == noBarWin.Winline {
				// check which prize is higher
				matched = true
				if barWin.Payout.Multiplier > noBarWin.Payout.Multiplier {
					highestWinsPerLine = append(highestWinsPerLine, barWin)
				} else {
					highestWinsPerLine = append(highestWinsPerLine, noBarWin)
				}
				break // assume only one win per winline
			}
		}
		if matched == false {
			highestWinsPerLine = append(highestWinsPerLine, barWin)
		}
	}
	return highestWinsPerLine
}

type wayWin struct {
	multiplier      int
	symbolPositions []int
	isAllWild       bool
}

func updateSymbols(variation wayWin, newSymbol int) (symbols []int) {
	// we have to do this instead of using append because of the way go represents slices
	symbols = make([]int, len(variation.symbolPositions)+1)
	for i := 0; i < len(variation.symbolPositions); i++ {
		symbols[i] = variation.symbolPositions[i]
	}
	symbols[len(variation.symbolPositions)] = newSymbol
	return
}

// DetermineWaysWins ...
func DetermineWaysWins(symbolGrid [][]int, waysPayouts []Payout, wilds []wild) []Prize {
	// Input :: symbolGrid 2d 2dslice
	// Output :: slice of prize structs
	var waysWins []Prize
	var matchedSymbols []int

	// Iterate through waysPayouts to see if a match can be found
	for _, waysPayout := range waysPayouts {
		alreadyMatched := false
		for _, symbol := range matchedSymbols {
			// if a waysPayout with the same symbol has already been matched, continue (3x symbol payout will not return match in line of 5x symbol)
			if waysPayout.Symbol == symbol {
				alreadyMatched = true
				break
			}
		}
		if alreadyMatched == true {
			continue
		}
		var variations []wayWin
		symbolIndex := 0
		for _, reel := range symbolGrid {
			// on each reel, check if a symbol matching the payout or a wild is in the view
			var extendableVariations []wayWin
			match := false
			for _, symbol := range reel {
				if waysPayout.Symbol == symbol {
					if len(variations) == 0 {
						// on first round, we need to add an initial variation for each symbol that matches this win type in the first reel
						newVariation := wayWin{1, []int{symbolIndex}, false}
						extendableVariations = append(extendableVariations, newVariation)
					}
					for _, variation := range variations {
						// for each win that can be extended, add this variation
						symbols := updateSymbols(variation, symbolIndex)
						newVariation := wayWin{variation.multiplier, symbols, false}
						extendableVariations = append(extendableVariations, newVariation)
					}
					match = true
				} else {
					//check against wilds
					for _, engineWild := range wilds {
						if symbol == engineWild.Symbol {
							match = true
							engineWildMultiplier := SelectFromWeightedOptions(engineWild.Multiplier.Multipliers, engineWild.Multiplier.Probabilities)
							if len(variations) == 0 {
								// on first round, we need to add an initial variation
								newVariation := wayWin{engineWildMultiplier, []int{symbolIndex}, true}
								extendableVariations = append(extendableVariations, newVariation)
							}
							for _, variation := range variations {
								// for each win that can be extended, add this symbol
								multiplier := variation.multiplier
								if engineWildMultiplier > multiplier {
									multiplier = engineWildMultiplier
								}
								symbols := updateSymbols(variation, symbolIndex)
								newVariation := wayWin{multiplier, symbols, variation.isAllWild}
								extendableVariations = append(extendableVariations, newVariation)
							}
							break
						}
					}
				}
				symbolIndex++
			}
			if match == false {
				// when no matches on a reel, check variations against prize
				break
			} else {
				variations = extendableVariations
			}
		}
		for i := 0; i < len(variations); i++ {
			variation := variations[i]
			if len(variation.symbolPositions) == waysPayout.Count {
				// ignore variations consisting of only wild symbols (isAllWild will be false if matching hard-coded win for 5 wilds)
				if variation.isAllWild == true {
					continue
				}
				winIndex := strconv.Itoa(waysPayout.Symbol) + ":" + strconv.Itoa(waysPayout.Count)
				// copy waysPayout to avoid reference changing NB: we can only do this because no reference types in Payout struct
				prizePayout := waysPayout
				waysWins = append(waysWins, Prize{Payout: prizePayout, Index: winIndex, Multiplier: variation.multiplier, SymbolPositions: variation.symbolPositions}) //, Winline: -1})
				matchedSymbols = append(matchedSymbols, waysPayout.Symbol)
			}
		}

	}
	return waysWins
}

func TransposeGrid(symbolGrid [][]int) [][]int {
	if len(symbolGrid) == 0 {
		return [][]int{}
	}
	newGrid := make([][]int, len(symbolGrid[0]))
	for _, row := range symbolGrid {
		for j, symbol := range row {
			newGrid[j] = append(newGrid[j], symbol)
		}
	}
	return newGrid
}

func GetSymbolPositions(symbolGrid [][]int, symbol int) []int {
	symbolPositions := make([]int, 0)
	//view := TransposeGrid(symbolGrid)
	view := symbolGrid
	pos := -1
	for _, reel := range view {
		for _, rsymbol := range reel {
			pos++
			if rsymbol == symbol {
				symbolPositions = append(symbolPositions, pos)
			}
		}
	}
	return symbolPositions
}

// DetermineSpecialWins ...
func DetermineSpecialWins(symbolGrid [][]int, specialPayouts []Prize) Prize {

	// Looks for matching symbols in any position
	// returns at first match, so highest-ranked prize will override any others
	for _, specialPayout := range specialPayouts {
		count := 0
		for _, reel := range symbolGrid {
			for _, symbol := range reel {
				if symbol == specialPayout.Payout.Symbol {
					count++
				}
			}
		}

		if count == specialPayout.Payout.Count {
			specialPayout.SymbolPositions = GetSymbolPositions(symbolGrid, specialPayout.Payout.Symbol)
			//specialPayout.Winline = -1
			return specialPayout
		}
	}
	return Prize{}
}

func determinePrimeAndFlopWins(symbolGrid [][]int, payouts []Payout, wilds []wild) []Prize {
	// by default, prime is first reel
	// check if symbol grid is more than one row, if so throw error
	// win only if any symbols on flop match the prime symbol
	if len(symbolGrid[0]) > 1 || len(symbolGrid) < 3 {
		panic(errors.New("too many rows for this engine type or not enough columns"))
	}

	prime := symbolGrid[0][0]
	multiplier := 1

	for w := 0; w < len(wilds); w++ {
		if prime == wilds[w].Symbol {
			logger.Debugf("prime is wild, choose highest-paying combo")
			prime = symbolGrid[1][0]
			for s := 2; s < len(symbolGrid); s++ {
				if symbolGrid[s][0] > prime {
					prime = symbolGrid[s][0]
				}
			}
		}
	}
	logger.Debugf("prime is %v", prime)
	numMatch := 1
	winLocations := []int{0}
	for i := 1; i < len(symbolGrid); i++ {
		symbol := symbolGrid[i][0]
		if symbol == prime {
			numMatch++
			logger.Debugf("got a win")
			winLocations = append(winLocations, i)
		} else {
			// check if symbol is a wild
			for w := 0; w < len(wilds); w++ {
				if wilds[w].Symbol == symbol {
					// this is  a wild win
					numMatch++
					logger.Debugf("got a wild win")
					winLocations = append(winLocations, i)
					mulW := SelectFromWeightedOptions(wilds[w].Multiplier.Multipliers, wilds[w].Multiplier.Probabilities)
					if multiplier < mulW {
						multiplier = mulW
					}
				}
			}
		}
	}
	logger.Debugf("symbol %v, num %v", prime, numMatch)
	for _, payout := range payouts {
		if prime == payout.Symbol && numMatch == payout.Count {
			logger.Debugf("got a match for win %v", payout)
			// copy payout NB can only do because no reference fields in Payout
			pfPayout := payout
			return []Prize{{Payout: pfPayout, Index: strconv.Itoa(prime) + ":" + strconv.Itoa(numMatch), Multiplier: multiplier, Winline: 0, SymbolPositions: winLocations}}
			// Only one win possible per line, earliest payout in payout dict takes precedence
		}
	}
	return []Prize{}
}

// Play ...
func Play(previousGamestate Gamestate, betPerLine Fixed, currency string, parameters GameParams) (Gamestate, EngineConfig) {
	logger.Debugf("Playing round with parameters: %#v", parameters)

	engineConf, err := previousGamestate.Engine()
	if err != nil {
		return Gamestate{}, EngineConfig{}
	}
	var totalBet Money
	chargeWager := true
	var actions []string

	if len(previousGamestate.NextActions) == 1 && previousGamestate.NextActions[0] == "finish" {
		// the old game round should be closed and a new round started

		// if this is a respin, special case:
		if parameters.Action == "respin" {
			// if action is respin, WAGER is dependent on reel configuration
			// index of the reel to be respun must be passed
			if parameters.RespinReel < 0 {
				logger.Errorf("ERROR, NO RESPIN REEL INDEX PASSED")
				return Gamestate{}, EngineConfig{}
			}
			actions = []string{parameters.Action, "finish"}
			betPerLine = previousGamestate.BetPerLine.Amount
			totalBet = RoundUpToNearestCCYUnit(Money{previousGamestate.RespinPriceReel(parameters.RespinReel), currency})
			parameters.previousGamestate = previousGamestate
		} else if parameters.Action == "gamble" {
			// verify that the previous action was freespin and nextaction is finish
			if !(strings.Contains(previousGamestate.Action, "freespin") && len(previousGamestate.NextActions) == 1 && previousGamestate.NextActions[0] == "finish") {
				// this is not allowed
				logger.Errorf("ERROR, NOT A VALID GAMBLE ROUND")
				return Gamestate{}, EngineConfig{}
			}
			if parameters.RespinReel < 0 {
				logger.Errorf("ERROR, NO GAMBLE INDEX PASSED")
				return Gamestate{}, EngineConfig{}
			}
			actions = []string{fmt.Sprintf("%v%v",parameters.Action, parameters.RespinReel), "finish"}
			parameters.Action = actions[0]
			betPerLine = previousGamestate.CumulativeWin
			totalBet = Money{previousGamestate.CumulativeWin, currency}

		} else {
			// new gameplay round
			// totalbet is set after gameplay
			actions = []string{parameters.Action, "finish"}
		}
	} else {
		chargeWager = false
		logger.Debugf("Continuing game round, no WAGER charged")
		actions = previousGamestate.NextActions
		parameters.Action = actions[0] // in theory this should be passed in by the client, once all games migrated to v2 do a check for this
		betPerLine = previousGamestate.BetPerLine.Amount
		parameters.previousGamestate = previousGamestate
		parameters.SelectedWinLines = previousGamestate.SelectedWinLines
	}

	var method reflect.Value
	switch parameters.Action {
	case "cascade":
		// action must be performed on the same engine as previous round
		method = reflect.ValueOf(engineConf.EngineDefs[previousGamestate.DefID]).MethodByName(engineConf.EngineDefs[previousGamestate.DefID].Function)
	case "respin":
		// action must be performed on the same engine as previous round, but method will always be respin
		method = reflect.ValueOf(engineConf.EngineDefs[previousGamestate.DefID].Respin)
	//case "gamble":
	//	// action must be performed on the gamble engine
	//	method = reflect.ValueOf(engineConf.EngineDefs[previousGamestate.DefID]).MethodByName(engineConf.EngineDefs[previousGamestate.DefID].Function)
	default:
		method, err = engineConf.getEngineAndMethod(parameters.Action)
	}

	if err != nil {
		panic(err)
	}

	gamestateAndNextActions := method.Call([]reflect.Value{reflect.ValueOf(parameters)})

	gamestate, ok := gamestateAndNextActions[0].Interface().(Gamestate)
	if !ok {
		panic("value not a gamestate")
	}
	gamestate.PostProcess(previousGamestate, chargeWager, totalBet, engineConf, betPerLine, actions, currency)
	return gamestate, engineConf
}

func (gamestate *Gamestate) PostProcess(previousGamestate Gamestate, chargeWager bool, totalBet Money, engineConf EngineConfig, betPerLine Fixed, actions []string, currency string) {
	// separated for forcetool consistency
	if chargeWager {
		if totalBet.Currency == "" {
			sd := engineConf.EngineDefs[0].StakeDivisor
			if len(gamestate.SelectedWinLines) > 0 {
				sd = len(gamestate.SelectedWinLines)
			}
			totalBet = Money{Amount: betPerLine.Mul(NewFixedFromInt(sd)), Currency: currency}
		}
		gamestate.Transactions = []WalletTransaction{{
			Id:     previousGamestate.NextGamestate,
			Amount: totalBet,
			Type:   "WAGER",
		}}
	}
	if gamestate.Action == "" {
		// allow the function to set its own action
		gamestate.Action = actions[0]
	}
	gamestate.BetPerLine = Money{betPerLine, currency}
	gamestate.PrepareActions(actions)
	gamestate.Gamification = previousGamestate.Gamification
	gamestate.Game = previousGamestate.Game
	gamestate.UpdateGamification(previousGamestate)

	gamestate.Id = previousGamestate.NextGamestate
	gamestate.PreviousGamestate = previousGamestate.Id

	nextID := uuid.NewV4().String()
	gamestate.NextGamestate = nextID
	gamestate.PrepareTransactions(previousGamestate)
	logger.Debugf("gamestate: %#v", gamestate)
	return
}

func (engineConf EngineConfig) DetectSpecialWins(defIndex int, p Prize) string {
	winId := p.Index
	for _, specialPayout := range engineConf.EngineDefs[defIndex].SpecialPayouts {
		if p.Index == fmt.Sprintf("%v:%v", specialPayout.Payout.Symbol, specialPayout.Payout.Count) {
			winId = specialPayout.Index
		}
	}
	return winId
}

func (gamestate *Gamestate) PrepareTransactions(previousGamestate Gamestate) {
	// prepares transactions and sets round ID

	// check if WAGER TX exists
	if len(gamestate.Transactions) == 0 || gamestate.Action == "respin" || strings.Contains(gamestate.Action, "gamble") {
		if previousGamestate.RoundID == "" {
			// hack for games begun before this fix was implemented
			previousGamestate.RoundID = previousGamestate.Id
		}
		// this is a continued game round
		gamestate.RoundID = previousGamestate.RoundID
		// require a payout
	} else {
		gamestate.RoundID = gamestate.Id
	}

	relativePayout := NewFixedFromInt(gamestate.RelativePayout * gamestate.Multiplier)
	var gamestateWin Money
	if relativePayout != 0 || gamestate.RoundID != gamestate.Id {
		txID := gamestate.Id
		if gamestate.RoundID == gamestate.Id {
			txID = uuid.NewV4().String()
		}
		// add win transaction
		gamestateWin = Money{Amount: relativePayout.Mul(gamestate.BetPerLine.Amount), Currency: gamestate.BetPerLine.Currency} // this is in fixed notation i.e. 1.00 == 1000000
		gamestate.Transactions = append(gamestate.Transactions, WalletTransaction{Id: txID, Amount: gamestateWin, Type: "PAYOUT"})
	}

	if gamestate.Action != "base" {
		gamestate.PlaySequence = previousGamestate.PlaySequence + 1
		gamestate.CumulativeWin = previousGamestate.CumulativeWin + gamestateWin.Amount
		if gamestate.Action == "cascade" {
			gamestate.SpinWin = previousGamestate.SpinWin + gamestateWin.Amount
		} else {
			gamestate.SpinWin = gamestateWin.Amount
		}
	} else {
		gamestate.PlaySequence = 0
		gamestate.CumulativeWin = gamestateWin.Amount
		gamestate.SpinWin = gamestateWin.Amount
	}

}

func (gamestate *Gamestate) PrepareActions(previousActions []string) {
	logger.Debugf("Preparing actions, Previous: %v || New: %v", previousActions, gamestate.NextActions)
	newActions := gamestate.NextActions

	if len(newActions) == 0 {
		gamestate.NextActions = previousActions[1:]
		return
	}
	// handle some actions immediately
	switch newActions[0] {
	case "replaceQueuedActionType":
		// todo: add graceful failure when length of new actions is not 2 or more
		// i.e. for incrementing multipliers in freespins with additional triggers
		//replacementAction := newActions[1]
		//fmt.Printf("Replacement Type %v", replacementAction)
		for i := 1; i < len(previousActions)-1; i++ {
			previousActions[i] = newActions[1]
		}
		newActions = newActions[1:]

	case "replaceQueuedActions":
		// if a new feature trigger cancels queued actions
		previousActions = []string{"", "finish"}
		newActions = newActions[1:]

	case "queueActionsAfter":
		// instead of adding new actions to the front of the queue, add them to the end
		queuedActions := previousActions[1 : len(previousActions)-1]
		previousActions = append(newActions, "finish") // the first element of nextActions is queueActionsAfter and should be removed, but it will be on line 309
		newActions = queuedActions
	}
	gamestate.NextActions = append(newActions, previousActions[1:]...)
}

func (engine EngineDef) DetermineWins(symbolGrid [][]int) ([]Prize, int) {
	// calculate wins
	var wins []Prize
	switch engine.WinType {
	case "ways":
		wins = DetermineWaysWins(symbolGrid, engine.Payouts, engine.Wilds)
	case "lines":
		wins = DetermineLineWins(symbolGrid, engine.WinLines, engine.Payouts, engine.Wilds, engine.Compounding)
	case "barLines":
		wins = determineBarLineWins(symbolGrid, engine.WinLines, engine.Payouts, engine.Bars, engine.Wilds, engine.Compounding)
	case "blazeLines":
		// this is a special kind of line win defined for blaze games-- horizontal lines are mirrored vertically
		// and wins can exist anywhere within the line-- multiple payouts per line are possible
		// this will only work if grid width is same dimension as height
		if len(symbolGrid) != len(symbolGrid[0]) {
			logger.Errorf("Requesting vertical and horizontal line win calculation on non-standard grid size")
			return []Prize{}, 0
		}
		wins = DetermineLineWinsAnywhere(symbolGrid, engine.WinLines, engine.Payouts, engine.Wilds, engine.Compounding)
		// transpose grid
		sGTransposed := TransposeGrid(symbolGrid)
		vWins := DetermineLineWinsAnywhere(sGTransposed, engine.WinLines, engine.Payouts, engine.Wilds, engine.Compounding)
		for w:=0; w<len(vWins); w++{
			// add prefix to index and adjust line number
			// get base ref which is i reel first symbol
			base := vWins[w].SymbolPositions[0] / (len(symbolGrid[0]))
			vWins[w].Winline += len(engine.WinLines)
			vWins[w].Index = fmt.Sprintf("V%v", vWins[w].Index)
			var convSymbolPos []int
			for i:=0; i<len(vWins[w].SymbolPositions); i++ {
				convSymbolPos = append(convSymbolPos, (vWins[w].SymbolPositions[i]-(engine.ViewSize[0]*(i+base)))*engine.ViewSize[0] + i+base)
			}
			vWins[w].SymbolPositions = convSymbolPos
		}
		wins = append(wins, vWins...)
	case "pAndF":
		wins = determinePrimeAndFlopWins(symbolGrid, engine.Payouts, engine.Wilds)
	}
	relativePayout := calculatePayoutWins(wins)
	return wins, relativePayout
}

func (engine EngineDef) addStickyWilds(previousGamestate Gamestate, symbolGrid [][]int) [][]int {
	// this will fail if previous symbolGrid is of different dimensions than current
	if len(symbolGrid) != len(previousGamestate.SymbolGrid) {
		return symbolGrid
	}

	for w:=0; w<len(engine.Wilds); w++ {
		if engine.Wilds[w].Sticky {
			for i:=0; i<len(previousGamestate.SymbolGrid); i++ {
				for j:=0; j<len(previousGamestate.SymbolGrid[i]); j++ {
					if previousGamestate.SymbolGrid[i][j] == engine.Wilds[w].Symbol {
						if previousGamestate.Action == "base" {
							logger.Infof("Adding sticky wild %v from previous round", engine.Wilds[w].Symbol)

						}
						symbolGrid[i][j] = engine.Wilds[w].Symbol
					}
				}
			}


		}
	}
	return symbolGrid
}

func (engine EngineDef) BaseRound(parameters GameParams) Gamestate {
	// the base gameplay round
	// uses round multiplier if included
	// no dynamic reel calculation

	wl := engine.ProcessWinLines(parameters.SelectedWinLines)

	// spin
	symbolGrid, stopList := engine.Spin()

	// replace any symbols with sticky wilds
	symbolGrid = engine.addStickyWilds(parameters.previousGamestate, symbolGrid)

	wins, relativePayout := engine.DetermineWins(symbolGrid)
	// calculate specialWin
	var nextActions []string
	specialWin := DetermineSpecialWins(symbolGrid, engine.SpecialPayouts)
	if specialWin.Index != "" {
		var specialPayout int
		specialPayout, nextActions = engine.CalculatePayoutSpecialWin(specialWin)
		relativePayout += specialPayout
		wins = append(wins, specialWin)
	}
	logger.Debugf("got %v wins: %v", len(wins), wins)
	// get Multiplier
	multiplier := 1
	if len(engine.Multiplier.Multipliers) > 0 {
		multiplier = SelectFromWeightedOptions(engine.Multiplier.Multipliers, engine.Multiplier.Probabilities)
	}
	// Build gamestate
	gamestate := Gamestate{DefID: engine.Index, Prizes: wins, SymbolGrid: symbolGrid, RelativePayout: relativePayout, Multiplier: multiplier, StopList: stopList, NextActions: nextActions, SelectedWinLines: wl}

	return gamestate
}

func (engine EngineDef) MultiplierXWilds(parameters GameParams) Gamestate {
	// wild multipliers multiply all winnings

	gamestate := engine.BaseRound(parameters)
	ctWilds := 0
	for w:= 0; w<len(engine.Wilds); w++ {

		for r:= 0; r<len(gamestate.SymbolGrid); r++ {
			//iterate through reels
			for x:=0;x<len(gamestate.SymbolGrid[r]);x++ {
				if engine.Wilds[w].Symbol == gamestate.SymbolGrid[r][x] {
					ctWilds ++
				}
			}
		}
	}
	gamestate.Multiplier = 1

	for w:=0;w<ctWilds;w++ {
		gamestate.Multiplier *= SelectFromWeightedOptions(engine.Multiplier.Multipliers, engine.Multiplier.Probabilities)
	}

	return gamestate
}
// Guaranteed win round
func (engine EngineDef) GuaranteedWin(parameters GameParams) Gamestate {
	var gamestate Gamestate
	for len(gamestate.Prizes) == 0 {
		gamestate = engine.BaseRound(parameters)
		if config.GlobalConfig.DevMode == true && len(engine.force) > 0 {
			logger.Warnf("removing force to stop infinite iterations")
			engine.SetForce([]int{})
		}
	}
	return gamestate
}

// Cascading round
func (engine EngineDef) Cascade(parameters GameParams) Gamestate {
	symbolGrid := make([][]int, len(engine.ViewSize))
	stopList := make([]int, len(engine.ViewSize))
	var cascadePositions []int
	if parameters.Action == "cascade" {
		previousGamestate := parameters.previousGamestate
		// if previous gamestate contains a win, we need to cascade new tiles into the old space
		previousGrid := previousGamestate.SymbolGrid
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
		// return grid to full size by filling in empty spaces
		for i := 0; i < len(engine.ViewSize); i++ {
			numToAdd := engine.ViewSize[i] - len(remainingGrid[i])
			cascadePositions = append(cascadePositions, numToAdd)
			stop := previousGamestate.StopList[i] - numToAdd
			// get adjusted index if the previous win was at the top of the reel
			if stop < 0 {
				stop = len(engine.Reels[i]) + stop
			}
			symbolGrid[i] = append(adjustedReels[i][stop:stop+numToAdd], remainingGrid[i]...)
			stopList[i] = stop
		}

	} else {
		symbolGrid, stopList = engine.Spin()
	}

	// calculate wins
	wins, relativePayout := engine.DetermineWins(symbolGrid)
	// calculate specialWin
	var nextActions []string
	cascade := false
	// if any win is present, next action should be cascade
	if len(wins) > 0 {
		logger.Debugf("cascade is true")
		cascade = true
	} else {
		// only check for special win after cascading has completed
		logger.Debugf("determining special wins")
		specialWin := DetermineSpecialWins(symbolGrid, engine.SpecialPayouts)
		if specialWin.Index != "" {
			var specialPayout int
			specialPayout, nextActions = engine.CalculatePayoutSpecialWin(specialWin)
			relativePayout += specialPayout
			wins = append(wins, specialWin)
		}
	}
	if cascade {
		nextActions = append([]string{"cascade"}, nextActions...)
	}

	// get first Multiplier
	multiplier := 1
	if len(engine.Multiplier.Multipliers) > 0 {
		//multiplier = SelectFromWeightedOptions(engine.Multiplier.Multipliers, engine.Multiplier.Probabilities)
		multiplier = engine.Multiplier.Multipliers[0]
	}
	// for now, do a bit of a hack to get the cascade positions. as soon as we need to implement cascade with variable
	// win lines, this will need to be adjusted to be added into a new field in the gamestate message

	winlines := parameters.SelectedWinLines
	if len(winlines) == 0 {
		winlines = cascadePositions
	}
	// Build gamestate
	gamestate := Gamestate{DefID: engine.Index, Prizes: wins, SymbolGrid: symbolGrid, RelativePayout: relativePayout, Multiplier: multiplier, StopList: stopList, NextActions: nextActions, SelectedWinLines: winlines}
	return gamestate
}

func (engine EngineDef) CascadeMultiply(parameters GameParams) Gamestate {
	// multiplier increments for each cascade, up to the highest multiplier in the engine
	gamestate := engine.Cascade(parameters)
	// get next multiplier
	if parameters.Action == "cascade" {
		prevIndex := getIndex(parameters.previousGamestate.Multiplier, engine.Multiplier.Multipliers)
		//logger.Infof("index %v multiplier %v")
		if prevIndex < 0 {
			// fallback to first multiplier
			gamestate.Multiplier = engine.Multiplier.Multipliers[0]
		} else {

			gamestate.Multiplier = engine.Multiplier.Multipliers[minInt(prevIndex+1, len(engine.Multiplier.Multipliers)-1)]
		}
	} else {
		gamestate.Multiplier = engine.Multiplier.Multipliers[0]
	}

	return gamestate
}

func getIndex(a int, s []int) int {
	// gets the first index of appearance of value a in slice s
	for i := 0; i < len(s); i++ {
		if a == s[i] {
			return i
		}
	}
	return -1
}

func minInt(a int, b int) int {
	// returns the minimum of two integers
	if a > b {
		return b
	}
	return a
}

// Prize selection round
func (engine EngineDef) SelectPrize(parameters GameParams) Gamestate {
	// prize selection

	var specialWin Prize
	for _, prize := range engine.SpecialPayouts {
		if prize.Index == parameters.Selection {
			specialWin = prize
			break
		}
	}
	if specialWin.Index == "" {
		logger.Errorf("selection %v does not exist in prizes for engine %v", parameters.Selection, engine.ID)
		return Gamestate{}
	}
	relativePayout, nextActions := engine.CalculatePayoutSpecialWin(specialWin)
	// get Multiplier
	multiplier := 1
	if len(engine.Multiplier.Multipliers) > 0 {
		multiplier = SelectFromWeightedOptions(engine.Multiplier.Multipliers, engine.Multiplier.Probabilities)
	}
	// Build gamestate

	gamestate := Gamestate{DefID: engine.Index, Prizes: []Prize{specialWin}, NextActions: nextActions, RelativePayout: relativePayout, Multiplier: multiplier, SelectedWinLines: parameters.previousGamestate.SelectedWinLines} // most often, relativePayout will be zero
	return gamestate
}

// Respin Round
func (engine EngineDef) Respin(parameters GameParams) Gamestate {
	// this functionality can be enabled on any engine, but a boolean must be added in the engineconfig to allow it
	// the boolean must be set on each individual enginendef explicitly, it is not inherited as other properties
	if !engine.RespinAllowed {
		logger.Errorf("No respin allowed ont his engine")
		return Gamestate{}
	}
	// spin only one reel
	respinIndex := parameters.RespinReel
	previousGamestate := parameters.previousGamestate

	newSymbols, newStopValue := EngineDef{Reels: [][]int{engine.Reels[respinIndex]}, ViewSize: []int{engine.ViewSize[respinIndex]}}.Spin()
	symbolGrid := previousGamestate.SymbolGrid
	symbolGrid[respinIndex] = newSymbols[0]
	stopList := previousGamestate.StopList
	stopList[respinIndex] = newStopValue[0]

	wins, relativePayout := engine.DetermineWins(symbolGrid)
	var nextActions []string

	// Get scatter wins
	specialWin := DetermineSpecialWins(symbolGrid, engine.SpecialPayouts)
	if specialWin.Index != "" {
		var specialPayout int
		specialPayout, nextActions = engine.CalculatePayoutSpecialWin(specialWin)
		relativePayout += specialPayout
		wins = append(wins, specialWin)
	}

	// Build gamestate
	gamestate := Gamestate{DefID: engine.Index, Prizes: wins, SymbolGrid: symbolGrid, RelativePayout: relativePayout, Multiplier: 1, StopList: stopList, NextActions: nextActions}
	return gamestate
}

func (engine EngineDef) ShuffleFlop(parameters GameParams) Gamestate {

	return engine.ShuffleBase(parameters, "flop")
}

func (engine EngineDef) ShufflePrime(parameters GameParams) Gamestate {
	return engine.ShuffleBase(parameters, "prime")
}

func (engine EngineDef) Shuffle(parameters GameParams) Gamestate {
	return engine.ShuffleBase(parameters, "")
}

// Shuffle is similar to respin but no wager is charged
func (engine EngineDef) ShuffleBase(parameters GameParams, shuffleID string) Gamestate {
	// shuffle action should include information about which reels to shuffle
	var shuffleReels []int
	switch shuffleID {
	case "prime":
		shuffleReels = append(shuffleReels, 0)
	case "flop":
		for i := 1; i < len(engine.ViewSize); i++ {
			shuffleReels = append(shuffleReels, i)
		}
	default:
		for i := 0; i < len(engine.ViewSize); i++ {
			shuffleReels = append(shuffleReels, i)
		}
	}

	previousGamestate := parameters.previousGamestate
	symbolGrid := previousGamestate.SymbolGrid
	stopList := previousGamestate.StopList
	logger.Debugf("previous reels : %v", symbolGrid)

	for i := 0; i < len(shuffleReels); i++ {
		newSymbols, newStopValue := EngineDef{Reels: [][]int{engine.Reels[shuffleReels[i]]}, ViewSize: []int{engine.ViewSize[shuffleReels[i]]}}.Spin()
		symbolGrid[shuffleReels[i]] = newSymbols[0]
		stopList[shuffleReels[i]] = newStopValue[0]
	}
	logger.Debugf("new reels : %v", symbolGrid)

	wins, relativePayout := engine.DetermineWins(symbolGrid)

	var nextActions []string

	// Get scatter wins
	specialWin := DetermineSpecialWins(symbolGrid, engine.SpecialPayouts)
	if specialWin.Index != "" {
		var addlPayout int
		addlPayout, nextActions = engine.CalculatePayoutSpecialWin(specialWin)
		relativePayout += addlPayout
		wins = append(wins, specialWin)
	}

	// Build gamestate
	gamestate := Gamestate{DefID: engine.Index, Prizes: wins, SymbolGrid: symbolGrid, RelativePayout: relativePayout, Multiplier: 1, StopList: stopList, NextActions: nextActions}

	return gamestate
}

func calculatePayoutWins(wins []Prize) int {
	var relativePayout int
	for _, win := range wins {
		// add win amount
		relativePayout += win.Payout.Multiplier * win.Multiplier
	}
	return relativePayout
}

func (engine EngineDef) CalculatePayoutSpecialWin(specialWin Prize) (int, []string) {
	var nextActions []string
	var relativePayout int
	if specialWin.Index != "" {
		nextActionInfo := strings.Split(specialWin.Index, ":")
		count, _ := strconv.Atoi(nextActionInfo[1])
		for i := 0; i < count; i++ {
			nextActions = append(nextActions, nextActionInfo[0])
		}
		// special payout is not paid per line, it's a total stake multiplier
		relativePayout = specialWin.Payout.Multiplier * engine.StakeDivisor
	}
	logger.Debugf("Calculated payout for special win %v: %v", specialWin.Index, relativePayout)

	return relativePayout, nextActions
}

func (engine EngineDef) MaxWildRound(parameters GameParams) Gamestate {
	// this function takes the highest-level wild of the engine present on the view for that spin and replaces all other in-view wilds with that value

	// process parameters
	engine.ProcessWinLines(parameters.SelectedWinLines)

	// spin
	symbolGrid, stopList := engine.Spin()

	// replace wilds with highest-level wild
	type visibleWild struct {
		reelIndex    int
		reelPosition int
		symbol       int
	}
	var wildsPresent []visibleWild
	var highestWild int
	for i := 0; i < len(symbolGrid); i++ {
		for j := 0; j < len(symbolGrid[i]); j++ {
			for k := 0; k < len(engine.Wilds); k++ {
				symbol := engine.Wilds[k].Symbol
				if symbolGrid[i][j] == symbol {
					if symbol > highestWild {
						// new highest wild encountered
						highestWild = symbol
					}
					// add a new wild to wildsPresent
					wildsPresent = append(wildsPresent, visibleWild{
						reelIndex:    i,
						reelPosition: j,
						symbol:       symbol,
					})
				}
			}
		}
	}

	// overwrite all wilds that are not of the highest symbol with the highest symbol
	for i := 0; i < len(wildsPresent); i++ {
		symbolGrid[wildsPresent[i].reelIndex][wildsPresent[i].reelPosition] = highestWild
	}
	wins, relativePayout := engine.DetermineWins(symbolGrid)

	// calculate specialWin
	var nextActions []string
	specialWin := DetermineSpecialWins(symbolGrid, engine.SpecialPayouts)
	if specialWin.Index != "" {
		var specialPayout int
		specialPayout, nextActions = engine.CalculatePayoutSpecialWin(specialWin)
		relativePayout += specialPayout
		wins = append(wins, specialWin)
	}

	// Build gamestate
	gamestate := Gamestate{DefID: engine.Index, Prizes: wins, SymbolGrid: symbolGrid, RelativePayout: relativePayout, Multiplier: 1, StopList: stopList, NextActions: nextActions}
	return gamestate
}

func (engine EngineDef) DynamicWildWaysRound(parameters GameParams) Gamestate {
	// feature is not retriggerable, empty slice of next actions always returns
	// fs reels built randomly on each spin

	// build wild reels
	// choose the number of wilds to appear
	numWilds := GetWeightedIndex([]int{32, 48, 10, 10})
	logger.Debugf("Number of Wilds selected: %v", numWilds)
	// choose the locations of the wilds

	// possible wild locations
	potentialWildLocations := [][][]int{{{0, 0, 0, 0, 0}}, // for 0 wilds
		{{1, 0, 0, 0, 0}, {0, 1, 0, 0, 0}, {0, 0, 1, 0, 0}, {0, 0, 0, 1, 0}, {0, 0, 0, 0, 1}}, // for 1 wild
		{{2, 0, 0, 0, 0}, {1, 1, 0, 0, 0}, {1, 0, 1, 0, 0}, {1, 0, 0, 1, 0}, {1, 0, 0, 0, 1}, {0, 2, 0, 0, 0}, {0, 1, 1, 0, 0}, {0, 1, 0, 1, 0}, {0, 1, 0, 0, 1}, {0, 0, 2, 0, 0}, {0, 0, 1, 1, 0}, {0, 0, 1, 0, 1}, {0, 0, 0, 2, 0}, {0, 0, 0, 1, 1}, {0, 0, 0, 0, 2}}, // for 2 wilds
		{{3, 0, 0, 0, 0}, {0, 3, 0, 0, 0}, {0, 0, 3, 0, 0}, {0, 0, 0, 3, 0}, {0, 0, 0, 0, 3},
			{2, 1, 0, 0, 0}, {2, 0, 1, 0, 0}, {2, 0, 0, 1, 0}, {2, 0, 0, 0, 1},
			{1, 2, 0, 0, 0}, {0, 2, 1, 0, 0}, {0, 2, 0, 1, 0}, {0, 2, 0, 0, 1},
			{1, 0, 2, 0, 0}, {0, 1, 2, 0, 0}, {0, 0, 2, 1, 0}, {0, 0, 2, 0, 1},
			{1, 0, 0, 2, 0}, {0, 1, 0, 2, 0}, {0, 0, 1, 2, 0}, {0, 0, 0, 2, 1},
			{1, 0, 0, 0, 2}, {0, 1, 0, 0, 2}, {0, 0, 1, 0, 2}, {0, 0, 0, 1, 2},
			{1, 1, 1, 0, 0}, {1, 1, 0, 1, 0}, {1, 1, 0, 0, 1}, {1, 0, 1, 1, 0}, {1, 0, 1, 0, 1}, {1, 0, 0, 1, 1},
			{0, 1, 1, 1, 0}, {0, 1, 1, 0, 1}, {0, 1, 0, 1, 1}, {0, 0, 1, 1, 1},
		}, // for 3 wilds
	}

	wildLocations := potentialWildLocations[numWilds][rng.RandFromRange(len(potentialWildLocations[numWilds]))]

	logger.Debugf("Wild locations: %v", wildLocations)

	//choose fs reels based on numWilds
	fsReels := make([][]int, len(engine.ViewSize))
	for i := 0; i < len(fsReels); i++ {
		fsReels[i] = engine.Reels[wildLocations[i]]
	}
	logger.Debugf("FS reels: %v", fsReels)
	engine.Reels = fsReels

	//choose fs multiplier based on numWilds
	// todo define in constants in config file
	fsMultiplierP := [][]int{{7, 6, 2}, {11, 2, 2}, {12, 2, 1}, {13, 1, 1}}[numWilds] // each slice is weight of multiplier {1,2,3}
	freespinMultiplier := SelectFromWeightedOptions([]int{1, 2, 3}, fsMultiplierP)
	logger.Debugf("P: %v; Selected Multiplier: %v", fsMultiplierP, freespinMultiplier)

	symbolGrid, stopList := engine.Spin()
	wins := DetermineWaysWins(symbolGrid, engine.Payouts, engine.Wilds)
	logger.Debugf("symbolgrid: %v; wins: %v", symbolGrid, wins)
	relativePayout := calculatePayoutWins(wins)
	gamestate := Gamestate{DefID: engine.Index, Prizes: wins, SymbolGrid: symbolGrid, RelativePayout: relativePayout, Multiplier: freespinMultiplier, StopList: stopList, NextActions: []string{}}
	return gamestate
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (engine *EngineDef) ProcessWinLines(selectedWinLines []int) (wl []int) {
	logger.Debugf("processing %v winlines: %v", selectedWinLines, engine.WinLines)
	if len(selectedWinLines) == 0 || len(engine.WinLines) == 0 {
		for i := 0; i < len(engine.WinLines); i++ {
			wl = append(wl, i)
		}
		if engine.StakeDivisor == 0 {
			engine.StakeDivisor = len(wl)
		}
		return
	}
	winLines := make([][]int, min(len(selectedWinLines), len(engine.WinLines)))

	for i, line := range selectedWinLines {
		if line < len(engine.WinLines) {
			winLines[i] = engine.WinLines[line]
			wl = append(wl, line)
		}
	}
	engine.WinLines = winLines
	engine.StakeDivisor = len(wl)
	return
}

func (engine EngineDef) LinesRoundReplaceType(parameters GameParams) Gamestate {
	// debatable whether we would ever want not to do this
	// for most games it doesn't make a difference as there's only one feature type
	// for games where feature progresses unidirectionally this must be called in explicitly
	// potentially make this the default behavior?
	gamestate := engine.BaseRound(parameters)
	if len(gamestate.NextActions) > 0 {
		gamestate.NextActions = append([]string{"replaceQueuedActionType"}, gamestate.NextActions...)

	}
	return gamestate
}


func (engine EngineDef) TwoStageExpand(parameters GameParams) Gamestate {
	// this is a base round but with a special expanding symbol (indicated by the value after ":" in Action field of previous gamestate
	// the expansion happens in the second phase, so this function serves only to prepare the next phase
	// if it is the first in a series of gamestates of this type, the special symbol needs to be inferred from the multiplier of the previous gs, which should have payout zero

	gamestate := engine.BaseRound(parameters)

	var expandSymbol int
	var err error
	if strings.Contains(parameters.previousGamestate.Action, parameters.Action) {
		// expand symbol should be propogated from previous state
		expandSymbol, err = strconv.Atoi(strings.Split(parameters.previousGamestate.Action, "E")[1])
		if err != nil {
			logger.Errorf("misconfigured engine: %v", engine.ID)
			return Gamestate{}
		}
	} else {
		// this is the first expansion state, get expand symbol from the multiplier of the previous state
		expandSymbol = parameters.previousGamestate.Multiplier
	}

	// check if the reels contain the expand symbol
	expandReels := make([]int, len(gamestate.SymbolGrid)) // this will be a grid of zeroes and ones, we will take the sum to get the prize
	ctExpandReels := 0
	for i:=0; i<len(gamestate.SymbolGrid); i++ {
		for j:=0; j<len(gamestate.SymbolGrid[i]); j++ {
			if gamestate.SymbolGrid[i][j] == expandSymbol {
				expandReels[i] = 1
				ctExpandReels ++
				// this should only break innermost for loop and continue to next reel
				break
			}
		}
	}
	// search for prize matching the count and symbol
	for p:=0; p<len(engine.Payouts); p++ {
		if engine.Payouts[p].Symbol == expandSymbol && engine.Payouts[p].Count == ctExpandReels {
			// in theory we could make expand a separate gs but let's do it all together to reduce api calls
			// this is hardcoded to work for engine XVII, where the payout after expansion is different than base payout
			// rather than following strict lines, the payout can happen in nonadjacent positions

			// the client will know which symbol to display as the expand symbol, so we don't need to calculate a new grid.
			// we simply need to know which reels contain the symbol that has expanded, and we will do a special win calculation on that

			// create a win matching the payout type
			winPos := engine.WinLines[0]
			//	// turn every reel with no matching symbol to -1
			for i:=0; i<len(expandReels); i++ {
				if expandReels[i] == 0 {
					winPos[i] = -1
				} else {
					winPos[i] = 1
				}
			}
			gamestate.Prizes = append(gamestate.Prizes, Prize{Payout: engine.Payouts[p], Index: fmt.Sprintf("%v:%v", expandSymbol, ctExpandReels), Multiplier: len(engine.WinLines), SymbolPositions: winPos, Winline: -1})
			// this should only match once per round
			break
		}
	}
	relativePayout := calculatePayoutWins(gamestate.Prizes)
	gamestate.RelativePayout = relativePayout // override
	gamestate.Action = fmt.Sprintf("%vE%v",parameters.Action, expandSymbol)
	return gamestate
}

