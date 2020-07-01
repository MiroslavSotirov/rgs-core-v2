package engine

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"strconv"
	"strings"
)

// Calculate price of reel respin
// value of a reel respin is equal to the expected payout given other reels stay as they are


func (gamestate Gamestate) RespinPrices(ccy string) (prices []Fixed, err rgse.RGSErr) {
	// we want to hold RTP constant, so the price of a respin should be relative to the value of the respin and the RTP:
	// value of respin / price to respin = RTP
	// price to respin = value of respin / RTP

	// prices are relative to betPerLine

	for reelIndex := 0; reelIndex < len(gamestate.SymbolGrid); reelIndex++ {
		prices = append(prices, RoundUpToNearestCCYUnit(Money{gamestate.RespinPriceReel(reelIndex), ccy}).Amount)
	}
	return
}

func (gamestate Gamestate) RespinPriceReel(reelIndex int) Fixed {
	// we want to hold RTP constant, so the price of a respin should be relative to the value of the respin and the RTP:
	// value of respin / price to respin = RTP
	// price to respin = value of respin / RTP
	EC, err := gamestate.Engine()
	if err != nil {
		return 0
	}
	return gamestate.ExpectedReelValue(reelIndex).Div(NewFixedFromFloat(EC.RTP))
}


// calculate expected value of a reel, given the other reels do not move
func (gamestate Gamestate) ExpectedReelValue(reelIndex int) Fixed {
	// NB ::: THIS DOES NOT IMNCLUDE GAMESTATE MULTIPLIERS i.e. RANDOM WHOLE ROUND MULTIPLIERS
	// gamestate must already have SymbolGrid calculated
	// get engine data
	logger.Debugf("Getting expected value of reel %v", reelIndex)
	def, err := gamestate.EngineDef()
	if err != nil {
		return 0
	}
	view := make([][]int, len(gamestate.SymbolGrid))
	copy(view, gamestate.SymbolGrid)
	reel := def.Reels[reelIndex]
	reel = append(reel, reel[:def.ViewSize[reelIndex]]...)

	var potentialWinValue Fixed

	// iterate through reel positions for the given reel to determine payouts
	for i := 0; i < len(reel)-def.ViewSize[reelIndex]; i++ {

		view[reelIndex] = reel[i : i+def.ViewSize[reelIndex]]
		logger.Debugf("Simulating result of view %v", view)
		// calculate win
		var wins []Prize

		switch def.WinType {
		case "ways":
			wins = DetermineWaysWins(view, def.Payouts, def.Wilds)
		case "lines":
			wins = DetermineLineWins(view, def.WinLines, def.Payouts, def.Wilds)
		}
		for _, win := range wins {
			// add win amount (multipliers are relative to betPerLine)
			potentialWinValue += NewFixedFromInt(win.Payout.Multiplier).Mul(NewFixedFromInt(win.Multiplier))
			logger.Debugf("added value %v for win %v for total of %v", NewFixedFromInt(win.Payout.Multiplier).Mul(NewFixedFromInt(win.Multiplier)), win.Index, potentialWinValue)
		}
		specialWin := DetermineSpecialWins(view, def.SpecialPayouts)

		if specialWin.Index != "" {
			// add prize value (multiplier is for total stake, multiply by bet multiplier
			// total prize = totalstake * multiplier = betperline * stakedivisor * multiplier
			potentialWinValue += NewFixedFromInt(specialWin.Payout.Multiplier).Mul(NewFixedFromInt(specialWin.Multiplier)).Mul(NewFixedFromInt(def.StakeDivisor))
			logger.Debugf("non-null special prize detected %v, win value %v", specialWin, NewFixedFromInt(specialWin.Payout.Multiplier).Mul(NewFixedFromInt(specialWin.Multiplier)).Mul(NewFixedFromInt(def.StakeDivisor)))
			// include estimated value of each round for the number of rounds that have been won
			// todo: analyze how to make this work for engines with multiple matching defs
			winInfo := strings.Split(specialWin.Index, ":")
			engineConfig, err := gamestate.Engine()
			if err != nil {
				logger.Errorf("Error getting Engine Def from game id")
				return Fixed(0)
			}
			numRounds, cerr := strconv.Atoi(winInfo[1])
			if numRounds < 1 {
				continue
			}
			specialDefID := engineConfig.DefIdByName(winInfo[0])
			specialEngineDef := engineConfig.EngineDefs[specialDefID]
			// expectedPayout is relative to the total stake, so multiply by the bet multiplier
			singleRoundPayout := specialEngineDef.ExpectedPayout.Mul(NewFixedFromInt(specialEngineDef.StakeDivisor))
			if cerr != nil {
				logger.Errorf("Error in special win index: %v", specialWin.Index)
				return 0
			}
			for j := 0; j < numRounds; j++ {
				potentialWinValue += singleRoundPayout
				logger.Debugf("Adding single round payout %v for total %v", singleRoundPayout, potentialWinValue)
			}

		}
	}
	logger.Debugf("potential total: %v", potentialWinValue)
	logger.Debugf("length of reel %v", NewFixedFromInt(len(reel)-def.ViewSize[reelIndex]))
	logger.Debugf("division result %v", potentialWinValue.Div(NewFixedFromInt(len(reel)-def.ViewSize[reelIndex])))

	logger.Debugf("betperline %v", gamestate.BetPerLine.Amount)
	logger.Debugf("result %v", potentialWinValue.Div(NewFixedFromInt(len(reel)-def.ViewSize[reelIndex])).Mul(gamestate.BetPerLine.Amount))

	// calculate average for all reel positions, divide total reel potential payout by number of reel positions, multiply by bet per line and stakedivisor
	return potentialWinValue.Div(NewFixedFromInt(len(reel)-def.ViewSize[reelIndex])).Mul(gamestate.BetPerLine.Amount)
}

