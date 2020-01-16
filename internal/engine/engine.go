package engine

// Spin and play functions of engines

import (
	"errors"
	"fmt"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/config"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"reflect"
	"strconv"
	"strings"
)

func Spin(reels [][]int, viewSize []int) ([][]int, []int) {
	// Performs a random spin of the reels
	if len(viewSize) != len(reels) {
		logger.Fatalf("ViewSize (%v) and Reel size (%v) mismatch", len(viewSize), len(reels))
		return [][]int{}, []int{}
	}
	stopList := make([]int, len(viewSize))

	for index, reel := range reels {
		// choose a random index on the reel
		reelIndex := rng.RandFromRange(len(reel))
		stopList[index] = reelIndex
	}
	symbolGrid := GetSymbolGridFromStopList(reels, viewSize, stopList)
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

// DetermineLineWins ...
func DetermineLineWins(symbolGrid [][]int, WinLines [][]int, linePayouts []Payout, wilds []wild) []Prize {
	// determines prizes from line wins including wilds with multipliers
	// highest wild multiplier takes precedence for multiple wilds on the same line (i.e. wild multipliers do not compound)

	var lineWins []Prize

	for winLineIndex, winLine := range WinLines {
		lineContent := make([]int, len(symbolGrid))
		symbolPositions := make([]int, len(symbolGrid))
		for reel, index := range winLine {
			lineContent[reel] = symbolGrid[reel][index]
			// to determine a unique symbol position per location in the view, this only works if view is regularly sized (i.e. all reels show same number of symbols. todo: make this dynamic for all reel configurations
			symbolPositions[reel] = len(symbolGrid[reel])*reel + index
		}
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
					engineWildMultiplier := SelectFromWeightedOptions(engineWild.Multiplier.Multipliers, engineWild.Multiplier.Probabilities)
					if lineSymbol == engineWild.Symbol {
						// if the lineSymbol is a wild then all symbols so far have been wild, and this one is not, so update lineSymbol
						// as is, if all 5 symbols are wild then no multiplier will be added. this should be defined in the payout
						lineSymbol = symbol
						// multipliers do not compound, there may be only one line multiplier
						if engineWildMultiplier > multiplier {
							multiplier = engineWildMultiplier
						}
						numMatch++
						// stop checking for wilds if one has been matched
						match = true
						break
					} else if symbol == engineWild.Symbol {
						// if the new symbol on the line is a wild count it as a match
						numMatch++
						// NB: highest multiplier replaces others, multipliers do not compound. uncomment next line to change that. potentially make this an option in inputs (	to compound: //multiplier = multiplier*engineWildMultiplier)
						if engineWildMultiplier > multiplier {
							multiplier = engineWildMultiplier
						}
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
				lineWins = append(lineWins, Prize{Payout: linePayout, Index: fmt.Sprintf("%v:%v", lineSymbol, numMatch), Multiplier: multiplier, SymbolPositions: symbolPositions[:numMatch], Winline: winLineIndex})
				// Only one win possible per line, earliest payout in payout dict takes precedence
				break
			}
		}
	}
	return lineWins
}

func determineBarLineWins(symbolGrid [][]int, winLines [][]int, payouts []Payout, bars []bar, wilds []wild) []Prize {
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
	lineWinsWithBar := DetermineLineWins(adjustedSymbolGrid, winLines, payouts, wilds)
	lineWinsWithoutBar := DetermineLineWins(symbolGrid, winLines, payouts, wilds)
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
				waysWins = append(waysWins, Prize{Payout: prizePayout, Index: winIndex, Multiplier: variation.multiplier, SymbolPositions: variation.symbolPositions, Winline: -1})
				matchedSymbols = append(matchedSymbols, waysPayout.Symbol)
			}
		}

	}
	return waysWins
}

func TransposeGrid(symbolGrid [][]int) [][]int {
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
			specialPayout.Winline = -1
			return specialPayout
		}
	}
	return Prize{}
}

