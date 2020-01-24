package engine

import (
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

// store and retrieve gamification info

func (data *GamestatePB_Gamification) Increment(stagesPerLevel int32) {
	data.Stage++
	if data.Stage == stagesPerLevel {
		data.Level++
		data.Stage = 0
	}
}
func (data *GamestatePB_Gamification) IncrementSpins(spinsToStageup int32, stagesPerLevel int32) {
	data.RemainingSpins--
	if data.RemainingSpins < 0 {
		if data.SpinsToStageUp > 0 {
			data.SpinsToStageUp = spinsToStageup
			data.RemainingSpins = spinsToStageup
		} else {
			data.SpinsToStageUp = spinsToStageup
			data.RemainingSpins = spinsToStageup
			data.RemainingSpins--
		}
		if data.TotalSpins > 0 {
			data.Stage++
		}
	}
	if data.Stage == stagesPerLevel {
		data.Level++
		data.Stage = 0
	}
	data.TotalSpins++
}

func (data GamestatePB_Gamification) GetLevelAndStage() (int32, int32) {
	return data.Level, data.Stage
}
func (data *GamestatePB_Gamification) GetSpins() int32 {
	// for initialization only
	if data.TotalSpins == 0 {
		initVal := randomRangeInt32()
		data.Level, data.Stage, data.SpinsToStageUp, data.TotalSpins, data.RemainingSpins = 0, 0, initVal, 0, initVal
		logger.Debugf("Initialize Gamification: %+v", data)
	}
	return data.RemainingSpins
}
