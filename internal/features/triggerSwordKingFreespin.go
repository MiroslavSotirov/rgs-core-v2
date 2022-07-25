package features

import (
	"fmt"

	"gitlab.maverick-ops.com/maverick/rgs-core-v2/internal/rng"
	"gitlab.maverick-ops.com/maverick/rgs-core-v2/utils/logger"
)

type TriggerSwordKingFreespin struct {
	FeatureDef
}

func (f *TriggerSwordKingFreespin) DefPtr() *FeatureDef {
	return &f.FeatureDef
}

func (f *TriggerSwordKingFreespin) DataPtr() interface{} {
	return nil
}

func (f *TriggerSwordKingFreespin) Init(def FeatureDef) error {
	return deserializeFeatureDef(f, def)
}

func (f *TriggerSwordKingFreespin) OnInit(state *FeatureState) {
}

func (f TriggerSwordKingFreespin) Trigger(state *FeatureState, params FeatureParams) {
	if f.ForceTrigger(state, params) {
		return
	}

	if state.PureWins ||
		(params.HasKey("RunWilds") && params.GetBool("RunWilds")) ||
		(params.HasKey("RunRespin") && params.GetBool("RunRespin")) ||
		(params.HasKey("RunBonusScatter") && params.GetBool("RunBonusScatter")) ||
		(params.HasKey("RunBonus") && params.GetBool("RunBonus")) {
		logger.Debugf("skipping freespin scatters")
		return
	}

	Probability := params.GetInt("Probability")
	if rng.RandFromRange(10000) < Probability {
		numIdx := WeightedRandomIndex(params.GetIntSlice("NumProbabilities"))
		numScatters := params.GetIntSlice("NumScatters")[numIdx]
		positions := []int{}
		gridh := len(state.SymbolGrid[0])

		logger.Debugf("placing %d freespin scatters", numScatters)

		reels := RandomPermutation([]int{0, 1, 2, 3, 4})

		for s := 0; s < numScatters; s++ {
			reel := reels[s]
			row := rng.RandFromRange(4)
			pos := reel*gridh + row
			positions = append(positions, pos)
		}

		if len(positions) > 0 {
			params["RunFSScatter"] = true
			params["Positions"] = positions
			logger.Debugf("positions: %v", positions)
			activateFeatures(f.FeatureDef, state, params)

			numFreespins := params.GetIntSlice("NumFreespins")[numIdx]
			if numFreespins > 0 {
				fstype := params.GetString("FSType")

				state.Wins = append(state.Wins, FeatureWin{
					Index:           fmt.Sprintf("%s:%d", fstype, numFreespins),
					SymbolPositions: positions,
				})
			}
		}
	}

	return
}

func (f TriggerSwordKingFreespin) ForceTrigger(state *FeatureState, params FeatureParams) bool {
	return false
}

func (f *TriggerSwordKingFreespin) Serialize() ([]byte, error) {
	return serializeTriggerToBytes(f)
}

func (f *TriggerSwordKingFreespin) Deserialize(data []byte) (err error) {
	return deserializeTriggerFromBytes(f, data)
}