func determinePrimeAndFlopWins(symbolGrid [][]int, payouts []Payout, wilds []wild) []Prize {
	// by default, prime is last reel
	// check if symbol grid is more than one row, if so throw error
	// win only if any symbols on flop match the prime symbol
	if len(symbolGrid[0]) > 1 {
		panic(errors.New("too many rows for this engine type"))
	}
	symbols := symbolGrid[0]
	prime := symbols[len(symbols)-1]
	numMatch := 0

	for _, symbol := range symbols {
		if symbol == prime {
			numMatch++
		}
	}
	for _, payout := range payouts {
		if prime == payout.Symbol && numMatch == payout.Count {
			// copy payout NB can only do because no reference fields in Payout
			pfPayout := payout
			return []Prize{Prize{Payout: pfPayout, Index: strconv.Itoa(int(prime)) + "flop" + strconv.Itoa(numMatch), Multiplier: 1, Winline: -1}}
			// Only one win possible per line, earliest payout in payout dict takes precedence
		}
	}
	return []Prize{Prize{}}
}

// Play ...
func Play(previousGamestate Gamestate, betPerLine Fixed, currency string, parameters GameParams) (Gamestate, EngineConfig) {
	logger.Debugf("Playing round with parameters: %#v", parameters)
	gameID, _ := GetGameIDAndReelset(previousGamestate.GameID)
	engineID, err := config.GetEngineFromGame(gameID)
	engineConf := BuildEngineDefs(engineID)
	totalBet := Money{Amount: betPerLine.Mul(NewFixedFromInt(engineConf.EngineDefs[0].StakeDivisor)), Currency: currency}
	var transactions []WalletTransaction
	var actions []string
	//relativePayout := StrToDec("0.000")
	if len(previousGamestate.NextActions) == 1 && previousGamestate.NextActions[0] == "finish" {
		// if this is a respin, special case:
		if parameters.Action == "reSpin" {
			// if action is respin, WAGER is dependent on reel configuration
			// index of the reel to be respun must be passed
			if parameters.RespinReel < 0 { //|| parameters[0].Type() != Fixed {
				logger.Errorf("ERROR, NO RESPIN REEL INDEX PASSED")
			}
			actions = append([]string{previousGamestate.Action}, previousGamestate.NextActions...)
			betPerLine = previousGamestate.BetPerLine.Amount
			reelIndex := parameters.RespinReel
			reelCost := engineConf.EngineDefs[engineConf.getDefIdByName(previousGamestate.Action)].GetRespinPriceReel(reelIndex, engineConf, previousGamestate)
			transactions = append(transactions, WalletTransaction{Id: previousGamestate.NextGamestate, Type: "WAGER", Amount: Money{reelCost, currency}})
			parameters.previousGamestate = previousGamestate
		} else {
			// new gameplay round
			transactions = append(transactions, WalletTransaction{Id: previousGamestate.NextGamestate, Type: "WAGER", Amount: totalBet})
			// if this is engine X or any other offering maxBase, check if max winlines are selected
			if parameters.Action == "maxBase" {
				maxDef := engineConf.getDefIdByName("maxBase")
				if maxDef < 0 || len(parameters.SelectedWinLines) != len(engineConf.EngineDefs[maxDef].WinLines) {
					// if maxBase has been passed from api, then no action was submitted and the bet was the maximum possible\
					// if the selected winlines is also maximum, then maintain maxBase, otherwise revert to base action
					parameters.Action = "base"
				}
			}
			actions = []string{parameters.Action, "finish"}
		}
	} else {
		logger.Debugf("Continuing game round, no WAGER charged")
		// every gamestate should have wager
		transactions = append(transactions, WalletTransaction{Id: previousGamestate.NextGamestate, Type: "WAGER", Amount: Money{0, currency}})

		actions = previousGamestate.NextActions
		// todo: do we want to insist that betperline be the same as previous gamestate?
		if actions[0] != "base" {
			betPerLine = previousGamestate.BetPerLine.Amount
		}
		// todo: be smarter in passing on parameters, but for now just add selectedWinLines if it existed previously and this request has no parameters
		if len(previousGamestate.SelectedWinLines) > 0 && len(parameters.SelectedWinLines) == 0 {
			// assume this is a line game and info should be propogated
			parameters.SelectedWinLines = previousGamestate.SelectedWinLines
		}
	}

	method, err := engineConf.getEngineAndMethod(actions[0])

	if err != nil {
		panic(err)
	}

	gamestateAndNextActions := method.Call([]reflect.Value{reflect.ValueOf(parameters)})

	gamestate, ok := gamestateAndNextActions[0].Interface().(Gamestate)
	if !ok {
		panic("value not a gamestate")
	}
	gamestate.Action = actions[0]
	gamestate.BetPerLine = Money{betPerLine, currency}
	gamestate.Transactions = transactions
	gamestate.Gamification = previousGamestate.Gamification
	gamestate.UpdateGamification(gameID)
	gamestate.PrepareActions(actions)
	logger.Debugf("Next actions after processing: %v", gamestate.NextActions)

	gamestate.GameID = gameID + gamestate.GameID // engineDef should be set in method
	gamestate.Id = previousGamestate.NextGamestate

	nextID := rng.RandStringRunes(8)
	gamestate.NextGamestate = nextID
	gamestate.PrepareTransactions()

	return gamestate, engineConf
}

