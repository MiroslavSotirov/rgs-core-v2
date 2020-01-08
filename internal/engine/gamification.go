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
	logger.Debugf("incrementing for gi data: %#v", data)
	//data.TotalSpins ++
	data.RemainingSpins--
	if data.RemainingSpins <= 0 {
		data.RemainingSpins = spinsToStageup
		data.Stage++
	}
	if data.Stage == stagesPerLevel {
		data.Level++
		data.Stage = 0
	}
}

func (data GamestatePB_Gamification) GetLevelAndStage() (int32, int32) {
	return data.Level, data.Stage
}
func (data GamestatePB_Gamification) GetSpins() int32 {
	return data.RemainingSpins
}
