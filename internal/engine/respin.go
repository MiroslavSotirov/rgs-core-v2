package engine

import (
	rgse "gitlab.maverick-ops.com/maverick/rgs-core-v2/errors"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
	"strconv"
	"strings"
)

// Calculate price of reel respin
// value of a reel respin is equal to the expected payout given other reels stay as they are

func (gamestate Gamestate) RespinPrices() (prices []Fixed, err rgse.RGSErr) {
	// get the respin prices
	for reelIndex := 0; reelIndex < len(gamestate.SymbolGrid); reelIndex++ {
		prices = append(prices, RoundUpToNearestCCYUnit(Money{gamestate.RespinPriceReel(reelIndex), gamestate.BetPerLine.Currency}).Amount)
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
	// NB ::: THIS DOES NOT INCLUDE GAMESTATE MULTIPLIERS i.e. RANDOM WHOLE ROUND MULTIPLIERS

	logger.Debugf("Getting expected value of reel %v", reelIndex)
	def, err := gamestate.EngineDef()
	if err != nil {
		return 0
	}
	view := make([][]int, len(gamestate.SymbolGrid))
	copy(view, gamestate.SymbolGrid)
	reel := def.Reels[reelIndex]
	viewSize := def.ViewSize[reelIndex]
	reel = append(reel, reel[:viewSize]...)

	var potentialWinValue Fixed

	// iterate through reel positions for the given reel to determine payouts
	for i := 0; i < len(reel)-viewSize; i++ {

		view[reelIndex] = reel[i : i+viewSize]
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
		}
		specialWin := DetermineSpecialWins(view, def.SpecialPayouts)

		if specialWin.Index != "" {
			// add prize value (multiplier is for total stake, multiply by bet multiplier
			// total prize = totalstake * multiplier = betperline * stakedivisor * multiplier
			potentialWinValue += NewFixedFromInt(specialWin.Payout.Multiplier).Mul(NewFixedFromInt(specialWin.Multiplier)).Mul(NewFixedFromInt(def.StakeDivisor))

			// include estimated value of each round for the number of rounds that have been won
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
			if cerr != nil {
				logger.Errorf("Error in special win index: %v", specialWin.Index)
				return 0
			}

			specialEngineDef := engineConfig.EngineDefs[engineConfig.DefIdByName(winInfo[0])]

			// expectedPayout is relative to the total stake, so multiply by the bet multiplier
			// (N.B. the ExpectedPayout field must be set on the first def for the overall expected payout for all matching defs)

			singleRoundPayout := specialEngineDef.ExpectedPayout.Mul(NewFixedFromInt(specialEngineDef.StakeDivisor))

			for j := 0; j < numRounds; j++ {
				potentialWinValue += singleRoundPayout
			}

		}
	}

	// calculate average for all reel positions
	// divide total reel potential payout by number of reel positions
	// multiply by bet per line
	return potentialWinValue.Div(NewFixedFromInt(len(reel) - viewSize)).Mul(gamestate.BetPerLine.Amount)
}