func (gamestate *Gamestate) UpdateGamification(gameSlug string) {
	// update gamification status
	// this must happen before nextactions is handled
	switch gameSlug {
	case "a-fairy-tale", "a-candy-girls-christmas", "battlemech":
		if len(gamestate.NextActions) > 0 {
			gamestate.Gamification.Increment(3)
		}
	case "sky-jewels":
		gamestate.Gamification.IncrementSpins(randomRangeInt32(), 6)
	case "goal", "cookoff-champion":
		gamestate.Gamification.IncrementSpins(randomRangeInt32(), 3)
	}
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

func (gamestate *Gamestate) CalculateRelativePayout() {
	// calculates relativepayout from wins, if win index contains "freespin", assumption is that the payout is multiplied by the entire stake, and not only by the stake per line
	var relativePayout int

	engineConf, engineDef, err := GetEngineDefFromGame(gamestate.GameID)
	if err != nil {
		logger.Errorf("Error calulcating relative payout: %v", err)
		return
	}

	for i := 0; i < len(gamestate.Prizes); i++ {
		// add relative payout
		addlPayout := gamestate.Prizes[i].Payout.Multiplier * gamestate.Prizes[i].Multiplier * gamestate.Multiplier
		gamestate.Prizes[i].Index = engineConf.DetectSpecialWins(engineDef, gamestate.Prizes[i])
		if strings.Contains(gamestate.Prizes[i].Index, "freespin") {
			addlPayout = addlPayout * engineConf.EngineDefs[engineDef].StakeDivisor
		}
		relativePayout += addlPayout
	}
	gamestate.RelativePayout = relativePayout
}

func (gamestate *Gamestate) PrepareTransactions() {
	relativePayout := NewFixedFromInt(gamestate.RelativePayout * gamestate.Multiplier)
	if relativePayout != 0 {
		// add win transaction
		gamestateWin := Money{Amount: relativePayout.Mul(gamestate.BetPerLine.Amount), Currency: gamestate.BetPerLine.Currency} // this is in fixed notation i.e. 1.00 == 1000000
		gamestate.Transactions = append(gamestate.Transactions, WalletTransaction{Id: rng.RandStringRunes(8), Amount: gamestateWin, Type: "PAYOUT"})
	}

	// if finish is only remaining action, add endround transaction
	if len(gamestate.NextActions) == 1 && gamestate.NextActions[0] == "finish" {
		gamestate.Transactions = append(gamestate.Transactions, WalletTransaction{
			Id:     rng.RandStringRunes(8),
			Amount: Money{Amount: 0, Currency: gamestate.BetPerLine.Currency}, // todo do we want to set this to avoid unmarshalling of null values into something unsavory?
			Type:   "ENDROUND",
		})
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
		newActions = newActions[2:]

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
		wins = DetermineLineWins(symbolGrid, engine.WinLines, engine.Payouts, engine.Wilds)
	case "barLines":
		wins = determineBarLineWins(symbolGrid, engine.WinLines, engine.Payouts, engine.Bars, engine.Wilds)
	case "pAndF":
		wins = determinePrimeAndFlopWins(symbolGrid, engine.Payouts, engine.Wilds)
	}
	relativePayout := calculatePayoutWins(wins)
	return wins, relativePayout
}
func (engine EngineDef) BaseRound(parameters GameParams) Gamestate {
	// the base gameplay round
	// uses round multiplier if included
	// no dynamic reel calculation

	engine.ProcessWinLines(parameters.SelectedWinLines)

	// spin
	symbolGrid, stopList := Spin(engine.Reels, engine.ViewSize)

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
	// get Multiplier
	multiplier := 1
	if len(engine.Multiplier.Multipliers) > 0 {
		multiplier = SelectFromWeightedOptions(engine.Multiplier.Multipliers, engine.Multiplier.Probabilities)
	}
	// Build gamestate
	gamestate := Gamestate{GameID: fmt.Sprintf(":%v", engine.Index), Prizes: wins, SymbolGrid: symbolGrid, RelativePayout: relativePayout, Multiplier: multiplier, StopList: stopList, NextActions: nextActions, SelectedWinLines: parameters.SelectedWinLines}
	return gamestate
}

//func (engine EngineDef) processWins(symbolGrid ) ([]Prize, int){
//	// calculates wins by standard method and
//}

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

	// hack for engine III, this should be more flexible in future:
	gamestate := engine.BaseRound(parameters)
	gamestate.NextActions = append(nextActions[1:], gamestate.NextActions...)
	gamestate.RelativePayout += relativePayout
	// Build gamestate
	//gamestate := Gamestate{GameID: fmt.Sprintf(":%v", engine.Index), Prizes: []Prize{specialWin}, NextActions: nextActions, RelativePayout: relativePayout, Multiplier: 1} // most often, relativePayout will be zero
	return gamestate
}

// Respin Round
func (engine EngineDef) RespinRound(parameters GameParams) Gamestate {
	// spin only one reel
	respinIndex := parameters.RespinReel
	previousGamestate := parameters.previousGamestate
	newSymbols, newStopValue := Spin([][]int{engine.Reels[respinIndex]}, []int{engine.ViewSize[respinIndex]})

	symbolGrid := previousGamestate.SymbolGrid
	symbolGrid[respinIndex] = newSymbols[0]
	stopList := previousGamestate.StopList
	stopList[respinIndex] = newStopValue[0]

	wins, relativePayout := engine.DetermineWins(symbolGrid)

	var nextActions []string

	// Get scatter wins
	specialWin := DetermineSpecialWins(symbolGrid, engine.SpecialPayouts)

	addlPayout, nextActions := engine.CalculatePayoutSpecialWin(specialWin)
	relativePayout += addlPayout
	wins = append(wins, specialWin)

	// Build gamestate
	gamestate := Gamestate{GameID: fmt.Sprintf(":%v", engine.Index), Prizes: wins, SymbolGrid: symbolGrid, RelativePayout: relativePayout, Multiplier: 1, StopList: stopList, NextActions: nextActions}

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
	logger.Debugf("Calculagted payout for special win %v: %v", specialWin.Index, relativePayout)

	return relativePayout, nextActions
}

// Calculate price of reel respin
// value of a reel respin is equal to the expected payout given other reels stay as they are

func (engine EngineDef) GetRespinPrice(gamestate Gamestate, method string) []Fixed {
	// we want to hold RTP constant, so the price of a respin should be relative to the value of the respin and the RTP:
	// value of respin / price to respin = RTP
	// price to respin = value of respin / RTP

	// get engineConfig for RTP info
	engineConfig, _, err := GetEngineDefFromGame(gamestate.GameID)
	if err != nil {
		logger.Errorf("error in getting respin price: %v", err)
		return []Fixed{}
	}

	// prices are relative to betPerLine
	reelPrices := make([]Fixed, len(gamestate.SymbolGrid))

	for reelIndex := 0; reelIndex < len(gamestate.SymbolGrid); reelIndex++ {
		reelPrices[reelIndex] = engine.GetRespinPriceReel(reelIndex, engineConfig, gamestate)
	}
	return reelPrices
}

func (engine EngineDef) GetRespinPriceReel(reelIndex int, engineConfig EngineConfig, gamestate Gamestate) Fixed {
	// we want to hold RTP constant, so the price of a respin should be relative to the value of the respin and the RTP:
	// value of respin / price to respin = RTP
	// price to respin = value of respin / RTP
	return engine.GetExpectedReelValue(gamestate, reelIndex).Div(NewFixedFromFloat(engineConfig.RTP))

}

// calculate expected value of a reel, given the other reels do not move
func (engine EngineDef) GetExpectedReelValue(gamestate Gamestate, reelIndex int) Fixed {
	// NB ::: THIS DOES NOT IMNCLUDE GAMESTATE MULTIPLIERS i.e. RANDOM WHOLE ROUND MULTIPLIERS
	// gamestate must already have SymbolGrid calculated
	// get engine data
	view := gamestate.SymbolGrid
	reel := engine.Reels[reelIndex]
	reel = append(reel, reel[:engine.ViewSize[reelIndex]]...)

	var potentialWinValue Fixed

	// iterate through reel positions for the given reel to determine payouts
	for i := 0; i < len(reel); i++ {
		view[reelIndex] = reel[i : i+engine.ViewSize[reelIndex]]

		// calculate win
		var wins []Prize

		switch engine.WinType {
		case "ways":
			wins = DetermineWaysWins(view, engine.Payouts, engine.Wilds)
		case "lines":
			wins = DetermineLineWins(view, engine.WinLines, engine.Payouts, engine.Wilds)
		}
		for _, win := range wins {
			// add win amount (multipliers are relative to betPerLine)
			potentialWinValue += NewFixedFromInt(win.Payout.Multiplier).Mul(NewFixedFromInt(win.Multiplier))
		}
		specialWin := DetermineSpecialWins(view, engine.SpecialPayouts)

		if specialWin.Index != "" {
			// add prize value (multiplier is for total stake, divide by bet multiplier
			potentialWinValue += NewFixedFromInt(specialWin.Payout.Multiplier).Mul(NewFixedFromInt(specialWin.Multiplier)).Div(NewFixedFromInt(engine.StakeDivisor))

			// include estimated value of each round for the number of rounds that have been won
			// todo: analyze how this works for engines with multiple matching defs
			winInfo := strings.Split(specialWin.Index, ":")
			engineConfig, rsID, err := GetEngineDefFromGame(gamestate.GameID)
			if err != nil {
				logger.Errorf("Error getting Engine Def from game id")
				return Fixed(0)
			}

			specialEngineDef := engineConfig.EngineDefs[rsID]

			// expectedPayout is relative to the total stake, so divide by the bet multiplier
			singleRoundPayout := specialEngineDef.ExpectedPayout.Div(NewFixedFromInt(specialEngineDef.StakeDivisor))
			numRounds, err := strconv.Atoi(winInfo[1])
			if err != nil {
				logger.Errorf("Error in special win index: %v", specialWin.Index)
				return 0
			}
			for j := 0; j < numRounds; j++ {
				potentialWinValue += singleRoundPayout
			}

		}
	}

	// calculate average for all reel positions, divide total reel potential payout by number of reel positions
	return potentialWinValue.Div(NewFixedFromInt(len(reel)))
}

func (engine EngineDef) MaxWildRound(parameters GameParams) Gamestate {
	// this function takes the highest-level wild of the engine present on the view for that spin and replaces all other in-view wilds with that value

	// process parameters
	engine.ProcessWinLines(parameters.SelectedWinLines)

	// spin
	symbolGrid, stopList := Spin(engine.Reels, engine.ViewSize)

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
	gamestate := Gamestate{GameID: fmt.Sprintf(":%v", engine.Index), Prizes: wins, SymbolGrid: symbolGrid, RelativePayout: relativePayout, Multiplier: 1, StopList: stopList, NextActions: nextActions}
	return gamestate
}

// DynamicWildWaysRound ...
//func (engine EngineDef) ScrambleReelsRound(parameters GameParams) Gamestate {
//
//
//}
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
	//wildLocations := make([]int, len(engine.ViewSize))
	//for i := 0; i < numWilds; i++ {
	//	reelNum := rng.RandFromRange(5)
	//	wildLocations[reelNum]++
	//}
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

	symbolGrid, stopList := Spin(fsReels, engine.ViewSize)
	wins := DetermineWaysWins(symbolGrid, engine.Payouts, engine.Wilds)
	logger.Debugf("symbolgrid: %v; wins: %v", symbolGrid, wins)
	relativePayout := calculatePayoutWins(wins)
	gamestate := Gamestate{GameID: fmt.Sprintf(":%v", engine.Index), Prizes: wins, SymbolGrid: symbolGrid, RelativePayout: relativePayout, Multiplier: freespinMultiplier, StopList: stopList, NextActions: []string{}}
	return gamestate
}

func (engine *EngineDef) ProcessWinLines(selectedWinLines []int) {
	if len(selectedWinLines) == 0 {
		return
	}
	winLines := make([][]int, len(selectedWinLines))
	for i, line := range selectedWinLines {
		winLines[i] = engine.WinLines[line]
	}
	engine.WinLines = winLines
	engine.StakeDivisor = len(winLines)
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
